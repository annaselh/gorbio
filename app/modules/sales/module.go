package sales

import "github.com/annaselh/gorbio/core"

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "sales"
}

func (m *Module) Depends() []string {
	return []string{
		"base",
	}
}

func (m *Module) Register(app *core.App) error {

	println("Registering sales module")

	app.Router.GET(
		"/api/sales/orders",
		listOrders,
	)
	return nil
}

func (m *Module) Boot(app *core.App) error {
	println("Booting sales module")
	return nil
}
