package barcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultAPIBase = "https://world.openfoodfacts.org"

// Product is a barcode lookup result from an external catalog.
type Product struct {
	Name     string
	Category string
	Barcode  string
	Brand    string
	Quantity string
	ImageURL string
}

// Client queries Open Food Facts (or a compatible mock in tests).
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func (c *Client) Lookup(ctx context.Context, ean string) (Product, bool, error) {
	base := c.BaseURL
	if base == "" {
		base = DefaultAPIBase
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	url := fmt.Sprintf("%s/api/v2/product/%s.json", strings.TrimRight(base, "/"), ean)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Product{}, false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return Product{}, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Product{}, false, err
	}

	var payload struct {
		Status  int `json:"status"`
		Product struct {
			ProductName   string `json:"product_name"`
			Categories    string `json:"categories"`
			Brands        string `json:"brands"`
			Quantity      string `json:"quantity"`
			ImageFrontURL string `json:"image_front_url"`
			ImageURL      string `json:"image_url"`
		} `json:"product"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Product{}, false, err
	}
	if payload.Status != 1 || payload.Product.ProductName == "" {
		return Product{}, false, nil
	}

	imageURL := strings.TrimSpace(payload.Product.ImageFrontURL)
	if imageURL == "" {
		imageURL = strings.TrimSpace(payload.Product.ImageURL)
	}

	return Product{
		Name:     payload.Product.ProductName,
		Category: firstCategory(payload.Product.Categories),
		Barcode:  ean,
		Brand:    strings.TrimSpace(payload.Product.Brands),
		Quantity: strings.TrimSpace(payload.Product.Quantity),
		ImageURL: imageURL,
	}, true, nil
}

func firstCategory(categories string) string {
	for _, part := range strings.Split(categories, ",") {
		if s := strings.TrimSpace(part); s != "" {
			return s
		}
	}
	return "Geral"
}
