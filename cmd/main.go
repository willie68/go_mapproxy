package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"
	"github.com/willie68/go_mapproxy/configs"
	"github.com/willie68/go_mapproxy/internal"
	"github.com/willie68/go_mapproxy/internal/api"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/prefetch"
	"github.com/willie68/gowillie68/pkg/fileutils"
)

var (
	log         *logging.Logger
	configFile  string
	showVersion bool
	initConfig  bool
	pfZoom      int
	pfSystem    string
	port        int
)

func init() {
	flag.BoolVarP(&initConfig, "init", "i", false, "init config, writes out a default config.")
	flag.BoolVarP(&showVersion, "version", "v", false, "showing the version")
	flag.StringVarP(&configFile, "config", "c", "config.yaml", "this is the path and filename to the config file")
	flag.IntVarP(&port, "port", "p", 0, "overwrite the port (8580) of the config")
	flag.IntVarP(&pfZoom, "zoom", "z", 0, "max zoom for prefetch tiles")
	flag.StringVarP(&pfSystem, "system", "s", "", "prefetch system, if empty no prefetching will be done, csv if more than one needed.")
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println("more on https://github.com/willie68/go_mapproxy")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("examples:")
		fmt.Println("simply run as proxy: take the default config, add your needed provider and run")
		fmt.Printf("%s -c config.yaml\n", os.Args[0])
		fmt.Println("run as proxy with caching: take the default config, add your needed provider,switch caching to true and set a path. Than run")
		fmt.Printf("%s -c config.yaml\n", os.Args[0])
		fmt.Println("run as proxy with caching and prefetching zomm 5: take the default config, add your needed provider,switch caching to true and set a path. Than run")
		fmt.Printf("%s -c config.yaml -s <your system to be cached> -z 4\n", os.Args[0])
	}
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Println(config.NewVersion().String())
		os.Exit(0)
	}
	if initConfig {
		fmt.Println(configs.ConfigFile)
		os.Exit(0)
	}
	if !fileutils.FileExists(configFile) {
		fmt.Fprint(os.Stderr, "no config given or dosn't exists.\r\n\r\n")
		showUsage()
		os.Exit(1)
	}
	err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	config.SetParameter(config.WithPort(port))
	js := config.JSON()
	if js == "" {
		panic("error on marshal config to json")
	}
	fmt.Printf("Config:\n%s\n", js)
	log = logging.New().WithName("main")
	log.Info("starting tms service")

	internal.Init()

	if pfSystem != "" && pfZoom > 0 {
		go func() {
			log.Infof("starting prefetch for system %s with zoom %d", pfSystem, pfZoom)
			prefetch.Prefetch(pfSystem, pfZoom)
			log.Info("prefetch finnished")
		}()
	}
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info("shutting down server...")
		internal.Stop()
		os.Exit(0)
	}()

	http.HandleFunc("/", api.NewTMSHandler(internal.Inj).Handler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port()), nil)
	if err != nil {
		log.Fatalf("error on listen and serv: %v", err)
	}
	log.Info("server finished")
	internal.Stop()
}

func showUsage() {
	flag.Usage()
}
