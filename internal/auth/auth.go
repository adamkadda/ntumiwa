package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/internal/hash"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/session"
)

var (
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrPasswordTooShort   = errors.New("password does not meet minimum length")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInternalError      = errors.New("internal error")
)

/*
	IsAuthenticated returns true if the session is authenticated.

	Panics if type mismatch, because the session's "authenticated"
	key must ALWAYS return a bool.
*/

func IsAuthenticated(r *http.Request) bool {
	s := session.GetSession(r)

	result, ok := s.Get("authenticated").(bool)
	if !ok {
		panic("type assertion failed: authenticated value not bool")
	}

	return result
}

func Register(db *sql.DB, username string, password string) error {

	// check if user exists in db
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)",
		username,
	).Scan(&exists)
	if exists {
		return ErrUsernameTaken
	} else if err != nil {
		return err
	}

	// generate hash from password
	hashedPassword, err := hash.GenerateFromPassword(password)
	if err != nil {
		return err
	}

	// insert into db
	_, err = db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		username,
		hashedPassword,
	)
	if err != nil {
		return err
	}

	return nil
}

// Invoke only after authentication
func Login(
	m *session.SessionManager,
	r *http.Request,
	username string,
) error {
	session := session.GetSession(r)

	// Change of permissions -> migrate session
	err := m.Migrate(session)
	if err != nil {
		return fmt.Errorf("failed to migrate session: %v", err)
	}

	session.Put("authenticated", true)
	session.Put("username", username)

	return nil
}

func Logout(
	m *session.SessionManager,
	r *http.Request,
) error {
	session := session.GetSession(r)

	// Change of permissions -> migrate session
	err := m.Migrate(session)
	if err != nil {
		return fmt.Errorf("failed to migrate session: %v", err)
	}

	session.Put("authenticated", false)
	session.Delete("username")

	return nil
}

func VerifyCredentials(db *sql.DB, username string, password string) error {
	var hashedPassword string

	exists := true

	err := db.QueryRow(
		"SELECT password FROM users WHERE username = ?",
		username,
	).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		exists = false
		hashedPassword = hash.DummyHash()
	} else if err != nil {
		return ErrInternalError
	}

	ok := hash.CompareHashAndPassword(hashedPassword, password)

	if !(exists && ok) {
		return ErrInvalidCredentials
	}

	return nil
}

/*
	Essentially acts as a gatekeeper.

	What it does is check whether the relevant session is
	authenticated.

	To do this, we compare what the client sent, and what
	our backend has.

	First, we extract the *Session from the context
	and check if the session is authenticated. Then we
	check our DB if that same username exists. Make sure
	to also handle scanning errors.
*/

func Middleware(
	m *session.SessionManager,
	db *db.DB,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logging.GetLogger(r)

			if !IsAuthenticated(r) {
				l.Info("Request blocked; unauthenticated session")
				http.Error(w, "Unauthenticated", http.StatusForbidden)
				return
			}

			session := session.GetSession(r)

			username, ok := session.Get("username").(string)
			if !ok {
				l.Error("Non-string value in session username field")
				http.Error(w, "Unauthenticated", http.StatusForbidden)
				return
			}

			exists, err := db.UserExists(r.Context(), username)
			if err != nil {
				l.Error("Failed to query database: %w", slog.String("error", err.Error()))
				http.Error(w, "Unauthenticated", http.StatusForbidden)
				return
			}

			if !exists {
				l.Warn("Request blocked; potential hijacking")
				m.Migrate(session)
				http.Error(w, "Unauthenticated", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
