package core

import (
	"fmt"
	"sync"
)

// ServiceRegistry is a small application container for services shared by
// modules. Services are registered during module registration and read-only at
// request time.
type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]any
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{services: make(map[string]any)}
}

func (r *ServiceRegistry) Register(name string, service any) error {
	if name == "" || service == nil {
		return fmt.Errorf("service name and value are required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %q already registered", name)
	}
	r.services[name] = service
	return nil
}

func (r *ServiceRegistry) Get(name string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	service, ok := r.services[name]
	return service, ok
}
