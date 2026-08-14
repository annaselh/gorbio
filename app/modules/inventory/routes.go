package inventory

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
)

// registerRoutes wires the HTTP surface. Reads require inventory.read; anything
// that mutates stock requires inventory.manage.
func registerRoutes(app *core.App, auth *base.AuthService, h *handlers) {
	app.Router.GET("/api/inventory/items",
		auth.RequireAuth(), base.RequirePermission(PermissionRead), h.list)

	app.Router.GET("/api/inventory/items/:id",
		auth.RequireAuth(), base.RequirePermission(PermissionRead), h.get)

	app.Router.POST("/api/inventory/items",
		auth.RequireAuth(), base.RequirePermission(PermissionManage), h.create)

	app.Router.POST("/api/inventory/items/:id/adjust",
		auth.RequireAuth(), base.RequirePermission(PermissionManage), h.adjust)
}
