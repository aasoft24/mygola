// pkg/session/session.go
package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"
)

type Session interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
	Save() error
	ID() string
}

type Store interface {
	Get(sessionID string) (map[string]interface{}, error)
	Save(sessionID string, data map[string]interface{}, expiration time.Duration) error
	Delete(sessionID string) error
}

type Manager struct {
	store      Store
	cookieName string
	mu         sync.Mutex
}

func NewManager(store Store, cookieName string) *Manager {
	return &Manager{
		store:      store,
		cookieName: cookieName,
	}
}

func (m *Manager) Start(w http.ResponseWriter, r *http.Request) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get session ID from cookie
	var sessionID string
	cookie, err := r.Cookie(m.cookieName)
	if errors.Is(err, http.ErrNoCookie) {
		// Create new session
		sessionID = generateSessionID()
	} else if err != nil {
		return nil, err
	} else {
		sessionID = cookie.Value
	}

	// Get session data from store
	data, err := m.store.Get(sessionID)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &session{
		id:      sessionID,
		data:    data,
		store:   m.store,
		written: false,
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return session, nil
}

type session struct {
	id      string
	data    map[string]interface{}
	store   Store
	written bool
	mu      sync.Mutex
}

func (s *session) Get(key string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key]
}

func (s *session) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.written = true
}

func (s *session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.written = true
}

func (s *session) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written {
		err := s.store.Save(s.id, s.data, 24*time.Hour)
		if err != nil {
			return err
		}
		s.written = false
	}
	return nil
}

func (s *session) ID() string {
	return s.id
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Memory store implementation
type MemoryStore struct {
	sessions map[string]memorySession
	mu       sync.RWMutex
}

type memorySession struct {
	data       map[string]interface{}
	expiration time.Time
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		sessions: make(map[string]memorySession),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

func (s *MemoryStore) Get(sessionID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists || time.Now().After(session.expiration) {
		return make(map[string]interface{}), nil
	}

	// Return a copy
	data := make(map[string]interface{})
	for k, v := range session.data {
		data[k] = v
	}

	return data, nil
}

func (s *MemoryStore) Save(sessionID string, data map[string]interface{}, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = memorySession{
		data:       data,
		expiration: time.Now().Add(expiration),
	}

	return nil
}

func (s *MemoryStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.expiration) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}
