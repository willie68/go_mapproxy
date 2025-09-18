package wms

type WMSConfigMap map[string]Config

type Config struct {
	URL    string `yaml:"url"`
	Layers string `yaml:"layers"`
	Format string `yaml:"format"`
}
