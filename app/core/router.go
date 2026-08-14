package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter() *Router {
	return &Router{
		engine: gin.Default(),
	}
}

// Handler exposes the router to an http.Server without coupling application
// lifecycle management to Gin's Run helper.
func (r *Router) Handler() http.Handler {
	return r.engine
}

// Use installs global middleware. Call it before modules register their routes
// so cross-cutting concerns such as CORS wrap every handler.
func (r *Router) Use(middleware ...gin.HandlerFunc) {
	r.engine.Use(middleware...)
}

func (r *Router) GET(
	path string,
	handlers ...gin.HandlerFunc,
) {
	r.engine.GET(path, handlers...)
}

func (r *Router) POST(
	path string,
	handlers ...gin.HandlerFunc,
) {
	r.engine.POST(path, handlers...)
}

func (r *Router) PUT(
	path string,
	handlers ...gin.HandlerFunc,
) {
	r.engine.PUT(path, handlers...)
}

func (r *Router) DELETE(
	path string,
	handlers ...gin.HandlerFunc,
) {
	r.engine.DELETE(path, handlers...)
}

func (r *Router) Group(
	path string,
) *gin.RouterGroup {
	return r.engine.Group(path)
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
