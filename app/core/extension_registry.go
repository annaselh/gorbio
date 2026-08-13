package core

import "fmt"

type ExtensionRegistry struct {
	extensions map[string]Extension
	order      []string
}

func NewExtensionRegistry() *ExtensionRegistry {
	return &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}
}

func (r *ExtensionRegistry) Register(
	extension Extension,
) error {
	name := extension.Name()
	if name == "" {
		return fmt.Errorf("extension name must not be empty")
	}

	if _, exists := r.extensions[name]; exists {
		return fmt.Errorf(
			"extension %q already registered",
			name,
		)
	}

	r.extensions[name] = extension
	r.order = append(r.order, name)

	return nil
}

func (r *ExtensionRegistry) All() []Extension {
	result := make([]Extension, 0, len(r.extensions))

	for _, name := range r.order {
		result = append(result, r.extensions[name])
	}

	return result
}

func (r *ExtensionRegistry) Get(name string) (Extension, bool) {
	extension, ok := r.extensions[name]
	return extension, ok
}

// ResolveOrder validates module ownership and returns extensions after their
// extension dependencies in a deterministic registration order.
func (r *ExtensionRegistry) ResolveOrder(modules *Registry) ([]Extension, error) {
	result := make([]Extension, 0, len(r.extensions))
	visited := make(map[string]bool, len(r.extensions))
	visiting := make(map[string]bool, len(r.extensions))

	var visit func(string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("circular extension dependency detected at %q", name)
		}

		extension, exists := r.extensions[name]
		if !exists {
			return fmt.Errorf("extension %q not found", name)
		}

		if moduleName := extension.Module(); moduleName != "" {
			if _, exists := modules.Get(moduleName); !exists {
				return fmt.Errorf("extension %q requires module %q", name, moduleName)
			}
		}

		visiting[name] = true
		for _, dependency := range extension.Depends() {
			if err := visit(dependency); err != nil {
				return fmt.Errorf("resolve extension %q: %w", name, err)
			}
		}
		visiting[name] = false
		visited[name] = true
		result = append(result, extension)
		return nil
	}

	for _, name := range r.order {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
