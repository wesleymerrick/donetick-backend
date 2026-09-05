package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"donetick.com/core/config"
	"github.com/gin-gonic/gin"
)

func newTestRouter(basePath string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&config.Config{
		Server:   config.ServerConfig{ServeFrontend: true},
		BasePath: basePath,
	})
	router := gin.New()
	Routes(router, h)
	return router
}

func TestBasePathInjection(t *testing.T) {
	router := newTestRouter("/donetick")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `window.__BASE_PATH__="/donetick";`) {
		t.Errorf("expected injected script tag, got body: %s", body)
	}
}

func TestBasePathInjectionOnSPAFallback(t *testing.T) {
	router := newTestRouter("/donetick")

	req := httptest.NewRequest(http.MethodGet, "/chores", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `window.__BASE_PATH__="/donetick";`) {
		t.Errorf("expected injected script tag on SPA fallback route")
	}
}

func TestNoBasePathLeavesIndexUntouched(t *testing.T) {
	router := newTestRouter("")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "__BASE_PATH__") {
		t.Errorf("expected no injection when BasePath is unset, got body: %s", w.Body.String())
	}
}
