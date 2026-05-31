package api

import (
	"inspection-service/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunshineOfficial/golib/golog"
)

func TestInspectionAuthorizationPolicy(t *testing.T) {
	builder := NewServerBuilder(t.Context(), golog.NewLogger("test"), config.Settings{
		Port: 80,
	})
	builder.AddInspections(nil)

	t.Run("attach photo requires authorization", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/inspections/1/photo", nil)

		builder.router.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("get by task id allows internal calls without authorization", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/inspections/task/1", nil)

		builder.router.ServeHTTP(response, request)

		if response.Code == http.StatusUnauthorized {
			t.Fatalf("status = %d, route must stay open for internal service calls", response.Code)
		}
	})
}
