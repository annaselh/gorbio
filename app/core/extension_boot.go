package core

import (
	"context"
	"fmt"
)

func (r *ExtensionRegistry) RegisterAll(
	app *App,
) error {
	extensions, err := r.ResolveOrder(app.Modules)
	if err != nil {
		return err
	}

	for _, extension := range extensions {
		if err := extension.Register(app); err != nil {
			return fmt.Errorf("register extension %q: %w", extension.Name(), err)
		}
	}

	return nil
}

func (r *ExtensionRegistry) BootAll(
	app *App,
) error {
	extensions, err := r.ResolveOrder(app.Modules)
	if err != nil {
		return err
	}

	for _, extension := range extensions {
		if err := extension.Boot(app); err != nil {
			return fmt.Errorf("boot extension %q: %w", extension.Name(), err)
		}
	}

	return nil
}

// ShutdownAll calls optional extension shutdown hooks in reverse boot order.
func (r *ExtensionRegistry) ShutdownAll(ctx context.Context, app *App) error {
	extensions, err := r.ResolveOrder(app.Modules)
	if err != nil {
		return err
	}

	for i := len(extensions) - 1; i >= 0; i-- {
		extension, ok := extensions[i].(ExtensionShutdownable)
		if !ok {
			continue
		}

		if err := extension.Shutdown(ctx, app); err != nil {
			return fmt.Errorf("shutdown extension %q: %w", extensions[i].Name(), err)
		}
	}

	return nil
}
