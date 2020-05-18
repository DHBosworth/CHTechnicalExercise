package gameservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DHBosworth/technichalexercise/backend"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func hasRoute(router *mux.Router, routeS string) bool {
	routeExists := false

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		p, err := route.GetPathTemplate()
		if err == nil && p == routeS {
			routeExists = true
		}
		return nil
	})

	return routeExists
}

func TestHandler_RegisterEndpoints(t *testing.T) {
	tests := []struct {
		name  string
		gs    *Handler
		check func(handler *Handler)
	}{
		{
			name: "Has get Game route",
			gs:   &Handler{Router: mux.NewRouter()},
			check: func(handler *Handler) {
				if !hasRoute(handler.Router, "/{id:[0-9]+}") {
					t.Errorf("Get game endpoint not registered")
				}
			},
		},
		{
			name: "Has report route",
			gs:   &Handler{Router: mux.NewRouter()},
			check: func(handler *Handler) {
				if !hasRoute(handler.Router, "/report") {
					t.Errorf("Game report endpoint not registered")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.gs.RegisterEndpoints()
			tt.check(tt.gs)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		ds backend.GameDataSource
	}
	tests := []struct {
		name  string
		args  args
		check func(handler *Handler)
	}{
		{
			name: "Game Service Router created",
			args: args{ds: nil},
			check: func(handler *Handler) {
				if handler.Router == nil {
					t.Errorf("Game Service router not created")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.ds, nil)
			tt.check(got)
		})
	}
}

func createTime(s string) backend.EpochToReadable {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(fmt.Sprintf("createTime: unable to parse time %s", s))
	}

	return backend.EpochToReadable(t)
}

type mockGameDataSource struct{}

var mockGames = []backend.Game{
	{
		Title:       "Dummy",
		Description: "A game that exists solely for testing",
		By:          "me",
		Platform:    []string{"PC - obviously"},
		AgeRating:   "42+",
		Likes:       42,
		Comments: []backend.Comment{
			{
				User:        "Jacqueline Dodson",
				Message:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
				DateCreated: createTime("2004-03-19"),
				Like:        5,
			},
			{
				User:        "Courtney Knapp",
				Message:     "Nam urna ipsum, blandit vel ex ac, imperdiet venenatis justo.",
				DateCreated: createTime("1991-04-12"),
				Like:        1,
			},
		},
	},
	{
		Title:       "Solitary Voyage",
		Description: "Decsription goes here",
		By:          "Jimmie Bassett",
		Platform:    []string{"PC", "XBOX"},
		AgeRating:   "6+",
		Likes:       99,
		Comments: []backend.Comment{
			{
				User:        "Jacqueline Dodson",
				Message:     "Mauris blandit orci at magna venenatis euismod.",
				DateCreated: createTime("2001-08-16"),
				Like:        9,
			},
		},
	},
}

func (mockGameDataSource) Game(id string) (game backend.Game, err error) {
	switch id {
	case "1":
		game = mockGames[0]
	case "2":
		game = mockGames[1]
	default:
		err = fmt.Errorf("Unable to get game")
	}

	return game, err
}

var mockReport = backend.Report{}

func (mockGameDataSource) Report() (report backend.Report, err error) {

	return report, nil
}

func mustReq(method string, path string) *http.Request {
	r, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic("Unable to create request: " + err.Error())
	}

	return r
}

func TestHandler_ServeHTTP(t *testing.T) {
	gs := New(mockGameDataSource{}, nil)

	tests := []struct {
		name  string
		req   *http.Request
		check func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:  "Get game 1",
			req:   mustReq(http.MethodGet, "/1"),
			check: checkGame(mockGames[0]),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			gs.ServeHTTP(resp, tt.req)
			tt.check(t, resp)
		})
	}
}

func checkGame(expected backend.Game) func(t *testing.T, resp *httptest.ResponseRecorder) {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		assert.Equal(t, 200, resp.Code, "Request should have succeeded")

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unable to read response: %v", err)
		}

		var game backend.Game
		if err := json.Unmarshal(body, &game); err != nil {
			t.Errorf("Error decoding response: %v", err)
		}
		assert.Equal(t, expected, game, "Should have returned the first game")
	}
}

func checkGameError(expectedErr backend.Error) func(t *testing.T, resp *httptest.ResponseRecorder) {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		assert.Equal(t, http.StatusNotFound, resp.Code, "Game should not have been found")

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unable to read response: %v", err)
		}

		var errValue backend.Error
		if err := json.Unmarshal(body, &errValue); err != nil {
			t.Errorf("Error decoding response: %v", err)
		}
		assert.Equal(t, expectedErr, errValue, "Should have returned the first game")
	}
}

func TestHandler_getGameEndpoint(t *testing.T) {
	gs := New(mockGameDataSource{}, nil)

	tests := []struct {
		name  string
		req   *http.Request
		check func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:  "Get Game 1",
			req:   mux.SetURLVars(mustReq(http.MethodGet, "/1"), map[string]string{"id": "1"}),
			check: checkGame(mockGames[0]),
		},
		{
			name:  "Get Game 2",
			req:   mux.SetURLVars(mustReq(http.MethodGet, "/2"), map[string]string{"id": "2"}),
			check: checkGame(mockGames[1]),
		},
		{
			name:  "Get game not found",
			req:   mux.SetURLVars(mustReq(http.MethodGet, "/3"), map[string]string{"id": "3"}),
			check: checkGameError(backend.Error{Msg: "Unable to get game"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			gs.getGameEndpoint(resp, tt.req)
			tt.check(t, resp)
		})
	}
}
