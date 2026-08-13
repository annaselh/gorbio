package salesdiscount

import "github.com/annaselh/gorbio/core"

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
	app.Router.POST(
		"/api/sales/orders/:id/discount",
		applyDiscount,
	)
	return nil
}

func (e *Extension) Boot(app *core.App) error {
	return nil
}
