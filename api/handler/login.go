package handler

import (
	"database/sql"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/auth"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/session"
)

type LoginHandler struct {
	db      *sql.DB
	manager *session.SessionManager
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	method := r.Method

	switch method {
	case http.MethodGet:
		h.loginGET(w, r)
	case http.MethodPost:
		h.loginPOST(w, r)
	default:
		l.Warn("Unsupported method")
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *LoginHandler) loginGET(w http.ResponseWriter, r *http.Request) {
	s := session.GetSession(r)
	l := logging.GetLogger(r)

	authenticated := s.Get("authenticated")
	if authenticated == true {
		// NOTE: NoContent -> signal frontend to redirect
		l.Info("Already logged in.")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *LoginHandler) loginPOST(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	username := r.FormValue("username")
	password := r.FormValue("password")

	err := auth.VerifyCredentials(h.db, username, password)
	if err != nil {
		l.Warn("Failed login", "username", username)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
		return
	}

	auth.Login(h.manager, r, username)

	l.Info("Successful login", "username", username)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login success"))
}
