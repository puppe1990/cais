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

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("GET "+pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("POST "+pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("DELETE "+pattern, handler)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

func (r *Router) Static(prefix, dir string) {
	fs := http.FileServer(http.Dir(dir))
	r.mux.Handle("GET "+prefix+"/", http.StripPrefix(prefix, fs))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := http.Handler(r.mux)
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	handler.ServeHTTP(w, req)
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
