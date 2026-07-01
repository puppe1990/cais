package cais

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntParam_ValidID(t *testing.T) {
	var got int64
	h := IntParam("id", func(w http.ResponseWriter, r *http.Request, id int64) {
		got = id
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/42/edit", nil)
	req.SetPathValue("id", "42")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d", rr.Code)
	}
	if got != 42 {
		t.Errorf("id = %d, want 42", got)
	}
}

func TestIntParam_InvalidID_Returns404(t *testing.T) {
	h := IntParam("id", func(w http.ResponseWriter, r *http.Request, id int64) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/items/x/edit", nil)
	req.SetPathValue("id", "x")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestStringParam_ExtractsSlug(t *testing.T) {
	var got string
	h := StringParam("slug", func(w http.ResponseWriter, r *http.Request, slug string) {
		got = slug
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/blog/hello", nil)
	req.SetPathValue("slug", "hello")
	rr := httptest.NewRecorder()
	h(rr, req)

	if got != "hello" {
		t.Errorf("slug = %q", got)
	}
}

func TestRouter_DeleteRoute(t *testing.T) {
	r := NewRouter()
	called := false
	r.Delete("/items/{id}", IntParam("id", func(w http.ResponseWriter, req *http.Request, id int64) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestRouter_GetRoute(t *testing.T) {
	r := NewRouter()
	called := false
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRouter_PostRoute(t *testing.T) {
	r := NewRouter()
	r.Post("/submit", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
}
