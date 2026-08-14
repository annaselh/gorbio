package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSEngine(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORS(origins))
	engine.GET("/api/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	engine.POST("/api/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	return engine
}

func TestCORSAllowsListedOriginWithCredentials(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected the origin echoed back, got %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("cookie auth needs credentials allowed, got %q", got)
	}
	if got := recorder.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary: Origin is required so caches do not cross origins, got %q", got)
	}
}

func TestCORSNeverEmitsWildcardOrigin(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Header().Get("Access-Control-Allow-Origin") == "*" {
		t.Fatal("credentialed CORS must never answer with a wildcard origin")
	}
}

func TestCORSOmitsHeadersForUnknownOrigin(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	request.Header.Set("Origin", "http://evil.example")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("an unlisted origin must receive no CORS grant, got %q", got)
	}
}

func TestCORSRejectsPreflightFromUnknownOrigin(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodOptions, "/api/ping", nil)
	request.Header.Set("Origin", "http://evil.example")
	request.Header.Set("Access-Control-Request-Method", "POST")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for an unlisted preflight, got %d", recorder.Code)
	}
}

func TestCORSAnswersPreflightForListedOrigin(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodOptions, "/api/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", "POST")
	request.Header.Set("Access-Control-Request-Headers", "Content-Type")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for an allowed preflight, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("requested headers should be echoed, got %q", got)
	}
}

func TestCORSIgnoresSameOriginRequests(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173"})

	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("a request without Origin must pass through, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("no CORS headers belong on a same-origin response, got %q", got)
	}
}

func TestCORSTreatsTrailingSlashAsSameOrigin(t *testing.T) {
	engine := newCORSEngine([]string{"http://localhost:5173/"})

	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("a configured trailing slash should still match, got %q", got)
	}
}
