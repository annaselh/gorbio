package base

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newPrincipalWith(permissions ...string) Principal {
	set := make(map[string]struct{}, len(permissions))
	for _, code := range permissions {
		set[code] = struct{}{}
	}
	return Principal{
		UserID:      uuid.New(),
		SessionID:   uuid.New(),
		TenantID:    uuid.New(),
		Permissions: set,
	}
}

// runGuarded exercises RequirePermission with an optional principal already on
// the context, mimicking RequireAuth having run (or not).
func runGuarded(t *testing.T, code string, principal *Principal) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/guarded", func(c *gin.Context) {
		if principal != nil {
			c.Set(principalContextKey, *principal)
		}
		c.Next()
	}, RequirePermission(code), func(c *gin.Context) {
		c.String(http.StatusOK, "reached")
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/guarded", nil))
	return recorder
}

func TestRequirePermissionAllowsHolder(t *testing.T) {
	principal := newPrincipalWith("inventory.read")
	recorder := runGuarded(t, "inventory.read", &principal)

	if recorder.Code != http.StatusOK {
		t.Fatalf("holder of the permission should pass, got %d", recorder.Code)
	}
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	principal := newPrincipalWith("inventory.read")
	recorder := runGuarded(t, "inventory.manage", &principal)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("a read-only principal must not reach a manage route, got %d", recorder.Code)
	}
	if recorder.Body.String() == "reached" {
		t.Fatal("handler ran despite the guard rejecting the request")
	}
}

// A route wired without RequireAuth has no principal on the context. The guard
// must fail closed rather than treat the absence as "no restrictions".
func TestRequirePermissionRejectsUnauthenticatedContext(t *testing.T) {
	recorder := runGuarded(t, "inventory.read", nil)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("missing principal must be denied, got %d", recorder.Code)
	}
}

func TestRequireAuthRejectsRequestWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &AuthService{}

	engine := gin.New()
	engine.GET("/private", service.RequireAuth(), func(c *gin.Context) {
		c.String(http.StatusOK, "reached")
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/private", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("a request with no session cookie must be rejected, got %d", recorder.Code)
	}
}

func TestPrincipalHasPermission(t *testing.T) {
	principal := newPrincipalWith("sales.read", "sales.manage")

	if !principal.HasPermission("sales.manage") {
		t.Fatal("granted permission should be reported as held")
	}
	if principal.HasPermission("inventory.manage") {
		t.Fatal("ungranted permission must not be reported as held")
	}
}

func TestPrincipalFromContextMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if _, ok := PrincipalFromContext(c); ok {
		t.Fatal("an empty context must not yield a principal")
	}
}

// Cookie flags decide whether a browser stores the session at all: Secure over
// plain HTTP is dropped, and SameSite=None is only honoured alongside Secure.
func TestSessionCookieFlagsMatchEnvironment(t *testing.T) {
	secure := &AuthService{cookieSecure: true}
	if secure.sameSiteMode() != http.SameSiteNoneMode {
		t.Fatal("a secure cross-site cookie needs SameSite=None")
	}

	insecure := &AuthService{cookieSecure: false}
	if insecure.sameSiteMode() != http.SameSiteLaxMode {
		t.Fatal("over plain HTTP the cookie must fall back to SameSite=Lax")
	}
}
