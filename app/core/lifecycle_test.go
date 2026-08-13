package core

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type testModule struct {
	name    string
	depends []string
	actions *[]string
}

func (m *testModule) Name() string      { return m.name }
func (m *testModule) Depends() []string { return m.depends }
func (m *testModule) Register(*App) error {
	*m.actions = append(*m.actions, "module:register")
	return nil
}
func (m *testModule) Boot(*App) error {
	*m.actions = append(*m.actions, "module:boot")
	return nil
}
func (m *testModule) Shutdown(context.Context, *App) error {
	*m.actions = append(*m.actions, "module:shutdown")
	return nil
}

type testExtension struct {
	name    string
	module  string
	depends []string
	actions *[]string
}

func (e *testExtension) Name() string      { return e.name }
func (e *testExtension) Module() string    { return e.module }
func (e *testExtension) Depends() []string { return e.depends }
func (e *testExtension) Register(*App) error {
	*e.actions = append(*e.actions, "extension:register")
	return nil
}
func (e *testExtension) Boot(*App) error {
	*e.actions = append(*e.actions, "extension:boot")
	return nil
}
func (e *testExtension) Shutdown(context.Context, *App) error {
	*e.actions = append(*e.actions, "extension:shutdown")
	return nil
}

func TestAppLifecycleOrder(t *testing.T) {
	var actions []string
	modules := NewRegistry()
	extensions := NewExtensionRegistry()

	if err := modules.Register(&testModule{name: "base", actions: &actions}); err != nil {
		t.Fatal(err)
	}
	if err := extensions.Register(&testExtension{name: "import", module: "base", actions: &actions}); err != nil {
		t.Fatal(err)
	}

	app := NewApp(nil, nil, modules, extensions)
	app.Hooks.Register(StageBeforeRegister, func(*HookContext) error {
		actions = append(actions, "hook:before-register")
		return nil
	})
	app.Hooks.Register(StageAfterRegister, func(*HookContext) error {
		actions = append(actions, "hook:after-register")
		return nil
	})
	app.Hooks.Register(StageBeforeBoot, func(*HookContext) error {
		actions = append(actions, "hook:before-boot")
		return nil
	})
	app.Hooks.Register(StageAfterBoot, func(*HookContext) error {
		actions = append(actions, "hook:after-boot")
		return nil
	})
	app.Hooks.Register(StageBeforeShutdown, func(*HookContext) error {
		actions = append(actions, "hook:before-shutdown")
		return nil
	})
	app.Hooks.Register(StageAfterShutdown, func(*HookContext) error {
		actions = append(actions, "hook:after-shutdown")
		return nil
	})

	if err := app.Register(); err != nil {
		t.Fatal(err)
	}
	if err := app.Boot(); err != nil {
		t.Fatal(err)
	}
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"hook:before-register", "module:register", "extension:register", "hook:after-register",
		"hook:before-boot", "module:boot", "extension:boot", "hook:after-boot",
		"hook:before-shutdown", "extension:shutdown", "module:shutdown", "hook:after-shutdown",
	}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("lifecycle order = %v, want %v", actions, want)
	}
}

func TestRegistryResolveOrderIsDeterministic(t *testing.T) {
	registry := NewRegistry()
	for _, module := range []*testModule{
		{name: "sales", depends: []string{"base"}},
		{name: "base"},
		{name: "inventory", depends: []string{"base"}},
	} {
		if err := registry.Register(module); err != nil {
			t.Fatal(err)
		}
	}

	for range 5 {
		modules, err := registry.ResolveOrder()
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, module := range modules {
			got = append(got, module.Name())
		}
		want := []string{"base", "sales", "inventory"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resolved order = %v, want %v", got, want)
		}
	}
}

func TestRegistryResolveOrderRejectsCycle(t *testing.T) {
	registry := NewRegistry()
	for _, module := range []*testModule{
		{name: "a", depends: []string{"b"}},
		{name: "b", depends: []string{"a"}},
	} {
		if err := registry.Register(module); err != nil {
			t.Fatal(err)
		}
	}

	_, err := registry.ResolveOrder()
	if err == nil || !strings.Contains(err.Error(), "circular dependency") {
		t.Fatalf("ResolveOrder error = %v, want circular dependency error", err)
	}
}
