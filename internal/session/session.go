package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"
	"time"
)

type Session struct {
	id             string
	data           map[string]any
	createdAt      time.Time
	lastActivityAt time.Time
}

// NOTE: Double encoding happens, once here and
// again during cookie reading/writing.
// I do this for better debugging.
func generateSessionID() string {
	id := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic("somehow failed to generate session identifier")
	}

	sessionID := base64.URLEncoding.EncodeToString(id)

	// TODO: Revert back to Debug level
	slog.Debug("Generated session id", slog.String("session_id", sessionID))

	return sessionID
}

func generateCSRFToken() string {
	token := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, token)
	if err != nil {
		panic("somehow failed to generate CSRF token")
	}

	encoded := base64.RawURLEncoding.EncodeToString(token)

	// TODO: Revert back to Debug level
	slog.Debug("Generated CSRF token", slog.String("token", encoded))

	return encoded
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
