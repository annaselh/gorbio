package inventory

import (
	"log/slog"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/inventory/migrations"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "inventory"
}

func (m *Module) Depends() []string {
	return []string{
		"base",
	}
}

func (m *Module) Migrate(app *core.App) error {
	return migrations.Init(app.DB, &StockItem{})
}

func (m *Module) Register(app *core.App) error {
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}

	service := NewService(app.DB)
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

// ServiceFromApp resolves the inventory service from the container.
func ServiceFromApp(app *core.App) (*Service, error) {
	return core.ServiceAs[*Service](app, ServiceName)
}
