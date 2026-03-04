package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Session struct {
	ID     string
	UserID int
	Role   string
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]Session{}}
}

func (s *SessionStore) Create(userID int, role string) string {
	id := newSessionID()
	s.mu.Lock()
	s.sessions[id] = Session{ID: id, UserID: userID, Role: role}
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(id string) (Session, bool) {
	s.mu.RLock()
	v, ok := s.sessions[id]
	s.mu.RUnlock()
	return v, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func newSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
