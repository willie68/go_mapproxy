package internal

import (
	"fmt"
	"sync"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/willie68/go_mapproxy/internal/mercantile"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/wms"
)

func TestPreload(t *testing.T) {
	const numWorkers = 20 // Anzahl paralleler Worker
	ast := assert.New(t)
	inj := do.New()

	wm := make(wms.WMSConfigMap)
	wm["gebco"] = wms.Config{
		URL:    "https://geoserver.openseamap.org/geoserver/gwc/service/wms",
		Layers: "gebco2021:gebco_2021",
		Format: "image/png",
	}

	do.ProvideValue(inj, wm)

	wms.Init(inj)
	wms := do.MustInvokeNamed[wms.Service](inj, "gebco")
	ast.NotNil(wms)

	tc := tilecache.Config{
		Active: true,
		Path:   "../testdata/Tilecache",
		MaxAge: 10000,
	}
	do.ProvideValue(inj, &tc)
	tilecache.Init(inj)

	cache := do.MustInvokeAs[*tilecache.Cache](inj)

	type job struct {
		tile    model.Tile
		z, x, y int
	}

	jobs := make(chan job, 100)
	wg := sync.WaitGroup{}

	// Worker starten
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for j := range jobs {
				fmt.Printf("caching for z: %d, x: %d, y: %d\r\n", j.z, j.x, j.y)
				rd, err := wms.WMSTile(tileToBBox(j.tile))
				ast.NoError(err)
				if err == nil {
					defer rd.Close()
					err = cache.Save(j.tile, rd)
					ast.NoError(err)
				}
			}
		})
	}

	// Jobs erzeugen
	for z := range 11 {
		rg := 1 << z
		for x := range rg {
			for y := range rg {
				tile := model.Tile{
					System: "gebco",
					X:      x,
					Y:      y,
					Z:      z,
				}
				jobs <- job{tile: tile, z: z, x: x, y: y}
			}
		}
	}
	close(jobs)
	wg.Wait()
}

// Hilfsfunktion für XYZ->BBOX-Konvertierung
func tileToBBox(t model.Tile) mercantile.Bbox {
	ti := mercantile.TileID{
		X: t.X,
		Y: t.Y,
		Z: t.Z,
	}
	return mercantile.XyBounds(ti)
}
