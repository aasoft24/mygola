// vendor/auth/auth.go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

type User interface {
	GetID() int
	GetEmail() string
	GetPassword() string
}

type Auth struct {
	store    UserStore
	sessions map[string]Session
}

type Session struct {
	UserID    int
	ExpiresAt time.Time
}

type UserStore interface {
	FindByID(id int) (User, error)
	FindByEmail(email string) (User, error)
}

func NewAuth(store UserStore) *Auth {
	return &Auth{
		store:    store,
		sessions: make(map[string]Session),
	}
}

func (a *Auth) Attempt(email, password string) (User, error) {
	user, err := a.store.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	// In real implementation, use bcrypt to compare hashed passwords
	if user.GetPassword() == password {
		return user, nil
	}

	return nil, ErrInvalidCredentials
}

func (a *Auth) Login(user User, w http.ResponseWriter) error {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return err
	}

	sessionToken := base64.StdEncoding.EncodeToString(token)
	a.sessions[sessionToken] = Session{
		UserID:    user.GetID(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	return nil
}

func (a *Auth) User(r *http.Request) (User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	session, exists := a.sessions[cookie.Value]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrUnauthenticated
	}

	return a.store.FindByID(session.UserID)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")
)
