package apiv1

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
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

type providerService interface {
	HasProvider(providerName string) bool
	FTile(tile model.Tile) (io.ReadCloser, error)
}

type XYZHandler struct {
	log     *slog.Logger
	tiles   providerService
	metrics *measurement.Service
}

func NewXYZHandler(inj do.Injector) *chi.Mux {
	th := &XYZHandler{
		log:     logging.New("api"),
		tiles:   do.MustInvokeAs[providerService](inj),
		metrics: do.MustInvokeAs[*measurement.Service](inj),
	}
	router := chi.NewRouter()
	router.Get("/{provider}/xyz/{z}/{x}/{y}.png", th.GetSystemHandler(inj))
	return router
}

func (h *XYZHandler) GetSystemHandler(inj do.Injector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td := h.metrics.Start("getTile")
		defer td.Stop()

		// URL: /tileserver/{provider}/xyz/{z}/{x}/{y}.png
		h.log.Info(fmt.Sprintf("path: %s", r.URL.Path))
		tile, err := h.getRequestParameter(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Path error: %s", err.Error()), http.StatusBadRequest)
			return
		}

		rd, err := h.tiles.FTile(tile)
		if err != nil {
			h.log.Error("System error: %v", err)
			http.Error(w, fmt.Sprintf("System error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer rd.Close()

		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, rd)
	})
}

func (h *XYZHandler) getRequestParameter(r *http.Request) (tile model.Tile, err error) {
	tile.Provider = chi.URLParam(r, "provider")
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

	if !h.tiles.HasProvider(tile.Provider) {
		return tile, errors.New("unknown provider")
	}
	if !h.isValidXYZCoord(tile.X, tile.Y, tile.Z) {
		return tile, errors.New("invalid tile coordinates")
	}
	return tile, nil
}

// Checks if the given TMS coordinates are valid for the given zoom level.
func (h *XYZHandler) isValidXYZCoord(x, y, zoom int) bool {
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
