package salesdiscount

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/sales"
)

type Extension struct{}

func New() *Extension {
	return &Extension{}
}

func (e *Extension) Name() string {
	return "sales-discount"
}

func (e *Extension) Module() string {
	return "sales"
}

func (e *Extension) Depends() []string {
	return nil
}

func (e *Extension) Register(app *core.App) error {
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}

	// Reach the host module through its published service rather than touching
	// its tables directly, so the discount goes through the same validation and
	// recalculation the module applies to its own writes.
	orders, err := sales.ServiceFromApp(app)
	if err != nil {
		return err
	}

	// An extension mutates its host module's data, so it must be guarded at
	// least as strictly as the module's own write routes.
	app.Router.POST(
		"/api/sales/orders/:id/discount",
		auth.RequireAuth(),
		base.RequirePermission(sales.PermissionManage),
		(&handlers{orders: orders}).applyDiscount,
	)
	return nil
}

func (e *Extension) Boot(app *core.App) error {
	return nil
}
