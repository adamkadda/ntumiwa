package handler

import (
	"database/sql"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/auth"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/session"
)

type LogoutHandler struct {
	db      *sql.DB
	manager *session.SessionManager
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	if r.Method != http.MethodPost {
		l.Warn("Failed logout: unsupported method")
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s := session.GetSession(r)

	authenticated := s.Get("authenticated")
	if authenticated == false {
		l.Warn("Failed logout: unauthenticated client")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	username := s.Get("username")

	if err := auth.Logout(h.manager, r); err != nil {
		l.Warn("Failed logout: migration error", "username", username)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful logout", "username", username)
	w.WriteHeader(http.StatusNoContent)
}
