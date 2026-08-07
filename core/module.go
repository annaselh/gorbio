package core

import "context"

type Module interface {
	Manifest() Manifest
}

type Manifest struct {
	Name         string
	Version      string
	Dependencies []string
	Author       string
	License      string
	Description  string
}

type Migratable interface {
	Migrations() []Migration
}

type Routable interface {
	RegisterRoutes(router Router)
}

type ModelProvider interface {
	RegisterModels(reg *ModelRegistry)
}

type EventSubscriber interface {
	RegisterHooks(bus *EventBus)
}

type MenuProvider interface {
	Menu() []MenuItem
}

type Lifecycle interface {
	OnInstall(ctx context.Context, tx DBTx) error
	OnUninstall(ctx context.Context, tx DBTx) error
	OUpgrade(ctx context.Context, tx DBTx, fromVersion string) error
}
