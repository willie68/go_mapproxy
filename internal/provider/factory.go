package provider

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/configs"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type Service interface {
	Tile(tile model.Tile) (io.ReadCloser, error)
}

type ConfigMap map[string]Config

type Config struct {
	URL        string            `yaml:"url"`
	Type       string            `yaml:"type"` // wmss, tms, xyz
	NoCached   bool              `yaml:"nocache"`
	Layers     string            `yaml:"layers"`
	Format     string            `yaml:"format"`
	Styles     string            `yaml:"styles"`
	Version    string            `yaml:"version"`
	Headers    map[string]string `yaml:"headers"`
	Path       string            `yaml:"path"` // for file based providers
	Fallback   string            `yaml:"fallback"`
	NoPrefetch bool              `yaml:"noprefetch"` // disable any prefetching of tiles
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
		case "wms":
			var s Service = &wmsProvider{
				name:   sname,
				log:    logging.New().WithName(fmt.Sprintf("wms: %s", sname)),
				config: config,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "tms":
			var s Service = &tmsProvider{
				name:   sname,
				log:    logging.New().WithName(fmt.Sprintf("tms: %s", sname)),
				config: config,
				isTMS:  true,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "xyz":
			var s Service = &tmsProvider{
				name:   sname,
				log:    logging.New().WithName(fmt.Sprintf("xyz: %s", sname)),
				config: config,
				isTMS:  false,
			}
			do.ProvideNamedValue(inj, sname, s)
			sf.services = append(sf.services, sname)
		case "mbtiles":
			var s Service = NewMBTilesProvider(sname, config, inj)
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

func (f *pFactory) IsPrefetchable(providerName string) bool {
	config, ok := f.configs[providerName]
	if !ok {
		return false
	}
	bl := configs.PrefetchBlacklist()
	for _, b := range bl {
		if strings.Contains(strings.ToLower(config.URL), strings.ToLower(b)) {
			return false
		}
	}
	return !config.NoPrefetch
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "go_mapproxy/0.1")
	req.Header.Set("Accept", "*/*")
}
