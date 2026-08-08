package sales

import (
	"log"

	"github.com/annaselh/gorbio/core"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:         "sales",
		Version:      "1.0.0",
		Dependencies: []string{"base", "inventory"},
		Description:  "Sales",
	}
}

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/sales/orders", func(c *core.Context) error {
		return c.JSON(200, []map[string]any{
			{"id": 1, "partner_id": 1, "total": 1500000, "state": "draft"},
		})
	})
}

func (m *Module) RegisterHooks(bus *core.EventBus) {
	bus.Subscribe("base.partner.created", func(e core.Event) {
		log.Printf("sales listen new partner has been created: %+v", e.Payload)
	})
}
