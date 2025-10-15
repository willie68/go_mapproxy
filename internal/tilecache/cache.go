package tilecache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type TileCache interface {
	Has(tile model.Tile) bool
	Tile(tile model.Tile) (io.Reader, bool)
	Save(tile model.Tile, data io.Reader) error
	IsActive() bool
}

type Config struct {
	Path   string `yaml:"path"`
	Active bool   `yaml:"active"`
	MaxAge int    `yaml:"maxage"` // in hours
}

type Cache struct {
	log    *logging.Logger
	path   string
	active bool
	maxage int // in hours

	flock sync.RWMutex
}

func Init(inj do.Injector) {
	cfg := do.MustInvoke[*Config](inj)
	c := &Cache{
		log:    logging.New().WithName("tilecache"),
		path:   cfg.Path,
		active: cfg.Active,
		maxage: cfg.MaxAge,
		flock:  sync.RWMutex{},
	}
	if c.active {
		c.startCacheCleanupJob()
	}
	do.ProvideValue(inj, c)
}

func (c *Cache) startCacheCleanupJob() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := c.CleanupOldFiles(time.Duration(c.maxage) * time.Hour)
			if err != nil {
				c.log.Errorf("cache cleanup error: %v", err)
			} else {
				c.log.Infof("cache cleanup completed")
			}
		}
	}()
}

func (c *Cache) IsActive() bool {
	return c.active
}

func (c *Cache) Has(tile model.Tile) bool {
	if !c.active {
		return false
	}
	fname := c.getFilename(tile)
	c.flock.RLock()
	defer c.flock.RUnlock()
	if _, err := os.Stat(fname); err != nil {
		return false
	}
	return true
}

func (c *Cache) Tile(tile model.Tile) (io.Reader, bool) {
	if !c.active {
		return nil, false
	}
	fname := c.getFilename(tile)
	c.flock.RLock()
	defer c.flock.RUnlock()
	if _, err := os.Stat(fname); err != nil {
		return nil, false
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, false
	}
	return f, true
}

func (c *Cache) Save(tile model.Tile, data io.Reader) error {
	if !c.active {
		return nil
	}
	fn := c.getFilename(tile)
	// only cache if the file does not exists
	c.flock.RLock()
	if _, err := os.Stat(fn); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(fn), 0o755); err != nil {
			return err
		}
		c.flock.RUnlock()
		c.flock.Lock()
		defer c.flock.Unlock()
		f, err := os.Create(fn)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, data)
		return err
	} else {
		c.flock.RUnlock()
	}
	return nil
}

// CleanupOldFiles deletes cache files older than the given duration.
func (c *Cache) CleanupOldFiles(olderThan time.Duration) error {
	root := c.path // adjust if your cache path is named differently
	now := time.Now()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		modTime := info.ModTime()
		if now.Sub(modTime) > olderThan {
			c.log.Debugf("removing old cache file: %s", path)
			c.deleteFile(path)
		}
		return nil
	})
}

func (c *Cache) deleteFile(path string) {
	c.flock.Lock()
	defer c.flock.Unlock()
	err := os.Remove(path)
	if err != nil {
		c.log.Errorf("error removing file %s: %v", path, err)
	}
}

func (c *Cache) getFilename(tile model.Tile) string {
	return filepath.Join(c.path, tile.System, strconv.Itoa(tile.Z), strconv.Itoa(tile.X), fmt.Sprintf("%d.png", tile.Y))
}

func (c *Cache) GetFileHash(fileStr string) string {
	f, err := os.Open(fileStr)
	if err != nil {
		c.log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		c.log.Fatalf("error building hash: %v", err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Cache) Close() error {
	return nil
}
