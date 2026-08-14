package sales

import (
	"log/slog"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/sales/migrations"
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

func (m *Module) Migrate(app *core.App) error {
	return migrations.Init(app.DB, &Order{}, &OrderLine{})
}

func (m *Module) Register(app *core.App) error {
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}

	service := NewService(app.DB)
	// Published so extensions can act on orders without importing the module's
	// HTTP layer; sales-discount resolves it by this name.
	if err := app.Services.Register(ServiceName, service); err != nil {
		return err
	}

	registerRoutes(app, auth, &handlers{service: service})

	slog.Info("module registered", "module", m.Name())
	return nil
}

func (m *Module) Boot(app *core.App) error {
	slog.Info("module booted", "module", m.Name())
	return nil
}

// ServiceFromApp resolves the sales service from the container. Extensions use
// it rather than constructing their own, so they share the module's
// transaction and validation rules.
func ServiceFromApp(app *core.App) (*Service, error) {
	return core.ServiceAs[*Service](app, ServiceName)
}
