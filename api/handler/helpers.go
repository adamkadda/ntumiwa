package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/logging"
)

/*
	respondJSON

	Pre-marshalling adds CPU and memory overhead. However it allows us
	to return the appropriate status code.

	I will consider alternatives if this becomes an issue.

	Also, don't forget to keep writing your own logs.
	This function only prepares logs in the event of an error.
*/

func respondJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")

	body, err := json.Marshal(data)
	if err != nil {
		l := logging.GetLogger(r)
		l.Error("Failed to marshal JSON", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(body)
}
