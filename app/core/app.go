package core

import "gorm.io/gorm"

type App struct {
	DB         *gorm.DB
	Router     *Router
	Modules    *Registry
	Extensions *ExtensionRegistry
	Services   *ServiceRegistry
	Hooks      *HookManager
	Events     *EventBus
	Settings   Settings
	Mailer     Mailer
}

func NewApp(
	db *gorm.DB,
	router *Router,
	registry *Registry,
	extensionRegistry *ExtensionRegistry,
) *App {
	return &App{
		DB:         db,
		Router:     router,
		Modules:    registry,
		Extensions: extensionRegistry,
		Services:   NewServiceRegistry(),
		Hooks:      NewHookManager(),
		Events:     NewEventBus(),
		Settings:   DefaultSettings(),
		Mailer:     LogMailer{},
	}
}

// WithSettings overrides the production-safe defaults. Call it during wiring,
// before Register, so modules observe the final values.
func (app *App) WithSettings(settings Settings) *App {
	app.Settings = settings
	return app
}

// WithMailer installs the outbound mail transport. The default is LogMailer,
// which delivers nothing; wiring must replace it for any real deployment.
func (app *App) WithMailer(mailer Mailer) *App {
	if mailer != nil {
		app.Mailer = mailer
	}
	return app
}
