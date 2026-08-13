package inventory

import "github.com/annaselh/gorbio/core"

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

func (m *Module) Register(app *core.App) error {
	return nil
}

func (m *Module) Boot(app *core.App) error {
	return nil
}
