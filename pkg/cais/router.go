package cais

import (
	"net/http"
	"strconv"
)

type Middleware func(http.Handler) http.Handler

type IntHandler func(http.ResponseWriter, *http.Request, int64)
type StringHandler func(http.ResponseWriter, *http.Request, string)
type StringParamsHandler func(http.ResponseWriter, *http.Request, string, string)
type IntStringParamsHandler func(http.ResponseWriter, *http.Request, int64, string)

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
// Child routers share the parent ServeMux — only middleware differs, not URL namespace.
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
	r.StaticForEnv(prefix, dir, Config{})
}

// StaticForEnv serves files from disk. In development, sets no-store so JS/CSS edits apply without rebuild.
func (r *Router) StaticForEnv(prefix, dir string, cfg Config) {
	fs := http.FileServer(http.Dir(dir))
	handler := http.StripPrefix(prefix, fs)
	if cfg.Env == "development" {
		handler = noCacheStatic(handler)
	}
	r.mux.Handle("GET "+prefix+"/", handler)
}

func noCacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
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

// IntStringParams wraps a handler with an int64 and string path parameter.
func IntStringParams(intName, stringName string, fn IntStringParamsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue(intName), 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		s := r.PathValue(stringName)
		if s == "" {
			http.NotFound(w, r)
			return
		}
		fn(w, r, id, s)
	}
}

// StringParams wraps a handler with two string path parameters (avoids nested StringParam callbacks).
func StringParams(nameA, nameB string, fn StringParamsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := r.PathValue(nameA)
		b := r.PathValue(nameB)
		if a == "" || b == "" {
			http.NotFound(w, r)
			return
		}
		fn(w, r, a, b)
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
