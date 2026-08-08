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
	RegisterRoutes(r Router)
}

type EventSubscriber interface {
	RegisterHooks(bus *EventBus)
}

type MenuProvider interface {
	Menu() []MenuItem
}

type Lifecycle interface {
	OnInstall(ctx context.Context) error
	OnUninstall(ctx context.Context) error
	OnUpgrade(ctx context.Context, fromVersion string) error
}

type Migration struct {
	ID   string
	Up   func() error
	Down func() error
}

type Router interface {
	Handle(method, path string, handler HandlerFunc)
}

type HandlerFunc func(ctx *Context) error
