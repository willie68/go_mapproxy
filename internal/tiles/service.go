package tiles

import (
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/tileserver"
)

type tileserverServiceFactory interface {
	HasSystem(name string) bool
}

type service struct {
	inj do.Injector
	log *logging.Logger
	wms tileserverServiceFactory
}

func Init(inj do.Injector) {
	do.ProvideValue(inj, &service{
		inj: inj,
		log: logging.New().WithName("tiles"),
		wms: do.MustInvokeAs[tileserverServiceFactory](inj),
	})
}

func (s *service) Tile(tile model.Tile) (io.ReadCloser, error) {
	if !s.HasSystem(tile.System) {
		return nil, tileserver.ErrNotFound
	}
	ts, err := do.InvokeNamed[tileserver.Service](s.inj, tile.System)
	if err != nil {
		s.log.Errorf("System error: %v", err)
		return nil, err
	}
	return ts.Tile(tile)
}

func (s *service) HasSystem(name string) bool {
	return s.wms.HasSystem(name)
}
