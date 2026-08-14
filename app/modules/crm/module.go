package crm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/crm/migrations"
	"github.com/annaselh/gorbio/modules/sales"
	"github.com/google/uuid"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "crm"
}

// Depends includes sales because Register hands it a customer resolver, so the
// sales service must already be in the container by the time CRM registers.
func (m *Module) Depends() []string {
	return []string{
		"base",
		"sales",
	}
}

func (m *Module) Migrate(app *core.App) error {
	return migrations.Init(app.DB, &Customer{})
}

func (m *Module) Register(app *core.App) error {
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}
	audit, err := base.AuditFromApp(app)
	if err != nil {
		return err
	}

	service := NewService(app.DB)
	if err := app.Services.Register(ServiceName, service); err != nil {
		return err
	}

	registerRoutes(app, auth, &handlers{service: service, audit: audit})

	// Hand sales a customer lookup so an order can reference a CRM record.
	// The dependency points this way on purpose: CRM knows about sales, sales
	// knows nothing about CRM, so sales still works with CRM uninstalled.
	if orders, err := sales.ServiceFromApp(app); err == nil {
		orders.WithCustomerResolver(func(ctx context.Context, tenantID, customerID uuid.UUID) (string, error) {
			customer, err := service.Get(ctx, tenantID, customerID)
			if err != nil {
				return "", err
			}
			if customer.Status != CustomerStatusActive {
				return "", fmt.Errorf("%w: customer %s is inactive", ErrInvalidInput, customer.Code)
			}
			return customer.Name, nil
		})
	}

	slog.Info("module registered", "module", m.Name())
	return nil
}

func (m *Module) Boot(app *core.App) error {
	slog.Info("module booted", "module", m.Name())
	return nil
}

// ServiceFromApp resolves the CRM service from the container.
func ServiceFromApp(app *core.App) (*Service, error) {
	return core.ServiceAs[*Service](app, ServiceName)
}
