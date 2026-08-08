package core

import "testing"

type fakeModule struct {
	name string
	deps []string
}

func (f fakeModule) Manifest() Manifest {
	return Manifest{Name: f.name, Dependencies: f.deps}
}

func TestResolveOrder(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "sales", deps: []string{"base", "inventory"}})
	r.Register(fakeModule{name: "base"})
	r.Register(fakeModule{name: "inventory", deps: []string{"base"}})

	ordered, err := r.ResolveOrder()

	if err != nil {
		t.Fatalf("unknown error: %v", err)
	}
	pos := map[string]int{}
	for i, m := range ordered {
		pos[m.Manifest().Name] = i
	}

	if pos["base"] > pos["inventory"] {
		t.Errorf("base must be before inventory")
	}
	if pos["inventory"] > pos["sales"] {
		t.Errorf("inventory must be before sales")
	}
}

func TestMissingDependency(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "sales", deps: []string{"not found"}})
	if _, err := r.ResolveOrder(); err == nil {
		t.Fatal(" must be error because missing dependency")
	}
}

func TestCyclicDependency(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "a", deps: []string{"b"}})
	r.Register(fakeModule{name: "b", deps: []string{"a"}})
	if _, err := r.ResolveOrder(); err == nil {
		t.Fatal("must be error because cyclic")
	}
}
