package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/models"
)

type ProgrammeHandler struct {
	db *db.DB
}

func NewProgrammeHandler(db *db.DB) *ProgrammeHandler {
	return &ProgrammeHandler{db: db}
}

func (h *ProgrammeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programmes/{id}", h.get)
	mux.HandleFunc("GET /programmes", h.list)
	mux.HandleFunc("POST /programmes", h.create)
	mux.HandleFunc("PUT /programmes/{id}", h.update)
	mux.HandleFunc("DELETE /programmes/{id}", h.delete)
}

func (h *ProgrammeHandler) get(w http.ResponseWriter, r *http.Request) {
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

func (h *ProgrammeHandler) list(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	programmes, err := h.db.ListProgrammes(r.Context())
	if err != nil {
		l.Error("Failed to fetch programmes", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful programme list fetch")
	respondJSON(w, r, http.StatusOK, programmes)
}

func (h *ProgrammeHandler) create(w http.ResponseWriter, r *http.Request) {
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

func (h *ProgrammeHandler) update(w http.ResponseWriter, r *http.Request) {
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

func (h *ProgrammeHandler) delete(w http.ResponseWriter, r *http.Request) {
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
