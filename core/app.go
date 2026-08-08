package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type App struct {
	registry *Registry
	bus      *EventBus
	router   *httpRouter
}

func New() *App {
	bus := NewEventBus()
	return &App{
		registry: NewRegistry(),
		bus:      bus,
		router:   newRouter(bus),
	}
}

func (a *App) Register(m Module) { a.registry.Register(m) }

func (a *App) Boot() error {
	ordered, err := a.registry.ResolveOrder()
	if err != nil {
		return err
	}
	for _, m := range ordered {
		name := m.Manifest().Name

		if mig, ok := m.(Migratable); ok {
			for _, mg := range mig.Migrations() {
				if err := mg.Up(); err != nil {
					return fmt.Errorf("migration %s on module %s failed: %w", mg.ID, name, err)
				}
			}
		}
		if rt, ok := m.(Routable); ok {
			rt.RegisterRoutes(a.router)
		}
		if es, ok := m.(EventSubscriber); ok {
			es.RegisterHooks(a.bus)
		}
		if lc, ok := m.(Lifecycle); ok {
			if err := lc.OnInstall(context.Background()); err != nil {
				return fmt.Errorf("OnInstall module %s failed: %w", name, err)
			}
		}
		log.Printf("module in-boot: %s", name)
	}
	return nil
}

func (a *App) Serve(addr string) error {
	log.Printf("server running on %s", addr)
	return http.ListenAndServe(addr, a.router)
}

func (a *App) Bus() *EventBus { return a.bus }
