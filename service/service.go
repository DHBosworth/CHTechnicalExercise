package service

import (
	"github.com/DHBosworth/technichalexercise/backend"
	"github.com/DHBosworth/technichalexercise/service/gameservice"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// Handler represents the http.Handler for the service
type Handler struct {
	*mux.Router
	dataSource backend.ServiceDataSource
}

// New creates a new service with the given data source
func New(ds backend.ServiceDataSource, router *mux.Router) *Handler {
	if router == nil {
		router = mux.NewRouter()
	}

	s := &Handler{
		dataSource: ds,
		Router:     router,
	}

	s.RegisterEndpoints()

	return s
}

const (
	gamesEnpointPath = "/games"
)

// RegisterEndpoints registers the services endpoints with the router
func (s *Handler) RegisterEndpoints() {
	log.Debugf("Registering Games endpoint")

	gamesRouter := s.PathPrefix(gamesEnpointPath).Subrouter()
	gameservice.New(s.dataSource, gamesRouter) // Game service endpoints are registered here
}
