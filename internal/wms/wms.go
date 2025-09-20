package wms

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
)

type WMSConfigMap map[string]Config

type Config struct {
	URL    string `yaml:"url"`
	Layers string `yaml:"layers"`
	Format string `yaml:"format"`
}

type ServiceFactory struct {
	log      *logging.Logger
	configs  WMSConfigMap
	services map[string]Service
}

type Service struct {
	log    *logging.Logger
	config Config
}

var (
	ErrNotFound = errors.New("service not found")
)

func Init(inj do.Injector) {
	sf := ServiceFactory{
		log:      logging.New().WithName("wms"),
		configs:  do.MustInvoke[WMSConfigMap](inj),
		services: make(map[string]Service),
	}
	do.ProvideValue(inj, &sf)
	for name, config := range sf.configs {
		s := Service{
			log:    logging.New().WithName(name),
			config: config,
		}
		sf.services[name] = s
		do.ProvideNamedValue(inj, name, s)
	}
}

// WMS getting a wms service. ErrNotFound if not in the config. Will check if the service is already in the service map.
// This will work without DI
func (f *ServiceFactory) WMS(name string) (*Service, error) {
	s, ok := f.services[name]
	if ok {
		return &s, nil
	}
	if _, ok := f.configs[name]; !ok {
		return nil, ErrNotFound
	}
	s = Service{
		log:    logging.New().WithName(name),
		config: f.configs[name],
	}
	f.services[name] = s
	return &s, nil
}

func (s *Service) WMSTile(bbox mercantile.Bbox) (io.ReadCloser, error) {
	wmsURL := s.buildWMSUrl(bbox)

	resp, err := http.Get(wmsURL)
	if err != nil || resp.StatusCode != 200 {
		if resp.Body != nil {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.New("read error")
			}
			bodyString := string(bodyBytes)
			s.log.Errorf("body: %s", bodyString)
		}
		s.log.Errorf("error on wms request, status: %s: %v", resp.Status, err)
		return nil, errors.New("Tile error")
	}
	return resp.Body, nil
}

func (s *Service) buildWMSUrl(bb mercantile.Bbox) string {

	base, err := url.Parse(s.config.URL)
	if err != nil {
		panic(err)
	}

	params := url.Values{}
	params.Add("request", "GetMap")
	params.Add("layers", s.config.Layers)
	params.Add("format", s.config.Format)
	params.Add("bbox", fmt.Sprintf("%.9f,%.9f,%.9f,%.9f", bb.Left, bb.Bottom, bb.Right, bb.Top))
	params.Add("width", "256")
	params.Add("height", "256")
	params.Add("srs", "EPSG:3857")

	base.RawQuery = params.Encode()
	wmsURL := base.String()

	s.log.Debugf("wms url: %s", wmsURL)
	return wmsURL
}
