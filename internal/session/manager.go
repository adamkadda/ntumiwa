package session

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"github.com/adamkadda/ntumiwa-site/internal/cookies"
	"github.com/adamkadda/ntumiwa-site/shared/logging"
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
	store SessionStore,
	gcInterval,
	absoluteExpiration,
	idleExpiration time.Duration,
	domain string,
	cookieName string,
	secretKey []byte,
) *SessionManager {

	m := &SessionManager{
		store:              store,
		absoluteExpiration: absoluteExpiration,
		idleExpiration:     idleExpiration,
		domain:             domain,
		cookieName:         cookieName,
		secretKey:          secretKey,
	}

	go m.gc(gcInterval)

	return m
}

func (m *SessionManager) valdiate(session *Session) bool {
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

	cookieValue, err := cookies.ReadSigned(r, m.cookieName, m.secretKey)
	if err == nil {
		session, err = m.store.read(cookieValue)
		if err != nil {
			// TODO: Access logger from context instead.
			log.Printf("Failed to read session from store: %v", err)
		}
	}

	if session == nil || !m.valdiate(session) {
		session = NewSession()
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
		return err
	}

	return nil
}

func (m *SessionManager) Migrate(session *Session) error {
	err := m.store.destroy(session.id)
	if err != nil {
		return err
	}

	session.id = generateSessionID()
	session.Put("csrf_token", generateCSRFToken())

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
		MaxAge:   int(m.idleExpiration / time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	err := cookies.WriteSigned(w, cookie, m.secretKey)
	if err != nil {
		// TODO: Access logger from context instead.
		log.Printf("Could not write signed cookie: %v", err)

		panic("Failed to write session cookie")
	}
}

func VerifyCSRFToken(session *Session, r *http.Request) bool {
	sessionCSRFToken, ok := session.Get("csrf_token").(string)
	if !ok {
		return false
	}

	requestCSRFToken := r.FormValue("csrf_token")

	if requestCSRFToken == "" {
		requestCSRFToken = r.Header.Get("X-XSRF-Token")
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

func Middleware(
	m *SessionManager,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logging.GetLogger(r)

			session, request := m.start(r)

			w.Header().Add("Vary", "Cookie")
			w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)

			if request.Method == http.MethodPost ||
				request.Method == http.MethodPut ||
				request.Method == http.MethodPatch ||
				request.Method == http.MethodDelete {

				if !VerifyCSRFToken(session, request) {
					l.Warn("CSRF Token mismatch")
					http.Error(w, "CSRF Token mismatch", http.StatusForbidden)
				}
			}

			next.ServeHTTP(w, r)

			m.save(session)

			ctx := r.Context()
			if !cookies.CookieWritten(ctx) {
				m.WriteCookie(w, r)
			}
		})
	}
}
