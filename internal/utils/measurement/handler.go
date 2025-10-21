package measurement

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/samber/do/v2"
)

func Routes(inj do.Injector) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", GetMetricsHandler(inj))
	router.Post("/reset", ResetMetricsHandler(inj))
	router.Post("/reset/{name}", ResetMonitorHandler(inj))
	return router
}

func GetMetricsHandler(inj do.Injector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms := do.MustInvoke[*Service](inj)
		datas := ms.Datas()

		render.Status(r, http.StatusOK)
		render.JSON(w, r, datas)
	})
}

func ResetMetricsHandler(inj do.Injector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms := do.MustInvoke[*Service](inj)
		ms.Reset()

		render.Status(r, http.StatusOK)
	})
}

func ResetMonitorHandler(inj do.Injector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		ms := do.MustInvoke[*Service](inj)
		point := ms.Point(name)
		point.Reset()

		render.Status(r, http.StatusOK)
	})
}
