package salesdiscount

import (
	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
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

	// An extension mutates its host module's data, so it must be guarded at
	// least as strictly as the module's own write routes.
	app.Router.POST(
		"/api/sales/orders/:id/discount",
		auth.RequireAuth(),
		base.RequirePermission("sales.manage"),
		applyDiscount,
	)
	return nil
}

func (e *Extension) Boot(app *core.App) error {
	return nil
}
