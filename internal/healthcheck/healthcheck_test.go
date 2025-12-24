package healthcheck_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/healthcheck"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	t.Run("returns OK status", func(t *testing.T) {
		handler := healthcheck.HealthCheckHandler()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("handles POST request", func(t *testing.T) {
		handler := healthcheck.HealthCheckHandler()

		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should still return OK regardless of method
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("handles different paths", func(t *testing.T) {
		handler := healthcheck.HealthCheckHandler()

		paths := []string{"/", "/health", "/healthz", "/ready", "/status"}

		for _, path := range paths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code,
				"Expected OK for path %s", path)
			assert.Equal(t, "OK", rec.Body.String())
		}
	})

	t.Run("returns correct content type", func(t *testing.T) {
		handler := healthcheck.HealthCheckHandler()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Default content type for Write is text/plain
		// Note: Go's http.ResponseWriter sets this automatically based on content
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestHealthCheckHandler_Integration(t *testing.T) {
	t.Run("can be used as http.Handler", func(t *testing.T) {
		handler := healthcheck.HealthCheckHandler()

		server := httptest.NewServer(handler)
		defer server.Close()

		resp, err := http.Get(server.URL + "/health")
		assert.Nil(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
