package service

import (
	"testing"

	"github.com/DHBosworth/technichalexercise/backend"
	"github.com/gorilla/mux"
)

type dummyDataSource struct {
}

func (dummyDataSource) Game(id string) (game backend.Game, err error) {
	return game, err
}

func (dummyDataSource) Report() (report backend.Report, err error) {
	return report, nil
}

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

func TestNew(t *testing.T) {
	type args struct {
		ds backend.ServiceDataSource
	}
	tests := []struct {
		name string
		args args
		pass func(handler *Handler) bool
	}{
		{
			name: "Simple",
			args: args{ds: dummyDataSource{}},
			pass: func(handler *Handler) bool {
				return handler.dataSource == dummyDataSource{} && handler.Router != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.ds, nil); !tt.pass(got) {
				t.Errorf("New() = %v", got)
			}
		})
	}
}

func TestHandler_RegsiterEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		serviceHandler *Handler
		check          func(handler *Handler)
	}{
		{
			name:           "Has games routes",
			serviceHandler: &Handler{Router: mux.NewRouter()},
			check: func(handler *Handler) {
				if !hasRoute(handler.Router, "/games/{id:[0-9]+}") {
					t.Errorf("Get Game endpoint not registered")
				}

				if !hasRoute(handler.Router, "/games/report") {
					t.Errorf("Report endpoint not registered")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.serviceHandler.RegisterEndpoints()
			tt.check(tt.serviceHandler)
		})
	}
}
