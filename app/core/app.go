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
	}
}
