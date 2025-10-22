package provider

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/i0tool5/mbtiles-go"

	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type mbtilesProvider struct {
	log *logging.Logger
	db  *mbtiles.MBtiles
}

func NewMBTilesProvider(config Config) *mbtilesProvider {
	db, err := mbtiles.Open(config.Path)
	if err != nil {
		log.Fatalf("failed to open mbtiles database: %v", err)
	}
	mbt := &mbtilesProvider{
		log: logging.New().WithName("mbtiles"),
		db:  db,
	}
	return mbt
}

func (s *mbtilesProvider) Tile(tile model.Tile) (io.ReadCloser, error) {
	var data []byte
	//	ymax := 1 << tile.Z
	//	y := ymax - tile.Y - 1
	err := s.db.ReadTile(int64(tile.Z), int64(tile.X), int64(tile.Y), &data)
	//err := s.db.ReadTile(int64(0), int64(0), int64(0), &data)
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("failed to read tile: %v", err)
	}
	return io.NopCloser(io.Reader(bytes.NewReader(data))), nil
}
