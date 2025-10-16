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
		log:      logging.New().WithName("wms"),
		configs:  do.MustInvokeAs[tileserverConfig](inj).GetTileserversConfig(),
		services: make([]string, 0),
	}
	do.ProvideValue(inj, &sf)
	for name, config := range sf.configs {
		switch config.Type {
		case "wmss":
			var s Service = &wmsService{
				log:    logging.New().WithName(name),
				config: config,
			}
			do.ProvideNamedValue(inj, name, s)
			sf.services = append(sf.services, name)
		case "tms":
			var s Service = &tmsService{
				log:    logging.New().WithName(name),
				config: config,
				isTMS:  true,
			}
			do.ProvideNamedValue(inj, name, s)
			sf.services = append(sf.services, name)
		case "xyz":
			var s Service = &tmsService{
				log:    logging.New().WithName(name),
				config: config,
				isTMS:  false,
			}
			do.ProvideNamedValue(inj, name, s)
			sf.services = append(sf.services, name)
		default:
			panic(fmt.Sprintf("unknown service type: %s", config.Type))
		}
	}
}

func (f *serviceFactory) HasSystem(name string) bool {
	_, ok := f.configs[name]
	return ok
}
