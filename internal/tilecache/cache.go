package tilecache

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

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
	db    *badger.DB
}

type tcConfig interface {
	GetCacheConfig() Config
}

type dbEntry struct {
	Hash      string
	Timestamp time.Time
}

func Init(inj do.Injector) {
	cfg := do.MustInvokeAs[tcConfig](inj).GetCacheConfig()
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
	if c.active {
		db, err := badger.Open(badger.DefaultOptions(c.getDBPath()).WithValueLogFileSize(100 * 1024 * 1024))
		if err != nil {
			c.log.Errorf("failed to open badger db: %v", err)
		}
		c.db = db
		c.startValueLogGCTicker()
	}
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

func (c *Cache) startValueLogGCTicker() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := c.db.RunValueLogGC(0.5)
			if err != nil {
				c.log.Errorf("value log GC error: %v", err)
			} else {
				c.log.Infof("value log GC completed")
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
	// Check if DB has entry
	return c.DBHas(tile)
}

func (c *Cache) Tile(tile model.Tile) (io.ReadCloser, bool) {
	if !c.active {
		return nil, false
	}
	if !c.DBHas(tile) {
		return nil, false
	}
	db, err := c.DBGet(tile)
	if err != nil || db == nil {
		return nil, false
	}
	_, file := c.getFilename(db.Hash)
	c.flock.RLock()
	defer c.flock.RUnlock()
	if _, err := os.Stat(file); err != nil {
		// File not found, remove DB entry
		c.log.Errorf("cache file %s not found", file)
		return nil, false
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, false
	}
	return f, true
}

func (c *Cache) Save(tile model.Tile, data io.Reader) error {
	if !c.active {
		return nil
	}
	if c.DBHas(tile) {
		db, err := c.DBGet(tile)
		if err != nil {
			return err
		}
		if db != nil {
			_, file := c.getFilename(db.Hash)
			if _, err := os.Stat(file); err == nil {
				// File with same hash already exists
				return nil
			}
		}
	}
	// Create temporary file to calculate hash
	tmpFile, err := os.CreateTemp("", "tile_cache_*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy data to temp file and calculate hash
	h := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, h)
	_, err = io.Copy(multiWriter, data)
	tmpFile.Close()
	if err != nil {
		return err
	}

	// Generate hash-based path
	hash := hex.EncodeToString(h.Sum(nil))
	hashDir, hashFile := c.getFilename(hash)

	// Check if hash-based file already exists
	if _, err := os.Stat(hashFile); err == nil {
		// File already exists, no need to save again
		if !c.DBHas(tile) {
			err = c.DBSet(tile, dbEntry{Hash: hash, Timestamp: time.Now()})
			if err != nil {
				return err
			}
		}
		return nil
	}
	// Create hash-based directory structure
	if err := os.MkdirAll(hashDir, 0o755); err != nil {
		return err
	}

	// Move temp file to final hash-based location
	c.flock.Lock()
	defer c.flock.Unlock()
	err = os.Rename(tmpPath, hashFile)
	if err != nil {
		// Fallback: copy if rename fails (cross-device link)
		src, err := os.Open(tmpPath)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(hashFile)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
	// Update DB entry
	err = c.DBSet(tile, dbEntry{Hash: hash, Timestamp: time.Now()})
	if err != nil {
		return err
	}
	return nil
}

// CleanupOldFiles deletes cache files older than the given duration.
func (c *Cache) CleanupOldFiles(olderThan time.Duration) error {
	root := c.getTilesPath()
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
	if c.db != nil {
		c.db.Close()
	}
	return nil
}

func (c *Cache) DBKey(tile model.Tile) []byte {
	systemBytes := []byte(tile.System)
	key := make([]byte, 9+len(systemBytes)) // 1+4+4 for Z,X,Y + system

	// Store Z as uint8, X, Y as binary integers
	key[0] = uint8(tile.Z)
	binary.LittleEndian.PutUint16(key[1:3], uint16(tile.X))
	binary.LittleEndian.PutUint16(key[4:6], uint16(tile.Y))

	// Store system string at the end without length
	copy(key[9:], systemBytes)

	return key
}

func (c *Cache) DBSet(tile model.Tile, data dbEntry) error {
	if c.db == nil {
		return fmt.Errorf("badger db is not initialized")
	}
	val, err := data.Marshal()
	if err != nil {
		return err
	}
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set(c.DBKey(tile), val)
	})
}

func (c *Cache) DBGet(tile model.Tile) (*dbEntry, error) {
	if c.db == nil {
		return nil, fmt.Errorf("badger db is not initialized")
	}
	var valCopy []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(c.DBKey(tile))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		valCopy = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	var entry dbEntry
	err = entry.Unmarshal(valCopy)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (c *Cache) DBHas(tile model.Tile) bool {
	if c.db == nil {
		return false
	}
	err := c.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(c.DBKey(tile))
		return err
	})
	return err == nil
}

func (d dbEntry) Marshal() ([]byte, error) {
	tsBytes, err := d.Timestamp.MarshalBinary()
	if err != nil {
		return nil, err
	}
	hashBytes := []byte(d.Hash)
	result := make([]byte, 4+len(hashBytes)+len(tsBytes))
	binary.LittleEndian.PutUint32(result[0:4], uint32(len(hashBytes)))
	copy(result[4:4+len(hashBytes)], hashBytes)
	copy(result[4+len(hashBytes):], tsBytes)
	return result, nil
}

func (d *dbEntry) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("data too short to unmarshal")
	}
	hashLen := binary.LittleEndian.Uint32(data[0:4])
	if len(data) < int(4+hashLen) {
		return fmt.Errorf("data too short for hash")
	}
	d.Hash = string(data[4 : 4+hashLen])
	err := d.Timestamp.UnmarshalBinary(data[4+hashLen:])
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) getTilesPath() string {
	return filepath.Join(c.path, "tiles")
}

func (c *Cache) getDBPath() string {
	return filepath.Join(c.path, "badger")
}

func (c *Cache) getFilename(hash string) (string, string) {
	hashDir := filepath.Join(c.getTilesPath(), hash[:3], hash[3:6])
	hashFile := filepath.Join(hashDir, hash+".png")
	return hashDir, hashFile
}
