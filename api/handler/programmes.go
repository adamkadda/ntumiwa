package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adamkadda/ntumiwa-site/internal/db"
	"github.com/adamkadda/ntumiwa-site/shared/logging"
	"github.com/adamkadda/ntumiwa-site/shared/models"
)

type ProgrammeHandler struct {
	db *db.DB
}

func NewProgrammeHandler(db *db.DB) *ProgrammeHandler {
	return &ProgrammeHandler{db: db}
}

func (h *ProgrammeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	if idStr == "" {
		switch r.Method {
		case http.MethodGet:
			h.programmesGET(w, r)
		case http.MethodPost:
			h.programmePOST(w, r)
		default:
			l.Info("Unsupported method")
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.programmeGET(w, r)
	case http.MethodPut:
		h.programmePUT(w, r)
	case http.MethodDelete:
		h.programmeDELETE(w, r)
	default:
		l.Info("Unsupported method")
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ProgrammeHandler) programmeGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid programme ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	programmeResponse, err := h.db.GetProgramme(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Programme not found", slog.Int("programme_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch programme", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme fetch")
	respondJSON(w, r, http.StatusOK, programmeResponse)
}

func (h *ProgrammeHandler) programmesGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	programmes, err := h.db.GetProgrammes(r.Context())
	if err != nil {
		l.Error("Failed to fetch programmes", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme list fetch")
	respondJSON(w, r, http.StatusOK, programmes)
}

func (h *ProgrammeHandler) programmePOST(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	var programmeRequest models.ProgrammeRequest
	if err := json.NewDecoder(r.Body).Decode(&programmeRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	programmeResponse, err := h.db.CreateProgramme(r.Context(), programmeRequest)
	if err != nil {
		if errors.Is(err, db.ErrTitleNotFound) {
			l.Info("Missing field; title field required", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to create programme", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme creation")
	respondJSON(w, r, http.StatusCreated, programmeResponse)
}

func (h *ProgrammeHandler) programmePUT(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid programme ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var programmeRequest models.ProgrammeRequest
	if err := json.NewDecoder(r.Body).Decode(&programmeRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	programmeResponse, err := h.db.UpdateProgramme(r.Context(), id, programmeRequest)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Programme not found", slog.Int("programme_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrForeignKeyViolation) {
			l.Info("Piece not found; see downstream logs", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to update programme", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme update")
	respondJSON(w, r, http.StatusOK, programmeResponse)
}

func (h *ProgrammeHandler) programmeDELETE(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid programme ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.db.DeleteProgramme(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Programme not found", slog.Int("programme_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to delete programme", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme deletion")
	w.WriteHeader(http.StatusNoContent)
}
