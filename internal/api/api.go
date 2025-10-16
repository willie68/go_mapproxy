package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type tileCache interface {
	Tile(tile model.Tile) (io.Reader, bool)
	Save(tile model.Tile, data io.Reader) error
	IsActive() bool
}

type tileserverService interface {
	HasSystem(name string) bool
	Tile(tile model.Tile) (io.ReadCloser, error)
}

type TMSHandler struct {
	log   *logging.Logger
	cache tileCache
	tiles tileserverService
}

func NewTMSHandler(inj do.Injector) *TMSHandler {
	return &TMSHandler{
		log:   logging.New().WithName("api"),
		cache: do.MustInvokeAs[tileCache](inj),
		tiles: do.MustInvokeAs[tileserverService](inj),
	}
}

func (h *TMSHandler) Handler(w http.ResponseWriter, r *http.Request) {
	// URL: /{system}/tms/{z}/{x}/{y}.png
	path := r.URL.Path
	h.log.Infof("path: %s", path)
	tile, err := h.getRequestParameter(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Path error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// try to get the cached tile
	if tr, ok := h.cache.Tile(tile); ok {
		h.log.Debugf("tile found in cache: %s", tile.String())
		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, tr)
		if rc, ok := tr.(io.ReadCloser); ok {
			rc.Close()
		}
		return
	}

	rd, err := h.tiles.Tile(tile)
	if err != nil {
		h.log.Errorf("System error: %v", err)
		http.Error(w, fmt.Sprintf("System error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer rd.Close()

	w.Header().Set("Content-Type", "image/png")
	// if cache is inactive simply, copy the content to the requester
	if !h.cache.IsActive() {
		io.Copy(w, rd)
		return
	}
	// otherwise read the data and write them in parallel to the cache and the requester
	data, err := io.ReadAll(rd)
	if err != nil {
		http.Error(w, "erorr reading data from wms server", http.StatusInternalServerError)
		return
	}

	wr := bytes.NewReader(data)
	err = h.cache.Save(tile, wr)
	if err != nil {
		http.Error(w, "error writing to the cache", http.StatusInternalServerError)
		return
	}
	wr = bytes.NewReader(data)
	io.Copy(w, wr)
}

func (h *TMSHandler) getRequestParameter(path string) (tile model.Tile, err error) {
	p := strings.Split(path, "/")
	if len(p) != 6 {
		return tile, errors.New("Path error")
	}
	tile.System = p[1]
	if !h.tiles.HasSystem(tile.System) {
		return tile, errors.New("unknown system")
	}
	tile.Z, err = strconv.Atoi(p[3])
	if err != nil {
		return tile, errors.New("error in zoom")
	}
	tile.X, err = strconv.Atoi(p[4])
	if err != nil {
		return tile, errors.New("error in x axis")
	}
	fn := filepath.Base(p[5])
	ys := strings.TrimSuffix(fn, filepath.Ext(fn))
	tile.Y, err = strconv.Atoi(ys)
	if err != nil {
		return tile, errors.New("error in y axis")
	}
	if !h.isValidTMSCoord(tile.X, tile.Y, tile.Z) {
		return tile, errors.New("invalid tile coordinates")
	}
	return tile, nil
}

// Checks if the given TMS coordinates are valid for the given zoom level.
func (h *TMSHandler) isValidTMSCoord(x, y, zoom int) bool {
	if zoom < 0 {
		return false
	}
	max := 1 << zoom // 2^zoom
	if x < 0 || x >= max {
		return false
	}
	if y < 0 || y >= max {
		return false
	}
	return true
}
