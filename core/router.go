package core

import (
	"net/http"
	"strings"
)

type route struct {
	method  string
	pattern []string
	handler HandlerFunc
}

type httpRouter struct {
	routes []route
	bus    *EventBus
}

func newRouter(bus *EventBus) *httpRouter {
	return &httpRouter{bus: bus}
}

func (r *httpRouter) Handle(method, path string, h HandlerFunc) {
	r.routes = append(r.routes, route{
		method:  method,
		pattern: splitPath(path),
		handler: h,
	})
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func (r *httpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	segs := splitPath(req.URL.Path)
	for _, rt := range r.routes {
		if rt.method != req.Method || len(rt.pattern) != len(segs) {
			continue
		}
		params := map[string]string{}
		matched := true
		for i, pat := range rt.pattern {
			if strings.HasPrefix(pat, ":") {
				params[pat[1:]] = segs[i]
			} else if pat != segs[i] {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		ctx := &Context{Writer: w, Request: req, bus: r.bus, params: params}
		if err := rt.handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, req)
}
