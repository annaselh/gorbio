package core

import "sync"

type EventHandler func(event any)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(
			map[string][]EventHandler,
		),
	}
}

func (b *EventBus) Subscribe(
	event string,
	handler EventHandler,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[event] = append(
		b.handlers[event],
		handler,
	)
}

func (b *EventBus) Publish(
	event string,
	payload any,
) {
	b.mu.RLock()
	handlers := append(
		[]EventHandler(nil),
		b.handlers[event]...,
	)
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(payload)
	}
}
