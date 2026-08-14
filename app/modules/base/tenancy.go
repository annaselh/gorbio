package base

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantScope constrains a query to a single tenant.
//
// Every business table that carries a tenant_id must be read and written
// through this scope. A bare db.Where on a tenant-owned model is the classic
// multi-tenant data leak: it works in a single-tenant test database and quietly
// returns other customers' rows in production. Module code should reach for
// ScopedDB rather than app.DB directly.
func TenantScope(tenantID uuid.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// ScopedDB returns a session already constrained to the principal's tenant.
func ScopedDB(db *gorm.DB, principal Principal) *gorm.DB {
	return db.Scopes(TenantScope(principal.TenantID))
}

// RequestScopedDB resolves the principal placed on the context by RequireAuth
// and returns a tenant-scoped session. The second result is false when the
// route was not wrapped in RequireAuth, which is a wiring bug rather than a
// runtime condition - fail the request rather than fall back to an unscoped
// query.
func RequestScopedDB(c *gin.Context, db *gorm.DB) (*gorm.DB, Principal, bool) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		return nil, Principal{}, false
	}
	return ScopedDB(db, principal), principal, true
}
