package session

import (
	"crypto/rand"
	"io"
	"time"
)

type Session struct {
	id             string
	data           map[string]any
	createdAt      time.Time
	lastActivityAt time.Time
}

func generateSessionID() string {
	id := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic("somehow failed to generate session identifier")
	}

	return string(id)
}

func generateCSRFToken() string {
	token := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, token)
	if err != nil {
		panic("somehow failed to generate CSRF token")
	}

	return string(token)
}

func NewSession() *Session {
	return &Session{
		id: generateSessionID(),
		data: map[string]any{
			"authenticated": false,
			"csrf_token":    generateCSRFToken(),
		},
		createdAt:      time.Now(),
		lastActivityAt: time.Now(),
	}
}

func (s *Session) Get(key string) any {
	return s.data[key]
}

func (s *Session) Put(key string, value any) {
	s.data[key] = value
}

func (s *Session) Delete(key string) {
	delete(s.data, key)
}
