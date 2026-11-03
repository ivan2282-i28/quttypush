
package models

import (
	"sync"
	"time"
)

// TokenRecord represents a registered device token. This is a pure Go
// (in-memory) representation that does not depend on any database driver.
type TokenRecord struct {
	Token     string    `json:"token"`
	Provider  string    `json:"provider"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is a concurrency-safe in-memory token store backed only by Go
// standard library primitives (no SQLite, no external driver).
type Store struct {
	mu      sync.RWMutex
	records map[string]*TokenRecord
}

// NewStore creates an empty in-memory store.
func NewStore() *Store {
	return &Store{records: make(map[string]*TokenRecord)}
}

// Create inserts (or replaces) a token record.
func (s *Store) Create(rec *TokenRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now
	s.records[rec.Token] = rec
}

// Delete removes a token record. Returns true if it existed.
func (s *Store) Delete(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[token]; !ok {
		return false
	}
	delete(s.records, token)
	return true
}

// Touch updates the UpdatedAt timestamp of a token. Returns true if found.
func (s *Store) Touch(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[token]
	if !ok {
		return false
	}
	rec.UpdatedAt = time.Now()
	return true
}

// Get returns a copy of the token record, if present.
func (s *Store) Get(token string) (TokenRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[token]
	if !ok {
		return TokenRecord{}, false
	}
	return *rec, true
}

// Count returns the number of stored tokens.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
