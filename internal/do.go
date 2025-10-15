package internal

import (
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/wms"
)

var Inj do.Injector

func Init() {
	Inj = do.New()

	config.Init(Inj)
	logging.Init(Inj)
	tilecache.Init(Inj)
	wms.Init(Inj)
}

type tileCache interface {
	Tile(tile model.Tile) (io.Reader, bool)
	Close() error
}

func Stop() {
	tc := do.MustInvokeAs[tileCache]()
	err := tc.Close()
	if err != nil {
		logging.New().WithName("internal").Errorf("error on close tilecache: %v", err)
	}
}
