package prefetch

import (
	"fmt"
	"io"
	"sync"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/gowillie68/pkg/extstrgutils"
)

var log = logging.New().WithName("prefetch")

type Config struct {
	Workers int `yaml:"workers"` // number of parallel workers
}

type pfConfig interface {
	GetPrefetchConfig() Config
}

type providerFactory interface {
	IsPrefetchable(providerName string) bool
	FTile(tile model.Tile) (io.ReadCloser, error)
}

type tileCache interface {
	Has(tile model.Tile) bool
}

var myinj do.Injector

func Init(inj do.Injector) {
	myinj = inj
}

// Prefetch lädt Kacheln für die angegebenen Systeme und Zoomstufen vor.
func Prefetch(providers string, maxzoom int) error {
	cfg := do.MustInvokeAs[pfConfig](myinj).GetPrefetchConfig()
	workers := 10
	if cfg.Workers > 0 {
		workers = cfg.Workers
	}
	syss := extstrgutils.SplitMultiValueParam(providers)
	fmt.Printf("syss: %v", syss)
	jobs := make(chan model.Tile, 1000)
	wg := sync.WaitGroup{}

	ts := do.MustInvokeAs[providerFactory](myinj)
	cache := do.MustInvokeAs[tileCache](myinj)

	// Worker starten
	for range workers {
		wg.Go(func() {
			for j := range jobs {
				if ts.IsPrefetchable(j.Provider) {
					rd, err := ts.FTile(j)
					if err != nil {
						log.Errorf("error getting tile: %v", err)
						continue
					}
					defer rd.Close()
					log.Infof("fetched tile: %v", j)
				}
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
						Provider: sys,
						X:        x,
						Y:        y,
						Z:        z,
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
