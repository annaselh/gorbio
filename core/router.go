package core

type Router interface {
	GET(path string, handler HandlerFunc)
	POST(path string, handler HandlerFunc)
	PUT(path string, handler HandlerFunc)
	DELETE(path string, handler HandlerFunc)
	GROUP(prefix string) Router
}

type HandlerFunc func(ctx Context) error
