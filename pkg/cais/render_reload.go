package cais

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/puppe1990/cais/pkg/cais/i18n"
)

var (
	templateWatcherMu   sync.Mutex
	templateWatcherStop func()
)

// NewRendererForEnv loads templates from disk with hot reload in development,
// otherwise from the embedded filesystem (production builds).
func NewRendererForEnv(cfg Config, embedded fs.FS, templatesDir string, catalog *i18n.Catalog) (*Renderer, error) {
	if cfg.Env != "production" && templatesDir != "" {
		if _, err := os.Stat(templatesDir); err == nil {
			r, err := NewRendererFromDir(templatesDir, catalog)
			if err == nil {
				StartTemplateWatcher(templatesDir, r)
				return r, nil
			}
			if isNoLayoutTemplatesErr(err) {
				return NewRendererStub(catalog), nil
			}
			return nil, err
		}
	}
	if embedded == nil {
		return nil, fmt.Errorf("embedded template filesystem is required when not in development")
	}
	r, err := NewRenderer(embedded, catalog)
	if err != nil {
		if isNoLayoutTemplatesErr(err) {
			return NewRendererStub(catalog), nil
		}
		return nil, err
	}
	return r, nil
}

func isNoLayoutTemplatesErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no layout templates found")
}

// ReloadFromDir re-parses templates from a directory (development hot reload).
func (r *Renderer) ReloadFromDir(dir string) error {
	return r.reloadFromFS(os.DirFS(dir))
}

func (r *Renderer) reloadFromFS(fsys fs.FS) error {
	next, err := NewRenderer(fsys, r.catalog)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.pages = next.pages
	r.partials = next.partials
	r.mu.Unlock()
	return nil
}

// StartTemplateWatcher polls template files and reloads the renderer when they change.
func StartTemplateWatcher(dir string, r *Renderer) {
	if dir == "" || r == nil {
		return
	}
	stopTemplateWatcher()

	done := make(chan struct{})
	templateWatcherMu.Lock()
	templateWatcherStop = func() { close(done) }
	templateWatcherMu.Unlock()

	go func() {
		var lastMod int64
		tick := time.NewTicker(400 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				mod, err := templateTreeModTime(dir)
				if err != nil || mod <= lastMod {
					continue
				}
				lastMod = mod
				if err := r.ReloadFromDir(dir); err != nil {
					log.Printf("cais: template reload failed: %v", err)
					continue
				}
				log.Printf("cais: templates reloaded")
			}
		}
	}()
}

func stopTemplateWatcher() {
	templateWatcherMu.Lock()
	stop := templateWatcherStop
	templateWatcherStop = nil
	templateWatcherMu.Unlock()
	if stop != nil {
		stop()
	}
}

func templateTreeModTime(dir string) (int64, error) {
	var latest int64
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if m := info.ModTime().UnixNano(); m > latest {
			latest = m
		}
		return nil
	})
	return latest, err
}
