package provider

import (
	"bytes"
	"fmt"
	"io"

	"github.com/i0tool5/mbtiles-go"
	"github.com/samber/do/v2"

	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
	"github.com/willie68/go_mapproxy/internal/model"
)

type metadata struct {
	Name    string
	Format  string
	Maxzoom int
	Minzoom int
	BBox    *mercantile.Bbox
}

type mbtilesProvider struct {
	log  *logging.Logger
	db   *mbtiles.MBtiles
	fb   string
	meta metadata
	inj  do.Injector
}

func NewMBTilesProvider(config Config, inj do.Injector) *mbtilesProvider {
	log := logging.New().WithName("mbtiles")
	db, err := mbtiles.Open(config.Path)
	if err != nil {
		log.Errorf("failed to open mbtiles database: %v", err)
	}
	tf := db.GetTileFormat()
	log.Infof("mbtiles format: %s", tf.String())
	meta, err := db.ReadMetadata()
	if err != nil {
		log.Errorf("failed to read mbtiles metadata: %v", err)
	}
	log.Infof("mbtiles metadata: %+v", meta)
	mbt := &mbtilesProvider{
		log: log,
		db:  db,
		fb:  config.Fallback,
		inj: inj,
	}
	mbt.parseMetadata(meta)
	return mbt
}

func (s *mbtilesProvider) Tile(tile model.Tile) (io.ReadCloser, error) {
	var data []byte
	if tile.Z == 0 {
		return s.fallback(tile)
	}
	ymax := 1 << tile.Z
	y := ymax - tile.Y - 1
	if tile.Z < s.meta.Minzoom || tile.Z > s.meta.Maxzoom {
		return nil, fmt.Errorf("zoom level %d out of bounds (%d - %d)", tile.Z, s.meta.Minzoom, s.meta.Maxzoom)
	}
	if s.meta.BBox != nil {
		tbox := mercantile.ULBounds(mercantile.TileID{X: tile.X, Y: tile.Y, Z: tile.Z})
		if tbox.Left > s.meta.BBox.Right || tbox.Right < s.meta.BBox.Left || tbox.Top < s.meta.BBox.Bottom || tbox.Bottom > s.meta.BBox.Top {
			return nil, fmt.Errorf("tile %d/%d/%d out of bounds", tile.Z, tile.X, tile.Y)
		}
	}
	err := s.db.ReadTile(int64(tile.Z), int64(tile.X), int64(y), &data)
	//err := s.db.ReadTile(int64(0), int64(0), int64(0), &data)
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("failed to read tile: %v", err)
	}
	return io.NopCloser(io.Reader(bytes.NewReader(data))), nil
}

func (s *mbtilesProvider) fallback(tile model.Tile) (io.ReadCloser, error) {
	ts, err := do.InvokeNamed[Service](s.inj, s.fb)
	if err != nil {
		s.log.Errorf("System error: %v", err)
		return nil, err
	}
	return ts.Tile(tile)
}

func (s *mbtilesProvider) parseMetadata(meta map[string]any) {
	s.meta.Name, _ = meta["name"].(string)
	s.meta.Format, _ = meta["format"].(string)
	if maxzoom, ok := meta["maxzoom"].(int); ok {
		s.meta.Maxzoom = int(maxzoom)
	}
	if minzoom, ok := meta["minzoom"].(int); ok {
		s.meta.Minzoom = int(minzoom)
	}
	if bbox, ok := meta["bounds"].([]float64); ok {
		if len(bbox) == 4 {
			s.meta.BBox = &mercantile.Bbox{Left: bbox[0], Bottom: bbox[1], Right: bbox[2], Top: bbox[3]}
		}
	}
}
