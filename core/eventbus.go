package core

import (
	"sync"
)

type Event struct {
	Name    string
	Payload any
}

// EventHandler defines a function type that handles events.
type EventHandler func(e Event)

// EvnetBus: publish/subscribe for communication between modules
type EventBus struct {
	mu   sync.RWMutex
	subs map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{subs: map[string][]EventHandler{}}
}

// Subscribe resgiter handler for event name
func (b *EventBus) Subscribe(name string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[name] = append(b.subs[name], h)
}

// Publish sends an event to all registered subscribers.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	handlers := append([]EventHandler(nil), b.subs[e.Name]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		h(e)
	}
}
