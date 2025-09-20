package internal

import (
	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
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
