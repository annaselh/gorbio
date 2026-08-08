package core

import "fmt"

type Registry struct {
	modules []Module
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Register(m Module) {
	r.modules = append(r.modules, m)
}

// ResolveOrder orders modules using topological sort.
func (r *Registry) ResolveOrder() ([]Module, error) {
	byName := map[string]Module{}
	for _, m := range r.modules {
		name := m.Manifest().Name
		if _, dup := byName[name]; dup {
			return nil, fmt.Errorf("duplicate module: %q", name)
		}
		byName[name] = m
	}

	var ordered []Module
	visited := map[string]bool{}
	inStack := map[string]bool{}

	var visit func(m Module) error
	visit = func(m Module) error {
		name := m.Manifest().Name
		if visited[name] {
			return nil
		}
		if inStack[name] {
			return fmt.Errorf("dependency cycle detected: %q", name)
		}
		inStack[name] = true
		for _, dep := range m.Manifest().Dependencies {
			depMod, ok := byName[dep]
			if !ok {
				return fmt.Errorf("dependency not found: %q needed by %q", dep, name)
			}
			if err := visit(depMod); err != nil {
				return err
			}
		}
		inStack[name] = false
		visited[name] = true
		ordered = append(ordered, m)
		return nil
	}

	for _, m := range r.modules {
		if err := visit(m); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
