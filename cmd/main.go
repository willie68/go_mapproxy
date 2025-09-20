package main

import (
	"fmt"
	"net/http"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/willie68/go_mapproxy/internal/api"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/tilecache"
)

var (
	log         *logging.Logger
	configFile  string
	showVersion bool
)

func init() {
	flag.BoolVarP(&showVersion, "version", "v", false, "showing the version")
	flag.StringVarP(&configFile, "config", "c", "config.yaml", "this is the path and filename to the config file")
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Println(config.NewVersion().String())
		fmt.Println("more on https://github.com/willie68/go_mapproxy")
		os.Exit(0)
	}
	err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	js, err := config.Get().ToJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Config:\n%s\n", js)

	logging.Init()
	log = logging.New().WithName("main")
	log.Info("starting tms service")

	tilecache.New()

	http.HandleFunc("/", api.NewTMSHandler().Handler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Get().Port), nil)
	if err != nil {
		log.Fatalf("error on listen and serv: %v", err)
	}
	log.Info("server finished")
}
