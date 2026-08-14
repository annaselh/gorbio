package procurement

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
)

// registerRoutes wires the HTTP surface. Reads require procurement.read;
// anything that creates or moves a record requires procurement.manage.
func registerRoutes(app *core.App, auth *base.AuthService, h *handlers) {
	read := base.RequirePermission(PermissionRead)
	manage := base.RequirePermission(PermissionManage)

	app.Router.GET("/api/procurement/vendors", auth.RequireAuth(), read, h.listVendors)
	app.Router.GET("/api/procurement/vendors/:id", auth.RequireAuth(), read, h.getVendor)
	app.Router.POST("/api/procurement/vendors", auth.RequireAuth(), manage, h.createVendor)
	app.Router.PUT("/api/procurement/vendors/:id", auth.RequireAuth(), manage, h.updateVendor)
	app.Router.PUT("/api/procurement/vendors/:id/status", auth.RequireAuth(), manage, h.setVendorStatus)

	app.Router.GET("/api/procurement/orders", auth.RequireAuth(), read, h.listPurchaseOrders)
	app.Router.GET("/api/procurement/orders/:id", auth.RequireAuth(), read, h.getPurchaseOrder)
	app.Router.POST("/api/procurement/orders", auth.RequireAuth(), manage, h.createPurchaseOrder)
	app.Router.PUT("/api/procurement/orders/:id/status", auth.RequireAuth(), manage, h.updatePurchaseStatus)
}
