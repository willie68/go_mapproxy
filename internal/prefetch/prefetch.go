package prefetch

import (
	"fmt"
	"sync"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/tilecache"
	"github.com/willie68/go_mapproxy/internal/wms"
)

var log = logging.New().WithName("prefetch")

func Prefetch(system string, maxzoom int) error {
	const numWorkers = 16 // Anzahl paralleler Worker

	wms := do.MustInvokeNamed[wms.Service](internal.Inj, system)
	cache := do.MustInvokeAs[*tilecache.Cache](internal.Inj)

	jobs := make(chan model.Tile, 1000)
	wg := sync.WaitGroup{}

	// Worker starten
	for range numWorkers {
		wg.Go(func() {
			for j := range jobs {
				fmt.Printf("caching for z: %d, x: %d, y: %d\r\n", j.Z, j.X, j.Y)
				rd, err := wms.WMSTile(tileToBBox(j))
				if err != nil {
					log.Errorf("error getting tile: %v", err)
					continue
				}
				defer rd.Close()
				err = cache.Save(j, rd)
				if err != nil {
					log.Errorf("error caching tile: %v", err)
				}
			}
		})
	}

	// Jobs erzeugen
	for z := range maxzoom + 1 {
		rg := 1 << z
		for x := range rg {
			for y := range rg {
				tile := model.Tile{
					System: system,
					X:      x,
					Y:      y,
					Z:      z,
				}
				if !cache.Has(tile) {
					jobs <- tile
				}
			}
		}
	}
	close(jobs)
	wg.Wait()
	return nil
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
