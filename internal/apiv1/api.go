package apiv1

import (
	"github.com/go-chi/chi/v5"
	"github.com/willie68/go_mapproxy/internal/logging"
)

// defining all sub pathes for api v1
const (
	// APIVersion the actual implemented api version
	APIVersion = "1"
)

var logger = logging.New("apiv1")

// Handler a http REST interface handler
type Handler interface {
	// Routes get the routes
	Routes() (string, *chi.Mux)
}
