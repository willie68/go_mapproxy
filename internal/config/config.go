package config

import (
	"fmt"
	"os"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/wms"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Port    int              `yaml:"port"`
	WMSS    wms.WMSConfigMap `yaml:"wmss"`
	Logging logging.Config   `yaml:"logging"`
	Cache   tilecache.Config `yaml:"cache"`
}

var (
	config Config
)

func Logging() *logging.Config {
	return &config.Logging
}

func Cache() *tilecache.Config {
	return &config.Cache
}

func WMSS() *wms.WMSConfigMap {
	return &config.WMSS
}

func SetPort(p int) {
	config.Port = p
}

func Port() int {
	return config.Port
}

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
	return nil
}

func Init(inj do.Injector) {
	do.ProvideValue(inj, &config)

	ver := NewVersion()
	do.ProvideValue(inj, *ver)
}

func (c *Config) JSON() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("can't marshal config to json: %s", err.Error())
	}
	return string(data), nil
}
