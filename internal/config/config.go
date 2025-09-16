package config

import (
	"fmt"
	"os"

	"github.com/willie68/go_mapproxy/internal/logging"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Port    int            `yaml:"port"`
	WMSS    map[string]WMS `yaml:"wmss"`
	Logging logging.Config `yaml:"logging"`
	Cache   string         `yaml:"cache"`
}

type WMS struct {
	URL    string `yaml:"url"`
	Layers string `yaml:"layers"`
}

var (
	config Config
)

func Get() Config {
	return config
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
