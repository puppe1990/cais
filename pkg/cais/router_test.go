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

func TestStringParams_extractsTwoParams(t *testing.T) {
	var gotA, gotB string
	h := StringParams("id", "permID", func(w http.ResponseWriter, r *http.Request, id, permID string) {
		gotA, gotB = id, permID
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/chat/abc/permissions/xyz/approve", nil)
	req.SetPathValue("id", "abc")
	req.SetPathValue("permID", "xyz")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d", rr.Code)
	}
	if gotA != "abc" || gotB != "xyz" {
		t.Errorf("params = (%q, %q), want (abc, xyz)", gotA, gotB)
	}
}

func TestStringParams_missingParam_Returns404(t *testing.T) {
	h := StringParams("id", "permID", func(w http.ResponseWriter, r *http.Request, id, permID string) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodPost, "/chat/abc/permissions//approve", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
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

func TestRouter_PutAndPatch(t *testing.T) {
	r := NewRouter()
	putCalled := false
	patchCalled := false

	r.Put("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		putCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	r.Patch("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		patchCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	putReq := httptest.NewRequest(http.MethodPut, "/items/1", nil)
	putReq.SetPathValue("id", "1")
	putRR := httptest.NewRecorder()
	r.ServeHTTP(putRR, putReq)

	if !putCalled {
		t.Error("PUT handler not called")
	}
	if putRR.Code != http.StatusNoContent {
		t.Errorf("PUT status = %d, want %d", putRR.Code, http.StatusNoContent)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/items/1", nil)
	patchReq.SetPathValue("id", "1")
	patchRR := httptest.NewRecorder()
	r.ServeHTTP(patchRR, patchReq)

	if !patchCalled {
		t.Error("PATCH handler not called")
	}
	if patchRR.Code != http.StatusNoContent {
		t.Errorf("PATCH status = %d, want %d", patchRR.Code, http.StatusNoContent)
	}
}

func TestRouter_Group_AppliesMiddleware(t *testing.T) {
	r := NewRouter()
	publicCalled := false
	adminCalled := false

	r.Get("/public", func(w http.ResponseWriter, req *http.Request) {
		publicCalled = true
		w.WriteHeader(http.StatusOK)
	})

	block := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "blocked", http.StatusForbidden)
		})
	}

	r.Group(block, func(g *Router) {
		g.Get("/admin", func(w http.ResponseWriter, req *http.Request) {
			adminCalled = true
		})
	})

	pub := httptest.NewRequest(http.MethodGet, "/public", nil)
	pubRR := httptest.NewRecorder()
	r.ServeHTTP(pubRR, pub)

	if !publicCalled {
		t.Error("public handler not called")
	}
	if pubRR.Code != http.StatusOK {
		t.Errorf("public status = %d", pubRR.Code)
	}

	adm := httptest.NewRequest(http.MethodGet, "/admin", nil)
	admRR := httptest.NewRecorder()
	r.ServeHTTP(admRR, adm)

	if adminCalled {
		t.Error("admin handler should not run behind block middleware")
	}
	if admRR.Code != http.StatusForbidden {
		t.Errorf("admin status = %d, want 403", admRR.Code)
	}
}

func TestRouter_Use_AppliesGlobally(t *testing.T) {
	r := NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Test", "1")
			next.ServeHTTP(w, req)
		})
	})
	r.Get("/ok", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ok", nil))

	if rr.Header().Get("X-Test") != "1" {
		t.Error("global middleware not applied")
	}
}
