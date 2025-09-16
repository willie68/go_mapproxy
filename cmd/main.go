package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/willie68/go_mapproxy/internal/config"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
)

var (
	log        *logging.Logger
	configFile string
)

func init() {
	flag.StringVarP(&configFile, "config", "c", "config.yaml", "this is the path and filename to the config file")
}

// Hilfsfunktion für XYZ->BBOX-Konvertierung
func tileToBBox(x, y, z int) mercantile.Bbox {
	t := mercantile.TileID{
		X: x,
		Y: y,
		Z: z,
	}
	return mercantile.XyBounds(t)
}

func tmsHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /tms/{z}/{x}/{y}.png
	path := r.URL.Path
	log.Infof("path: %s", path)
	p := strings.Split(path, "/")
	if len(p) != 6 {
		http.Error(w, "Path error", http.StatusBadRequest)
		return
	}
	sys := p[1]
	if _, ok := config.Get().WMSS[sys]; !ok {
		http.Error(w, "unknown system", http.StatusBadRequest)
		return
	}
	z, err := strconv.Atoi(p[3])
	if err != nil {
		http.Error(w, "error in zoom", http.StatusBadRequest)
		return
	}
	x, err := strconv.Atoi(p[4])
	if err != nil {
		http.Error(w, "error in x axis", http.StatusBadRequest)
		return
	}
	fn := filepath.Base(p[5])
	ys := strings.TrimSuffix(fn, filepath.Ext(fn))
	y, err := strconv.Atoi(ys)
	if err != nil {
		http.Error(w, "error in y axis", http.StatusBadRequest)
		return
	}

	cachePath := ""
	if config.Get().Cache != "" {
		// try to get the cached tile
		cachePath = fmt.Sprintf("%s/%s/%d/%d/%d.png", config.Get().Cache, sys, z, x, y)
		if _, err := os.Stat(cachePath); err == nil {
			// Tile aus Cache
			http.ServeFile(w, r, cachePath)
			return
		}
	}

	// BBOX berechnen
	bb := tileToBBox(x, y, z)

	// WMS-Request bauen
	// https://geoserver.openseamap.org/geoserver/gwc/service/wms?Request=GetMap&layers=gebco2021:gebco_2021&format=image/png&bbox=-180,-90,0,90&width=256&height=256&srs=EPSG:4326
	// minLon, minLat, maxLon, maxLat)
	tmplURL := "%s?request=GetMap&layers=%s&format=image/png&bbox=%.9f,%.9f,%.9f,%.9f&width=256&height=256&srs=EPSG:3857"
	wms := config.Get().WMSS["gebco"]
	wmsURL := fmt.Sprintf(tmplURL, wms.URL, wms.Layers, bb.Left, bb.Bottom, bb.Right, bb.Top)

	log.Debugf("wms url: %s", wmsURL)
	resp, err := http.Get(wmsURL)
	if err != nil || resp.StatusCode != 200 {
		if resp.Body != nil {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Read error", http.StatusInternalServerError)
				return
			}
			bodyString := string(bodyBytes)
			log.Errorf("body: %s", bodyString)
		}
		http.Error(w, "Tile error", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()
	// if cache is active put the tile to the cache
	if cachePath != "" {
		os.MkdirAll(filepath.Dir(cachePath), 0755)
		f, _ := os.Create(cachePath)
		io.Copy(f, resp.Body)
		f.Close()

		// Und ausliefern
		http.ServeFile(w, r, cachePath)
		return
	}
	io.Copy(w, resp.Body)
}

func main() {
	flag.Parse()
	err := config.Load(configFile)
	if err != nil {
		panic(err)
	}
	logging.Init(config.Get().Logging)
	log = logging.New().WithName("main")
	log.Info("starting tms service")

	http.HandleFunc("/", tmsHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Get().Port), nil)
	if err != nil {
		log.Fatalf("error on listen and serv: %v", err)
	}
	log.Info("server finished")
}
