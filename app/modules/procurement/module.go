package procurement

import (
	"context"
	"log/slog"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/inventory"
	"github.com/annaselh/gorbio/modules/procurement/migrations"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "procurement"
}

// Depends includes inventory because Register hands the purchase service a
// stock hook, so the inventory service must already be in the container by the
// time procurement registers.
func (m *Module) Depends() []string {
	return []string{
		"base",
		"inventory",
	}
}

func (m *Module) Migrate(app *core.App) error {
	return migrations.Init(app.DB, &Vendor{}, &PurchaseOrder{}, &PurchaseOrderLine{})
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

	// Receiving a purchase order raises stock. The adapter lives in the wiring
	// rather than in either service: the purchase logic stays free of inventory
	// types, and inventory keeps not knowing that procurement exists.
	if stock, err := inventory.ServiceFromApp(app); err == nil {
		service.WithStockReceiver(func(ctx context.Context, tx *gorm.DB, tenantID uuid.UUID, lines []PurchaseOrderLine) error {
			receipts := make([]inventory.Receipt, 0, len(lines))
			for _, line := range lines {
				receipts = append(receipts, inventory.Receipt{
					SKU:         line.SKU,
					Description: line.Description,
					Quantity:    line.Quantity,
				})
			}
			return stock.ReceiveTx(ctx, tx, tenantID, receipts)
		})
	}

	slog.Info("module registered", "module", m.Name())
	return nil
}

func (m *Module) Boot(app *core.App) error {
	slog.Info("module booted", "module", m.Name())
	return nil
}

// ServiceFromApp resolves the procurement service from the container.
func ServiceFromApp(app *core.App) (*Service, error) {
	return core.ServiceAs[*Service](app, ServiceName)
}
