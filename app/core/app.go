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
	}
}

// WithSettings overrides the production-safe defaults. Call it during wiring,
// before Register, so modules observe the final values.
func (app *App) WithSettings(settings Settings) *App {
	app.Settings = settings
	return app
}
