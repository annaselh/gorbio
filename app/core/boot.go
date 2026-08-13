package core

import (
	"context"
	"fmt"
)

func (r *Registry) MigrateAll(app *App) error {
	modules, err := r.ResolveOrder()
	if err != nil {
		return err
	}
	for _, module := range modules {
		migrator, ok := module.(Migrator)
		if !ok {
			continue
		}
		if err := migrator.Migrate(app); err != nil {
			return fmt.Errorf("migrate module %q: %w", module.Name(), err)
		}
	}
	return nil
}

func (r *Registry) RegisterAll(app *App) error {
	modules, err := r.ResolveOrder()

	if err != nil {
		return err
	}

	for _, module := range modules {

		if err := module.Register(app); err != nil {
			return fmt.Errorf(
				"register module %q: %w",
				module.Name(),
				err,
			)
		}
	}

	return nil
}

func (r *Registry) BootAll(app *App) error {
	modules, err := r.ResolveOrder()

	if err != nil {
		return err
	}

	for _, module := range modules {

		if err := module.Boot(app); err != nil {
			return fmt.Errorf(
				"boot module %q: %w",
				module.Name(),
				err,
			)
		}
	}

	return nil
}

// ShutdownAll calls optional module shutdown hooks in reverse boot order.
func (r *Registry) ShutdownAll(ctx context.Context, app *App) error {
	modules, err := r.ResolveOrder()
	if err != nil {
		return err
	}

	for i := len(modules) - 1; i >= 0; i-- {
		module, ok := modules[i].(Shutdownable)
		if !ok {
			continue
		}

		if err := module.Shutdown(ctx, app); err != nil {
			return fmt.Errorf("shutdown module %q: %w", modules[i].Name(), err)
		}
	}

	return nil
}
