package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/models"
)

type PerformanceHandler struct {
	db *db.DB
}

func NewPerformanceHandler(db *db.DB) *PerformanceHandler {
	return &PerformanceHandler{db: db}
}

// The endpoints exposed by this handler are available
// to the public frontend. Requesting for individual
// performances from the public frontend is forbidden.
func (h *PerformanceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /performances", h.list)
}

func (h *PerformanceHandler) list(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	query := r.URL.Query()

	timeframe := query.Get("timeframe")

	switch timeframe {
	case models.TimeframeUpcoming:
	case models.TimeframePast:
	default:
		l.Info("Failed performances fetch", slog.String("error", "invalid timeframe query parameter"), slog.String("timeframe", timeframe))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	performances, err := h.db.ListPerformances(r.Context(), timeframe)
	if err != nil {
		l.Error("Failed to fetch performances", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info(fmt.Sprintf("Successful %q performances fetch", timeframe))
	respondJSON(w, r, http.StatusOK, performances)
}
