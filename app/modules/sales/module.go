package sales

import (
	"log/slog"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
)

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
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}

	// Every business route sits behind authentication and an explicit
	// permission. Registering a bare handler here would publish tenant data to
	// the open internet.
	app.Router.GET(
		"/api/sales/orders",
		auth.RequireAuth(),
		base.RequirePermission("sales.read"),
		listOrders,
	)

	slog.Info("module registered", "module", m.Name())
	return nil
}

func (m *Module) Boot(app *core.App) error {
	slog.Info("module booted", "module", m.Name())
	return nil
}
