package base

import (
	"encoding/json"
	"strconv"

	"github.com/annaselh/gorbio/core"
)

func (m *Module) listPartners(c *core.Context) error {
	list := make([]Partner, 0, len(m.store.partners))
	for _, p := range m.store.partners {
		list = append(list, p)
	}
	return c.JSON(200, list)
}

func (m *Module) getPartner(c *core.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	p, ok := m.store.partners[id]
	if !ok {
		return c.JSON(404, map[string]string{"error": "partner not found"})
	}
	return c.JSON(200, p)
}

func (m *Module) createPartner(c *core.Context) error {
	var in Partner
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body is not valid"})
	}
	in.ID = m.store.nextID
	m.store.nextID++
	m.store.partners[in.ID] = in

	c.Bus().Publish(core.Event{Name: "base.partner.created", Payload: in})

	return c.JSON(201, in)
}
