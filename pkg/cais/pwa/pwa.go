package pwa

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed assets/*
var assets embed.FS

const ThemeColor = "#4f46e5"

type Config struct {
	Name        string
	ShortName   string
	Description string
	StartURL    string
	ThemeColor  string
}

func DefaultConfig(name string) Config {
	short := name
	if len(short) > 12 {
		short = short[:12]
	}
	return Config{
		Name:        name,
		ShortName:   short,
		Description: name + " — powered by Cais",
		StartURL:    "/",
		ThemeColor:  ThemeColor,
	}
}

// HeadHTML returns meta tags and links to include in layout <head>.
func HeadHTML() string {
	return `<link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="default" />
    <meta name="apple-mobile-web-app-title" content="Cais" />
    <link rel="apple-touch-icon" href="/static/icons/icon-192.png" />
    <link rel="icon" href="/static/icons/icon.svg" type="image/svg+xml" />`
}

// RegisterScript returns inline JS to register the service worker.
func RegisterScript() string {
	return `<script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>`
}

// WriteStatic writes default PWA assets into web/static for an app.
func WriteStatic(appDir string, cfg Config) error {
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = ThemeColor
	}
	if cfg.StartURL == "" {
		cfg.StartURL = "/"
	}

	staticDir := filepath.Join(appDir, "web", "static")
	if err := os.MkdirAll(filepath.Join(staticDir, "icons"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(staticDir, "js"), 0o755); err != nil {
		return err
	}

	if err := writeManifest(filepath.Join(staticDir, "manifest.webmanifest"), cfg); err != nil {
		return err
	}

	for _, pair := range []struct{ src, dst string }{
		{"assets/sw.js", "js/sw.js"},
		{"assets/htmx.min.js", "js/htmx.min.js"},
		{"assets/offline.html", "offline.html"},
		{"assets/icon.svg", "icons/icon.svg"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}

	for _, size := range []int{192, 512} {
		if err := writePNGIcon(filepath.Join(staticDir, "icons", fmt.Sprintf("icon-%d.png", size)), size); err != nil {
			return err
		}
	}

	return nil
}

// InstallTo writes PWA assets using DefaultConfig(name).
func InstallTo(appDir, name string) error {
	return WriteStatic(appDir, DefaultConfig(name))
}

func writeManifest(path string, cfg Config) error {
	const tpl = `{
  "name": {{printf "%q" .Name}},
  "short_name": {{printf "%q" .ShortName}},
  "description": {{printf "%q" .Description}},
  "start_url": {{printf "%q" .StartURL}},
  "display": "standalone",
  "background_color": "#f8fafc",
  "theme_color": {{printf "%q" .ThemeColor}},
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/static/icons/icon.svg",
      "sizes": "any",
      "type": "image/svg+xml",
      "purpose": "any"
    },
    {
      "src": "/static/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/static/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any maskable"
    }
  ]
}
`
	t, err := template.New("manifest").Parse(tpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func copyAsset(src, dst string) error {
	data, err := assets.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func writePNGIcon(path string, size int) error {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	indigo := color.RGBA{R: 79, G: 70, B: 229, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, indigo)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// FS returns embedded PWA assets (for tests).
func FS() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
