package provider

import (
	"bytes"
	"fmt"
	"io"

	"github.com/i0tool5/mbtiles-go"
	"github.com/samber/do/v2"

	"github.com/willie68/go_mapproxy/internal/assets"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
	"github.com/willie68/go_mapproxy/internal/model"
)

type providerService interface {
	HasProvider(providerName string) bool
	FTile(tile model.Tile) (io.ReadCloser, error)
}

type metadata struct {
	Name    string
	Format  string
	Maxzoom int
	Minzoom int
	BBox    *mercantile.Bbox
}

type mbtilesProvider struct {
	name   string
	log    *logging.Logger
	db     *mbtiles.MBtiles
	fbname string
	fb     bool
	meta   metadata
	inj    do.Injector
}

func NewMBTilesProvider(name string, config Config, inj do.Injector) *mbtilesProvider {
	log := logging.New().WithName(fmt.Sprintf("mbtiles: %s", name))
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
		name:   name,
		log:    log,
		db:     db,
		inj:    inj,
		fbname: config.Fallback,
		fb:     config.Fallback != "",
	}
	mbt.parseMetadata(meta)
	return mbt
}

func (s *mbtilesProvider) Tile(tile model.Tile) (io.ReadCloser, error) {
	var data []byte
	ymax := 1 << tile.Z
	y := ymax - tile.Y - 1
	if tile.Z < s.meta.Minzoom || tile.Z > s.meta.Maxzoom {
		if s.fb {
			return s.fallback(tile)
		}
		s.log.Errorf("zoom level %d out of bounds (%d - %d)", tile.Z, s.meta.Minzoom, s.meta.Maxzoom)
		return assets.EmptyPNG(), nil

	}
	if s.meta.BBox != nil {
		tbox := mercantile.ULBounds(mercantile.TileID{X: tile.X, Y: tile.Y, Z: tile.Z})
		if tbox.Left > s.meta.BBox.Right || tbox.Right < s.meta.BBox.Left || tbox.Top < s.meta.BBox.Bottom || tbox.Bottom > s.meta.BBox.Top {
			if s.fb {
				return s.fallback(tile)
			}
			s.log.Errorf("tile %d/%d/%d out of bounds", tile.Z, tile.X, tile.Y)
			return assets.EmptyPNG(), nil
		}
	}
	err := s.db.ReadTile(int64(tile.Z), int64(tile.X), int64(y), &data)
	if err != nil || len(data) == 0 {
		if s.fb {
			return s.fallback(tile)
		}
		s.log.Errorf("failed to read tile: %v", err)
		return assets.EmptyPNG(), nil
	}
	return io.NopCloser(io.Reader(bytes.NewReader(data))), nil
}

func (s *mbtilesProvider) fallback(tile model.Tile) (io.ReadCloser, error) {
	if s.fbname == "" || !s.fb {
		return assets.EmptyPNG(), nil
	}
	fbts, err := do.InvokeAs[providerService](s.inj)
	if err != nil {
		s.log.Errorf("failed to invoke fallback provider '%s': %v", s.fbname, err)
		s.fb = false
		return assets.EmptyPNG(), nil
	}
	if !fbts.HasProvider(s.fbname) {
		s.log.Errorf(fmt.Sprintf("fallback provider '%s' not found", s.fbname))
		s.fb = false
		return assets.EmptyPNG(), nil
	}
	tile.Provider = s.fbname
	return fbts.FTile(tile)
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
