package provider

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
	URL      string            `yaml:"url"`
	Type     string            `yaml:"type"` // wmss, tms, xyz
	NoCached bool              `yaml:"nocache"`
	Layers   string            `yaml:"layers"`
	Format   string            `yaml:"format"`
	Styles   string            `yaml:"styles"`
	Version  string            `yaml:"version"`
	Headers  map[string]string `yaml:"headers"`
	Path     string            `yaml:"path"` // for file based providers
	Fallback string            `yaml:"fallback"`
}

type pFactory struct {
	log      *logging.Logger
	configs  ConfigMap
	services []string
	inj      do.Injector
}

var (
	ErrNotFound = errors.New("service not found")
)

type providerConfig interface {
	GetProviderConfig() ConfigMap
}

func Init(inj do.Injector) {
	sf := pFactory{
		log:      logging.New().WithName("factory"),
		configs:  do.MustInvokeAs[providerConfig](inj).GetProviderConfig(),
		services: make([]string, 0),
		inj:      inj,
	}
	do.ProvideValue(inj, &sf)
	for sname, config := range sf.configs {
		switch config.Type {
		case "wmss":
			var s Service = &wmsProvider{
				log:    logging.New().WithName(sname),
				config: config,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "tms":
			var s Service = &tmsProvider{
				log:    logging.New().WithName(sname),
				config: config,
				isTMS:  true,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "xyz":
			var s Service = &tmsProvider{
				log:    logging.New().WithName(sname),
				config: config,
				isTMS:  false,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "mbtiles":
			var s Service = NewMBTilesProvider(config, inj)
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		default:
			panic(fmt.Sprintf("unknown service type: %s", config.Type))
		}
	}
}

func (f *pFactory) HasProvider(providerName string) bool {
	_, ok := f.configs[providerName]
	return ok
}

func (f *pFactory) IsCached(providerName string) bool {
	config, ok := f.configs[providerName]
	if !ok {
		return false
	}
	return !config.NoCached
}
