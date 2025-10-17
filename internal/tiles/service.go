package tiles

import (
	"bytes"
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/tileserver"
)

type tileserverServiceFactory interface {
	HasSystem(name string) bool
}

type tileCache interface {
	Tile(tile model.Tile) (io.ReadCloser, bool)
	Save(tile model.Tile, data io.Reader) error
	IsActive() bool
}

type service struct {
	inj   do.Injector
	log   *logging.Logger
	cache tileCache
	wms   tileserverServiceFactory
}

func Init(inj do.Injector) {
	do.ProvideValue(inj, &service{
		inj:   inj,
		log:   logging.New().WithName("tiles"),
		cache: do.MustInvokeAs[tileCache](inj),
		wms:   do.MustInvokeAs[tileserverServiceFactory](inj),
	})
}

func (s *service) FTile(tile model.Tile) (io.ReadCloser, error) {
	// try to get the cached tile
	if tr, ok := s.cache.Tile(tile); ok {
		s.log.Debugf("tile found in cache: %s", tile.String())
		return tr, nil
	}

	if !s.HasSystem(tile.System) {
		return nil, tileserver.ErrNotFound
	}
	ts, err := do.InvokeNamed[tileserver.Service](s.inj, tile.System)
	if err != nil {
		s.log.Errorf("System error: %v", err)
		return nil, err
	}

	// if cache is inactive simply, get the tile from the tileserver
	if !s.cache.IsActive() {
		return ts.Tile(tile)
	}

	rd, err := ts.Tile(tile)
	if err != nil {
		s.log.Errorf("error getting tile from tileserver: %v", err)
		return nil, err
	}

	data, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	go func() {
		rd = io.NopCloser(bytes.NewReader(data))
		err = s.cache.Save(tile, rd)
		if err != nil {
			s.log.Errorf("error saving tile to cache: %v", err)
		}
	}()
	rd = io.NopCloser(bytes.NewReader(data))
	return rd, nil
}

func (s *service) HasSystem(name string) bool {
	return s.wms.HasSystem(name)
}
