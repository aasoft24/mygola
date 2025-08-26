package web

import (
	"net/http"
)

type HandlerFunc func(*Ctx) error

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{routes: map[string]map[string]HandlerFunc{}}
}

func (r *Router) GET(path string, h HandlerFunc) {
	if r.routes["GET"] == nil {
		r.routes["GET"] = map[string]HandlerFunc{}
	}
	r.routes["GET"][path] = h
}

func (r *Router) POST(path string, h HandlerFunc) {
	if r.routes["POST"] == nil {
		r.routes["POST"] = map[string]HandlerFunc{}
	}
	r.routes["POST"][path] = h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if h, ok := r.routes[req.Method][req.URL.Path]; ok {
		c := &Ctx{W: w, R: req}
		_ = h(c)
		return
	}
	http.NotFound(w, req)
}
