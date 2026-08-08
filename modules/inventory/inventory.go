package inventory

import "github.com/annaselh/gorbio/core"

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:         "inventory",
		Version:      "1.0.0",
		Dependencies: []string{"base"},
		Description:  "Product & stock",
	}
}

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/inventory/products", func(c *core.Context) error {
		return c.JSON(200, []map[string]any{
			{"id": 1, "sku": "SKU-001", "name": "Product Example"},
		})
	})
}
