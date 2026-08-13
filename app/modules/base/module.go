package base

import "github.com/annaselh/gorbio/core"

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "base"
}

func (m *Module) Depends() []string {
	return nil
}

func (m *Module) Migrate(app *core.App) error {
	return Migrate(app.DB)
}

func (m *Module) Register(app *core.App) error {
	auth := NewAuthService(app.DB)
	if err := app.Services.Register(AuthServiceName, auth); err != nil {
		return err
	}

	app.Router.POST("/api/auth/login", auth.loginHandler)
	app.Router.POST("/api/auth/logout", auth.RequireAuth(), auth.logoutHandler)
	app.Router.GET("/api/auth/me", auth.RequireAuth(), auth.meHandler)
	return nil
}

func (m *Module) Boot(app *core.App) error {
	return nil
}
