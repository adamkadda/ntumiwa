package session

import (
	"context"
	"crypto/subtle"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/adamkadda/ntumiwa/internal/config"
	"github.com/adamkadda/ntumiwa/internal/cookies"
	"github.com/adamkadda/ntumiwa/internal/logging"
)

type SessionManager struct {
	store              SessionStore
	absoluteExpiration time.Duration
	idleExpiration     time.Duration
	domain             string
	cookieName         string
	secretKey          []byte
}

func (m *SessionManager) gc(d time.Duration) {
	ticker := time.NewTicker(d)

	for range ticker.C {
		m.store.gc(m.idleExpiration, m.absoluteExpiration)
	}
}

func NewSessionManager(
	cfg config.SessionConfig,
	store SessionStore,
) *SessionManager {

	m := &SessionManager{
		store:              store,
		absoluteExpiration: cfg.AbsoluteExpiration,
		idleExpiration:     cfg.IdleExpiration,
		domain:             cfg.Domain,
		cookieName:         cfg.CookieName,
		secretKey:          cfg.SecretKey,
	}

	go m.gc(cfg.GCInterval)

	return m
}

func (m *SessionManager) validate(session *Session) bool {
	if time.Since(session.createdAt) > m.absoluteExpiration ||
		time.Since(session.lastActivityAt) > m.idleExpiration {

		// Upon discovering an expired session,
		// invoke the SessionStore's destroy method.
		err := m.store.destroy(session.id)
		if err != nil {
			panic(err)
		}

		return false
	}

	return true
}

type sessionContextKey struct{}

var sessionKey = sessionContextKey{}

func (m *SessionManager) start(r *http.Request) (*Session, *http.Request) {
	var session *Session
	var reason string

	cookieValue, err := cookies.ReadSigned(r, m.cookieName, m.secretKey)
	if err != nil {
		reason = err.Error()
	} else {
		session, err = m.store.read(cookieValue)
		if err != nil {
			reason = "failed to read session from store: " + err.Error()
			slog.Error("Failed to read session from store", slog.String("error", err.Error()))
		}
	}

	if session == nil || !m.validate(session) {
		session = NewSession()
		slog.Debug("Created new session", slog.String("reason", reason))
	}

	ctx := context.WithValue(r.Context(), sessionKey, session)
	r = r.WithContext(ctx)

	return session, r
}

func GetSession(r *http.Request) *Session {
	session, ok := r.Context().Value(sessionKey).(*Session)
	if !ok {
		panic("could not find session in context")
	}

	return session
}

func (m *SessionManager) save(session *Session) error {
	session.lastActivityAt = time.Now()

	err := m.store.write(session)
	if err != nil {
		slog.Debug("Failed session write", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("Session saved", slog.String("session_id", session.id))
	return nil
}

func (m *SessionManager) Migrate(session *Session) error {
	old := session.id

	err := m.store.destroy(session.id)
	if err != nil {
		return err
	}

	session.id = generateSessionID()
	session.Put("csrf_token", generateCSRFToken())

	new := session.id

	slog.Debug("Session migrated",
		slog.String("old", old),
		slog.String("new", new),
	)

	return nil
}

func (m *SessionManager) WriteCookie(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)

	cookie := http.Cookie{
		Name:     m.cookieName,
		Value:    session.id,
		Domain:   m.domain,
		Path:     "/",
		Expires:  time.Now().Add(m.idleExpiration),
		HttpOnly: true,
		Secure:   false, // WARN: Change to true for prod
		SameSite: http.SameSiteLaxMode,
	}

	err := cookies.WriteSigned(w, cookie, m.secretKey)
	if err != nil {
		// TODO: Access logger from context instead.
		log.Printf("Could not write signed cookie: %v", err)

		panic("Failed to write session cookie")
	}
}

func verifyCSRFToken(session *Session, r *http.Request) bool {
	sessionCSRFToken, ok := session.Get("csrf_token").(string)
	if !ok {
		return false
	}

	requestCSRFToken := r.FormValue("csrf_token")

	if requestCSRFToken == "" {
		requestCSRFToken = r.Header.Get("X-CSRF-Token")
	}

	if len(sessionCSRFToken) != len(requestCSRFToken) {
		return false
	}

	match := subtle.ConstantTimeCompare(
		[]byte(sessionCSRFToken),
		[]byte(requestCSRFToken),
	) == 1

	return match
}

func Middleware(m *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, r := m.start(r)

			if r.Method == http.MethodPost ||
				r.Method == http.MethodPut ||
				r.Method == http.MethodPatch ||
				r.Method == http.MethodDelete {

				if !verifyCSRFToken(session, r) {
					logging.GetLogger(r).Warn("Token mismatch")
					http.Error(w, "CSRF token mismatch", http.StatusForbidden)
					return
				}
			}

			sw := &sessionWriter{
				ResponseWriter: w,
				request:        r,
				manager:        m,
			}

			sw.Header().Add("Vary", "Cookie")
			sw.Header().Add("Cache-Control", "no-cache")
			sw.Header().Set("X-CSRF-Token", session.Get("csrf_token").(string))

			next.ServeHTTP(sw, r)

			m.save(session)

			if sw.cookieSet == false {
				m.WriteCookie(sw, r)
			}
		})
	}
}
