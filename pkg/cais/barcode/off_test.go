package barcode

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_LookupProduct_found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/product/7894900011517.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": 1,
			"product": {
				"product_name": "Coca-Cola Original 2L",
				"categories": "Bebidas, Refrigerantes"
			}
		}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	prod, ok, err := c.Lookup(context.Background(), "7894900011517")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if !ok {
		t.Fatal("Lookup() ok = false, want true")
	}
	if prod.Name != "Coca-Cola Original 2L" {
		t.Errorf("Name = %q", prod.Name)
	}
	if prod.Category != "Bebidas" {
		t.Errorf("Category = %q, want Bebidas", prod.Category)
	}
	if prod.Barcode != "7894900011517" {
		t.Errorf("Barcode = %q", prod.Barcode)
	}
}

func TestClient_LookupProduct_notFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": 0}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	_, ok, err := c.Lookup(context.Background(), "0000000000000")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if ok {
		t.Fatal("Lookup() ok = true, want false")
	}
}
