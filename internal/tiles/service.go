package tiles

import (
	"bytes"
	"fmt"
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/provider"
	"github.com/willie68/go_mapproxy/internal/utils/measurement"
)

type providerFactory interface {
	HasProvider(providerName string) bool
	IsCached(providerName string) bool
}

type tileCache interface {
	Tile(tile model.Tile) (io.ReadCloser, bool)
	Save(tile model.Tile, data io.Reader) error
	IsActive() bool
}

type service struct {
	inj     do.Injector
	log     *logging.Logger
	cache   tileCache
	tssf    providerFactory
	metrics *measurement.Service
}

func Init(inj do.Injector) {
	do.ProvideValue(inj, &service{
		inj:     inj,
		log:     logging.New().WithName("tiles"),
		cache:   do.MustInvokeAs[tileCache](inj),
		tssf:    do.MustInvokeAs[providerFactory](inj),
		metrics: do.MustInvoke[*measurement.Service](inj),
	})
}

func (s *service) FTile(tile model.Tile) (io.ReadCloser, error) {
	// try to get the cached tile

	if !s.HasProvider(tile.Provider) {
		return nil, provider.ErrNotFound
	}

	if s.IsCached(tile.Provider) {
		td := s.metrics.Start("getTileFromCache")
		if tr, ok := s.cache.Tile(tile); ok {
			td.Stop()
			s.log.Debugf("tile found in cache: %s", tile.String())
			return tr, nil
		}
		td.Stop()
	}

	ts, err := do.InvokeNamed[provider.Service](s.inj, tile.Provider)
	if err != nil {
		s.log.Errorf("System error: %v", err)
		return nil, err
	}

	td := s.metrics.Start("getTileFromProvider")
	tsd := s.metrics.Start(fmt.Sprintf("getTileFromProvider:%s", tile.Provider))
	rd, err := ts.Tile(tile)
	if err != nil {
		s.log.Errorf("error getting tile from tileserver: %v", err)
		return nil, err
	}
	tsd.Stop()
	td.Stop()

	if s.IsCached(tile.Provider) {
		// if cache is inactive simply, get the tile from the tileserver
		if !s.cache.IsActive() {
			return rd, nil
		}
	}

	data, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if s.IsCached(tile.Provider) {
		go func() {
			rd = io.NopCloser(bytes.NewReader(data))
			td := s.metrics.Start("saveTileToCache")
			defer td.Stop()
			err = s.cache.Save(tile, rd)
			if err != nil {
				s.log.Errorf("error saving tile to cache: %v", err)
			}
		}()
	}
	rd = io.NopCloser(bytes.NewReader(data))
	return rd, nil
}

func (s *service) HasProvider(providerName string) bool {
	return s.tssf.HasProvider(providerName)
}

func (s *service) IsCached(providerName string) bool {
	return s.tssf.IsCached(providerName)
}
