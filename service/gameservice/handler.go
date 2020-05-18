package gameservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DHBosworth/technichalexercise/backend"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// New creates a new game service handler with the provided game data source
func New(ds backend.GameDataSource, router *mux.Router) *Handler {
	if router == nil {
		router = mux.NewRouter()
	}

	gameService := &Handler{
		ds:     ds,
		Router: router,
	}

	gameService.RegisterEndpoints()

	return gameService
}

// Handler is the http.Handler for the games service
type Handler struct {
	*mux.Router
	ds backend.GameDataSource
}

var nonGetMethods = []string{
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

// RegisterEndpoints registers the the game services endpoint handlers with the
// router
func (gs *Handler) RegisterEndpoints() {
	log.Debugf("Registering GetGame endpoint")
	getGamePath := gs.Path("/{id:[0-9]+}")
	getGamePath.Methods(http.MethodGet).HandlerFunc(gs.getGameEndpoint)
	// getGamePath.Methods(...nonGetMethods).HandlerFunc(invalidMethod)

	log.Debugf("Registering Report endpoint")
	reportPath := gs.Path("/report")
	reportPath.Methods(http.MethodGet).HandlerFunc(gs.reportEndpoint)
	// reportPath.Methods(...nonGetMethods).HandlerFunc(invalidMethod)

	gs.NotFoundHandler = http.HandlerFunc(invalidEnpoint)
}

func invalidMethod(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(backend.Error{
		Msg: fmt.Sprintf("Method %s invalid", r.Method),
	})
}

// reportEndpoint is the handler for the /report endpoint
func (gs *Handler) reportEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Get report")

	w.Header().Set("Content-Type", "application/json")
	report, err := gs.ds.Report()
	if err != nil {
		reportError(w, err)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(report)
}

// reportError encodes an error into json format and sets up the http.Response
func reportError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(backend.Error{Msg: err.Error()})
}

// getGameEndpoint is the handler for the /games/<game_id> endpoint
func (gs *Handler) getGameEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	gameID, reqHasID := vars["id"]
	if !reqHasID {
		noIDError(w)
		return
	}

	log.Debugf("Get Game %s", gameID)

	game, err := gs.ds.Game(gameID)
	if err != nil {
		gameNotFoundError(w, gameID, err)
		return
	}

	respEncoder := json.NewEncoder(w)
	respEncoder.Encode(game)
}

func noIDError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(backend.Error{
		Msg: "Invalid request. Require parameter id.",
	})
}

func gameNotFoundError(w http.ResponseWriter, id string, err error) {
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusNotFound)
	if err := enc.Encode(backend.Error{Msg: err.Error()}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func invalidEnpoint(w http.ResponseWriter, r *http.Request) {

}
