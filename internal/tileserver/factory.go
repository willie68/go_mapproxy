package tileserver

import (
	"errors"
	"fmt"
	"io"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type Service interface {
	Tile(tile model.Tile) (io.ReadCloser, error)
}

type ConfigMap map[string]Config

type Config struct {
	URL     string            `yaml:"url"`
	Type    string            `yaml:"type"` // wmss, tms, xyz
	Cached  bool              `yaml:"cached"`
	Layers  string            `yaml:"layers"`
	Format  string            `yaml:"format"`
	Styles  string            `yaml:"styles"`
	Version string            `yaml:"version"`
	Headers map[string]string `yaml:"headers"`
}

type serviceFactory struct {
	log      *logging.Logger
	configs  ConfigMap
	services []string
}

var (
	ErrNotFound = errors.New("service not found")
)

type tileserverConfig interface {
	GetTileserversConfig() ConfigMap
}

func Init(inj do.Injector) {
	sf := serviceFactory{
		log:      logging.New().WithName("factory"),
		configs:  do.MustInvokeAs[tileserverConfig](inj).GetTileserversConfig(),
		services: make([]string, 0),
	}
	do.ProvideValue(inj, &sf)
	for sname, config := range sf.configs {
		switch config.Type {
		case "wmss":
			var s Service = &wmsService{
				log:    logging.New().WithName(sname),
				config: config,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "tms":
			var s Service = &tmsService{
				log:    logging.New().WithName(sname),
				config: config,
				isTMS:  true,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "xyz":
			var s Service = &tmsService{
				log:    logging.New().WithName(sname),
				config: config,
				isTMS:  false,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		default:
			panic(fmt.Sprintf("unknown service type: %s", config.Type))
		}
	}
}

func (f *serviceFactory) HasSystem(systemname string) bool {
	_, ok := f.configs[systemname]
	return ok
}

func (f *serviceFactory) IsCached(systemname string) bool {
	config, ok := f.configs[systemname]
	if !ok {
		return false
	}
	return config.Cached
}
