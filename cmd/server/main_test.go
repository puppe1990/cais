package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_Smoke(t *testing.T) {
	t.Setenv("DB_PATH", ":memory:")

	a, err := bootstrap()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}