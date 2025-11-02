package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/samber/do/v2"
	"github.com/willie68/go_mapproxy/internal/apiv1"
	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/utils/measurement"
)

// MetricsEndpoint endpoint subpath  for metrics
const (
	// MetricsEndpoint endpoint subpath  for metrics
	metricsEndpoint = "/metrics"
	// TileserverEndpoint endpoint subpath for tile server
	tileserverEndpoint = "/tileserver"
)

var logger = logging.New("api")

// Handler a http REST interface handler
type Handler interface {
	// Routes get the routes
	Routes() (string, *chi.Mux)
}

// APIRoutes configuring the api routes for the main REST API
func APIRoutes(inj do.Injector) (*chi.Mux, error) {
	router := chi.NewRouter()
	setDefaultHandler(router)

	// building the routes
	router.Route("/", func(r chi.Router) {
		r.Mount(tileserverEndpoint, apiv1.NewXYZHandler(inj))
		r.Mount(metricsEndpoint, measurement.Routes(inj))
	})
	// adding a file server with web client asserts
	logger.Info("api routes")

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info(fmt.Sprintf("api route: %s %s", method, route))
		return nil
	}

	if err := chi.Walk(router, walkFunc); err != nil {
		logger.Warn(fmt.Sprintf("could not walk api routes. %s", err.Error()))
	}
	return router, nil
}

func setDefaultHandler(router *chi.Mux) {
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.Recoverer,
		cors.Handler(cors.Options{
			// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-mcs-username", "X-mcs-password", "X-mcs-profile"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	)
}

// HealthRoutes returning the health routes
func HealthRoutes(inj do.Injector) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.Recoverer,
	)

	router.Route("/", func(r chi.Router) {
		r.Mount("/health/metrics", measurement.Routes(inj))
	})

	logger.Info("health api routes")
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info(fmt.Sprintf("health route: %s %s", method, route))
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		logger.Warn(fmt.Sprintf("could not walk health routes. %s", err.Error()))
	}

	return router
}
