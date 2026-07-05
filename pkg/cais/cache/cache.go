package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type entry[V any] struct {
	val       V
	expiresAt time.Time
}

// Cache is a thread-safe in-memory key-value store with per-entry TTL.
type Cache[V any] struct {
	mu   sync.RWMutex
	ttl  time.Duration
	data map[string]entry[V]
}

// New returns a cache where each Set expires after ttl.
func New[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{
		ttl:  ttl,
		data: make(map[string]entry[V]),
	}
}

// Get returns the value for key when present and not expired.
func (c *Cache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.val, true
}

// Set stores val under key; it expires after the cache TTL.
func (c *Cache[V]) Set(key string, val V) {
	c.mu.Lock()
	c.data[key] = entry[V]{
		val:       val,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Delete removes key from the cache.
func (c *Cache[V]) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Key builds a deterministic cache key by joining the parts with "|".
// Recommended for list pages: combine stable identifiers + a Hash() of a version struct
// (e.g. count + max updated time) instead of serializing the entire list into the key.
func Key(parts ...any) string {
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteByte('|')
		}
		fmt.Fprintf(&b, "%v", p)
	}
	return b.String()
}

// Hash returns a short (16 hex chars) stable hash of v.
// Use it to create cache keys / ETags that change only when the important data changes.
func Hash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(fmt.Sprintf("%#v", v))
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
