package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zavieruka/video-platform/backend/internal/middleware"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestCORS_AllowlistReflectsKnownOrigin(t *testing.T) {
	h := middleware.CORSMiddleware([]string{"https://app.example"})(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	req.Header.Set("Origin", "https://app.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "https://app.example", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Vary"), "Origin")
}

func TestCORS_AllowlistRejectsUnknownOrigin(t *testing.T) {
	h := middleware.CORSMiddleware([]string{"https://app.example"})(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	req.Header.Set("Origin", "https://evil.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// Request still processed, but no CORS grant is issued to the browser.
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardAllowsAny(t *testing.T) {
	h := middleware.CORSMiddleware([]string{"*"})(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	req.Header.Set("Origin", "https://anything.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightShortCircuits(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	h := middleware.CORSMiddleware([]string{"*"})(next)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
	req.Header.Set("Origin", "https://app.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.False(t, called, "preflight must not reach the router")
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), http.MethodPatch)
}
