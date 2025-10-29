package config

import (
	"fmt"
	"os"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/prefetch"
	"github.com/willie68/go_mapproxy/internal/provider"
	"github.com/willie68/go_mapproxy/internal/shttp"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"go.yaml.in/yaml/v3"
)

type service struct {
	HTTP      shttp.Config       `yaml:"http"`
	Providers provider.ConfigMap `yaml:"provider"`
	Logging   logging.Config     `yaml:"logging"`
	Cache     tilecache.Config   `yaml:"cache"`
	Prefetch  prefetch.Config    `yaml:"prefetch"`
}

type ParameterOption func(*service)

func WithPort(port int) ParameterOption {
	return func(s *service) {
		s.SetPort(port)
	}
}

var (
	config service
)

func (c service) GetLoggingConfig() logging.Config {
	return c.Logging
}

func (c service) GetCacheConfig() tilecache.Config {
	return c.Cache
}

func (c service) GetProviderConfig() provider.ConfigMap {
	return c.Providers
}

func (c service) GetPrefetchConfig() prefetch.Config {
	return c.Prefetch
}

func (c service) GetHttpConfig() shttp.Config {
	return c.HTTP
}

func (c *service) SetPort(p int) {
	if p > 0 {
		c.HTTP.Port = p
	}
}

func (c *service) GetPort() int {
	return c.HTTP.Port
}

func SetParameter(params ...ParameterOption) {
	for _, p := range params {
		p(&config)
	}
}

func Port() int {
	return config.HTTP.Port
}

// JSON returns the config as json string
func JSON() string {
	js, err := config.JSON()
	if err != nil {
		return ""
	}
	return js
}

// Load loads the config
func Load(file string) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("can't load config file: %s", err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("can't unmarshal config file: %s", err.Error())
	}
	if config.HTTP.Port <= 0 {
		config.HTTP.Port = 8580
	}
	return nil
}

func Init(inj do.Injector) {
	do.ProvideValue(inj, config)

	ver := NewVersion()
	do.ProvideValue(inj, ver)
}

func (c *service) JSON() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("can't marshal config to json: %s", err.Error())
	}
	return string(data), nil
}
