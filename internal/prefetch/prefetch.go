package prefetch

import (
	"fmt"
	"io"
	"sync"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/gowillie68/pkg/extstrgutils"
)

var log = logging.New().WithName("prefetch")

type tileService interface {
	Tile(tile model.Tile) (io.ReadCloser, error)
}

type tileCache interface {
	Has(tile model.Tile) bool
}

// Prefetch lädt Kacheln für die angegebenen Systeme und Zoomstufen vor.
func Prefetch(systems string, maxzoom int) error {
	const numWorkers = 16 // Anzahl paralleler Worker
	syss := extstrgutils.SplitMultiValueParam(systems)
	fmt.Printf("syss: %v", syss)
	jobs := make(chan model.Tile, 1000)
	wg := sync.WaitGroup{}

	ts := do.MustInvokeAs[tileService](internal.Inj)
	cache := do.MustInvokeAs[tileCache](internal.Inj)

	// Worker starten
	for range numWorkers {
		wg.Go(func() {
			for j := range jobs {
				rd, err := ts.Tile(j)
				if err != nil {
					log.Errorf("error getting tile: %v", err)
					continue
				}
				defer rd.Close()
				log.Infof("fetched tile: %v", j)
			}
		})
	}

	for _, sys := range syss {
		// Jobs erzeugen
		for z := range maxzoom + 1 {
			rg := 1 << z
			for x := range rg {
				for y := range rg {
					tile := model.Tile{
						System: sys,
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
	}
	close(jobs)
	wg.Wait()
	return nil
}
