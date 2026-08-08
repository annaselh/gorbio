package base

import (
	"context"
	"log"

	"github.com/annaselh/gorbio/core"
)

type Partner struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type store struct {
	partners map[int]Partner
	nextID   int
}

type Module struct {
	store *store
}

func New() *Module {
	return &Module{store: &store{partners: map[int]Partner{}, nextID: 1}}
}

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:        "base",
		Version:     "1.0.0",
		Description: "Users, company, partner",
	}
}

func (m *Module) Migrations() []core.Migration {
	return []core.Migration{
		{ID: "001_seed_partner", Up: func() error {
			m.store.partners[m.store.nextID] = Partner{
				ID: m.store.nextID, Name: "PT Example", Email: "jhon@example.com",
			}
			m.store.nextID++
			return nil
		}},
	}
}

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/base/partners", m.listPartners)
	r.Handle("GET", "/base/partners/:id", m.getPartner)
	r.Handle("POST", "/base/partners", m.createPartner)
}

func (m *Module) OnInstall(ctx context.Context) error {
	log.Println("base: OnInstall called")
	return nil
}

func (m *Module) OnUninstall(ctx context.Context) error { return nil }
