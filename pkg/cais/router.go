package cais

import (
	"net/http"
	"strconv"
)

type Middleware func(http.Handler) http.Handler

type IntHandler func(http.ResponseWriter, *http.Request, int64)
type StringHandler func(http.ResponseWriter, *http.Request, string)

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Group registers routes with extra middleware (e.g. admin auth).
func (r *Router) Group(mw Middleware, fn func(*Router)) {
	child := &Router{
		mux:         r.mux,
		middlewares: append(append([]Middleware{}, r.middlewares...), mw),
	}
	fn(child)
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.register("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.register("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.register("PUT", pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.register("PATCH", pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.register("DELETE", pattern, handler)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, r.wrap(handler))
}

func (r *Router) Static(prefix, dir string) {
	fs := http.FileServer(http.Dir(dir))
	r.mux.Handle("GET "+prefix+"/", http.StripPrefix(prefix, fs))
}

func (r *Router) register(method, pattern string, handler http.HandlerFunc) {
	r.mux.Handle(method+" "+pattern, r.wrap(handler))
}

func (r *Router) wrap(handler http.Handler) http.Handler {
	h := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// IntParam wraps a handler that receives a parsed int64 path parameter.
func IntParam(name string, fn IntHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		fn(w, r, id)
	}
}

// StringParam wraps a handler that receives a string path parameter.
func StringParam(name string, fn StringHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.PathValue(name)
		if v == "" {
			http.NotFound(w, r)
			return
		}
		fn(w, r, v)
	}
}
