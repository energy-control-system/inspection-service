package api

import (
	"inspection-service/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunshineOfficial/golib/golog"
)

func TestInspectionRoutesAllowUnauthenticatedRequests(t *testing.T) {
	builder := NewServerBuilder(t.Context(), golog.NewLogger("test"), config.Settings{
		Port: 80,
	})
	builder.AddInspections(nil)

	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/inspections"},
		{method: http.MethodGet, path: "/inspections/1"},
		{method: http.MethodGet, path: "/inspections/task/1"},
		{method: http.MethodGet, path: "/inspections/brigades/1"},
		{method: http.MethodPost, path: "/inspections/1/photo"},
		{method: http.MethodPatch, path: "/inspections/1/finish"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(route.method, route.path, nil)

			builder.router.ServeHTTP(response, request)

			if response.Code == http.StatusUnauthorized {
				t.Fatalf("status = %d, route must be open without authorization", response.Code)
			}
		})
	}
}
