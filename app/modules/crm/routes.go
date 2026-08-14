package crm

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
)

// registerRoutes wires the HTTP surface. Reads require crm.read; anything that
// creates or changes a record requires crm.manage.
func registerRoutes(app *core.App, auth *base.AuthService, h *handlers) {
	read := base.RequirePermission(PermissionRead)
	manage := base.RequirePermission(PermissionManage)

	app.Router.GET("/api/crm/customers", auth.RequireAuth(), read, h.list)
	app.Router.GET("/api/crm/customers/:id", auth.RequireAuth(), read, h.get)
	app.Router.POST("/api/crm/customers", auth.RequireAuth(), manage, h.create)
	app.Router.PUT("/api/crm/customers/:id", auth.RequireAuth(), manage, h.update)
	app.Router.PUT("/api/crm/customers/:id/status", auth.RequireAuth(), manage, h.setStatus)
}
