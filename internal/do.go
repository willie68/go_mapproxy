package internal

import (
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/prefetch"
	"github.com/willie68/go_mapproxy/internal/provider"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/tiles"
	"github.com/willie68/go_mapproxy/internal/utils/measurement"
)

var Inj do.Injector

func Init() {
	Inj = do.New()

	config.Init(Inj)
	logging.Init(Inj)

	metrics := measurement.New(true)
	do.ProvideValue(Inj, metrics)

	prefetch.Init(Inj)

	tilecache.Init(Inj)
	provider.Init(Inj)
	tiles.Init(Inj)
}

type tileCache interface {
	Tile(tile model.Tile) (io.Reader, bool)
	Close() error
}

func Stop() {
	tc := do.MustInvokeAs[tileCache](Inj)
	err := tc.Close()
	if err != nil {
		logging.New().WithName("internal").Errorf("error on close tilecache: %v", err)
	}
}
