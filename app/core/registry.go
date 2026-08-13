package core

import "fmt"

type Registry struct {
	modules map[string]Module
	order   []string
}

func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
	}
}

func (r *Registry) Register(module Module) error {
	name := module.Name()

	if _, exists := r.modules[name]; exists {
		return fmt.Errorf(
			"module %q already registered",
			name,
		)
	}

	if name == "" {
		return fmt.Errorf("module name must not be empty")
	}

	r.modules[name] = module
	r.order = append(r.order, name)

	return nil
}

func (r *Registry) Get(name string) (Module, bool) {
	module, ok := r.modules[name]

	return module, ok
}

func (r *Registry) All() []Module {
	modules := make([]Module, 0, len(r.modules))

	for _, name := range r.order {
		modules = append(modules, r.modules[name])
	}

	return modules
}

func (r *Registry) ResolveOrder() ([]Module, error) {
	result := make([]Module, 0)

	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(string) error

	visit = func(name string) error {
		if visited[name] {
			return nil
		}

		if visiting[name] {
			return fmt.Errorf(
				"circular dependency detected: %s",
				name,
			)
		}

		module, exists := r.modules[name]

		if !exists {
			return fmt.Errorf(
				"module %q not found",
				name,
			)
		}

		visiting[name] = true

		for _, dependency := range module.Depends() {
			if err := visit(dependency); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true

		result = append(result, module)

		return nil
	}

	for _, name := range r.order {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
