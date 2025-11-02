package internal

import (
	"fmt"
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/prefetch"
	"github.com/willie68/go_mapproxy/internal/provider"
	"github.com/willie68/go_mapproxy/internal/shttp"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/tiles"
	"github.com/willie68/go_mapproxy/internal/utils/measurement"
)

func Init(inj do.Injector) {
	config.Init(inj)
	logging.Init(inj)

	metrics := measurement.New(true)
	do.ProvideValue(inj, metrics)

	prefetch.Init(inj)

	provider.Init(inj)
	tilecache.Init(inj)
	tiles.Init(inj)

	shttp.NewSHttp(inj)
}

type tileCache interface {
	Tile(tile model.Tile) (io.Reader, bool)
	Close() error
}

func Stop(inj do.Injector) {
	tc := do.MustInvokeAs[tileCache](inj)
	err := tc.Close()
	if err != nil {
		logging.New("internal").Error(fmt.Sprintf("error on close tilecache: %v", err))
	}
}
