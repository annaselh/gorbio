package inventoryimport

import "github.com/annaselh/gorbio/core"

type Extension struct{}

func New() *Extension {
	return &Extension{}
}

func (e *Extension) Name() string {
	return "inventory-import"
}

func (e *Extension) Module() string {
	return "inventory"
}

func (e *Extension) Depends() []string {
	return []string{}
}

func (e *Extension) Register(app *core.App) error {
	return nil
}

func (e *Extension) Boot(app *core.App) error {
	return nil
}
