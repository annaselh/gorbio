package core

import (
	"fmt"
	"sync"
)

type HookContext struct {
	App   *App
	Value any
}

type HookHandler func(*HookContext) error

type HookManager struct {
	mu    sync.RWMutex
	hooks map[LifecycleStage][]HookHandler
}

func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[LifecycleStage][]HookHandler),
	}
}

func (h *HookManager) Register(
	stage LifecycleStage,
	handler HookHandler,
) {
	if handler == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.hooks[stage] = append(h.hooks[stage], handler)
}

func (h *HookManager) Run(
	stage LifecycleStage,
	ctx *HookContext,
) error {
	h.mu.RLock()
	handlers := append([]HookHandler(nil), h.hooks[stage]...)
	h.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx); err != nil {
			return fmt.Errorf(
				"hook %s failed: %w",
				stage,
				err,
			)
		}
	}
	return nil
}
