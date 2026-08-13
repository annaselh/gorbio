package core

import (
	"context"
	"fmt"
)

type LifecycleStage string

const (
	StageBeforeRegister LifecycleStage = "before_register"
	StageAfterRegister  LifecycleStage = "after_register"

	StageBeforeBoot LifecycleStage = "before_boot"
	StageAfterBoot  LifecycleStage = "after_boot"

	StageBeforeShutdown LifecycleStage = "before_shutdown"
	StageAfterShutdown  LifecycleStage = "after_shutdown"
)

func (app *App) Migrate() error {
	return app.Modules.MigrateAll(app)
}

func (app *App) Register() error {
	if err := app.Hooks.Run(
		StageBeforeRegister,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}

	if err := app.Modules.RegisterAll(app); err != nil {
		return fmt.Errorf(
			"register modules: %w",
			err,
		)
	}

	if err := app.Extensions.RegisterAll(app); err != nil {
		return fmt.Errorf("register extensions: %w", err)
	}

	if err := app.Hooks.Run(
		StageAfterRegister,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}

	return nil
}

// RegisterModules is retained as a compatibility alias. New code should call Register.
func (app *App) RegisterModules() error {
	return app.Register()
}

func (app *App) Boot() error {
	if err := app.Hooks.Run(
		StageBeforeBoot,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}

	if err := app.Modules.BootAll(app); err != nil {
		return fmt.Errorf(
			"boot modules: %w",
			err,
		)
	}

	if err := app.Extensions.BootAll(app); err != nil {
		return fmt.Errorf("boot extensions: %w", err)
	}

	if err := app.Hooks.Run(
		StageAfterBoot,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}

	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	if err := app.Hooks.Run(
		StageBeforeShutdown,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}

	if err := app.Extensions.ShutdownAll(ctx, app); err != nil {
		return fmt.Errorf("shutdown extensions: %w", err)
	}

	if err := app.Modules.ShutdownAll(ctx, app); err != nil {
		return fmt.Errorf("shutdown modules: %w", err)
	}

	if err := app.Hooks.Run(
		StageAfterShutdown,
		&HookContext{
			App: app,
		},
	); err != nil {
		return err
	}
	return nil
}
