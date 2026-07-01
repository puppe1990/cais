package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Store interface {
	Create(userID int64) (token string, err error)
	Get(token string) (userID int64, ok bool)
	Delete(token string)
}

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]int64)}
}

func (s *MemoryStore) Create(userID int64) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.sessions[token] = userID
	s.mu.Unlock()
	return token, nil
}

func (s *MemoryStore) Get(token string) (int64, bool) {
	s.mu.RLock()
	id, ok := s.sessions[token]
	s.mu.RUnlock()
	return id, ok
}

func (s *MemoryStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
