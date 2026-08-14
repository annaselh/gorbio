package sales

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
)

// registerRoutes wires the HTTP surface. Reads require sales.read; anything
// that creates or moves an order requires sales.manage.
func registerRoutes(app *core.App, auth *base.AuthService, h *handlers) {
	app.Router.GET("/api/sales/orders",
		auth.RequireAuth(), base.RequirePermission(PermissionRead), h.list)

	app.Router.GET("/api/sales/orders/:id",
		auth.RequireAuth(), base.RequirePermission(PermissionRead), h.get)

	app.Router.POST("/api/sales/orders",
		auth.RequireAuth(), base.RequirePermission(PermissionManage), h.create)

	app.Router.PUT("/api/sales/orders/:id/status",
		auth.RequireAuth(), base.RequirePermission(PermissionManage), h.updateStatus)
}
