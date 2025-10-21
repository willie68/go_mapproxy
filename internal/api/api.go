package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
	"github.com/willie68/go_mapproxy/internal/utils/measurement"
)

type tileserverService interface {
	HasSystem(name string) bool
	FTile(tile model.Tile) (io.ReadCloser, error)
}

type TMSHandler struct {
	log     *logging.Logger
	tiles   tileserverService
	metrics *measurement.Service
}

func NewTMSHandler(inj do.Injector) *chi.Mux {
	th := &TMSHandler{
		log:     logging.New().WithName("api"),
		tiles:   do.MustInvokeAs[tileserverService](inj),
		metrics: do.MustInvokeAs[*measurement.Service](inj),
	}
	router := chi.NewRouter()
	router.Get("/{system}/xyz/{z}/{x}/{y}.png", th.GetSystemHandler(inj))
	return router
}

func (h *TMSHandler) GetSystemHandler(inj do.Injector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td := h.metrics.Start("getTile")
		defer td.Stop()

		// URL: /tileserver/{system}/xyz/{z}/{x}/{y}.png
		h.log.Infof("path: %s", r.URL.Path)
		tile, err := h.getRequestParameter(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Path error: %s", err.Error()), http.StatusBadRequest)
			return
		}

		rd, err := h.tiles.FTile(tile)
		if err != nil {
			h.log.Errorf("System error: %v", err)
			http.Error(w, fmt.Sprintf("System error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer rd.Close()

		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, rd)
	})
}

func (h *TMSHandler) getRequestParameter(r *http.Request) (tile model.Tile, err error) {
	tile.System = chi.URLParam(r, "system")
	zs := chi.URLParam(r, "z")
	xs := chi.URLParam(r, "x")
	ys := chi.URLParam(r, "y")

	tile.Z, err = strconv.Atoi(zs)
	if err != nil {
		return tile, errors.New("error in zoom level")
	}
	tile.X, err = strconv.Atoi(xs)
	if err != nil {
		return tile, errors.New("error in x axis")
	}
	ys = strings.TrimSuffix(ys, filepath.Ext(ys))
	tile.Y, err = strconv.Atoi(ys)
	if err != nil {
		return tile, errors.New("error in y axis")
	}

	if !h.tiles.HasSystem(tile.System) {
		return tile, errors.New("unknown system")
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
