package core

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Identity struct {
	UserID        uuid.UUID
	Email         string
	Roles         []string
	Authenticated bool
}

func (id Identity) HasRole(role string) bool {
	for _, r := range id.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type Context struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	bus      *EventBus
	params   map[string]string
	identity Identity
}

// JSON writes JSON response with specified status code and value.
func (c *Context) JSON(status int, v any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(v)
}

func (c *Context) Bus() *EventBus          { return c.bus }
func (c *Context) Param(key string) string { return c.params[key] }

func (c *Context) Identity() Identity       { return c.identity }
func (c *Context) SetIdentity(id Identity)  { c.identity = id }
func (c *Context) Header(key string) string { return c.Request.Header.Get(key) }
func (c *Context) IsAuthenticated() bool    { return c.identity.Authenticated }

type MenuItem struct {
	Name     string
	Icon     string
	Children []MenuItem
}
