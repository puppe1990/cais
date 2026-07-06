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
	Display     string
	ThemeColor  string
	IconPath    string
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
		Display:     "fullscreen",
		ThemeColor:  ThemeColor,
	}
}

// HeadHTML returns meta tags and links to include in layout <head>.
func HeadHTML() string {
	return `<link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="Cais" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />`
}

// RegisterScript returns inline JS to register the service worker.
func RegisterScript() string {
	return RegisterScriptForEnv("production")
}

// RegisterScriptForEnv skips the service worker in development so static assets are not cached during hot reload.
func RegisterScriptForEnv(env string) string {
	if env == "development" {
		return `<script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          regs.forEach(function (r) { r.unregister(); });
        });
        if ("caches" in window) {
          caches.keys().then(function (keys) {
            keys.forEach(function (k) { caches.delete(k); });
          });
        }
      }
    </script>`
	}
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

	if err := os.MkdirAll(filepath.Join(staticDir, "img"), 0o755); err != nil {
		return err
	}

	for _, pair := range []struct{ src, dst string }{
		{"assets/sw.js", "js/sw.js"},
		{"assets/htmx.min.js", "js/htmx.min.js"},
		{"assets/idiomorph-ext.min.js", "js/idiomorph-ext.min.js"},
		{"assets/sse-ext.min.js", "js/sse-ext.min.js"},
		{"assets/cais-core.js", "js/cais-core.js"},
		{"assets/cais-chat.js", "js/cais-chat.js"},
		{"assets/cais-chat-logic.mjs", "js/cais-chat-logic.mjs"},
		{"assets/offline.html", "offline.html"},
		{"assets/icon.png", "icons/icon.png"},
		{"assets/go-on-cais.jpg", "img/go-on-cais.jpg"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png")); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath); err != nil {
		return err
	}

	return nil
}

// InstallTo writes PWA assets using DefaultConfig(name).
func InstallTo(appDir, name string) error {
	return WriteStatic(appDir, DefaultConfig(name))
}

// WriteStaticInertia writes PWA assets for Inertia+Svelte apps (no HTMX JS bundles).
func WriteStaticInertia(appDir string, cfg Config) error {
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
	if err := os.MkdirAll(filepath.Join(staticDir, "img"), 0o755); err != nil {
		return err
	}

	for _, pair := range []struct{ src, dst string }{
		{"assets/sw.js", "js/sw.js"},
		{"assets/offline.html", "offline.html"},
		{"assets/icon.png", "icons/icon.png"},
		{"assets/go-on-cais.jpg", "img/go-on-cais.jpg"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png")); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath); err != nil {
		return err
	}

	return nil
}

// InstallForInertia writes PWA assets for Inertia scaffolds (no HTMX).
func InstallForInertia(appDir, name string) error {
	return WriteStaticInertia(appDir, DefaultConfig(name))
}

func writeManifest(path string, cfg Config) error {
	display := cfg.Display
	if display == "" {
		display = "fullscreen"
	}
	type manifestData struct {
		Config
		Display string
	}
	const tpl = `{
  "name": {{printf "%q" .Name}},
  "short_name": {{printf "%q" .ShortName}},
  "description": {{printf "%q" .Description}},
  "start_url": {{printf "%q" .StartURL}},
  "display": {{printf "%q" .Display}},
  "background_color": "#f8fafc",
  "theme_color": {{printf "%q" .ThemeColor}},
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/static/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
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
	if err := t.Execute(&buf, manifestData{Config: cfg, Display: display}); err != nil {
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

func writeOGImage(path string) error {
	const width, height = 1200, 630
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(img, color.RGBA{R: 15, G: 23, B: 42, A: 255})
	accent := color.RGBA{R: 79, G: 70, B: 229, A: 255}
	barHeight := height / 5
	for y := 0; y < barHeight; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, accent)
		}
	}
	return encodePNG(path, img)
}

func fill(img *image.RGBA, c color.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func encodePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

func writeAppIcons(dir string, iconPath string) error {
	var data []byte
	var err error
	if iconPath != "" {
		data, err = os.ReadFile(iconPath)
	} else {
		data, err = assets.ReadFile("assets/icon.png")
	}
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), data, 0o644); err != nil {
		return err
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	for _, size := range []int{192, 512} {
		dst := resizeNearest(src, size, size)
		if err := encodePNG(filepath.Join(dir, fmt.Sprintf("icon-%d.png", size)), dst); err != nil {
			return err
		}
	}
	return nil
}

func resizeNearest(src image.Image, w, h int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sw, sh := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := bounds.Min.X + x*sw/w
			sy := bounds.Min.Y + y*sh/h
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// FS returns embedded PWA assets (for tests).
func FS() (fs.FS, error) {
	return fs.Sub(assets, "assets")
}
