package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/shared/logging"
	"github.com/adamkadda/ntumiwa/shared/models"
)

type ComposerHandler struct {
	db *db.DB
}

func NewComposerHandler(db *db.DB) *ComposerHandler {
	return &ComposerHandler{db: db}
}

func (h *ComposerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	if idStr == "" {
		switch r.Method {
		case http.MethodGet:
			h.composersGET(w, r)
		case http.MethodPost:
			h.composerPOST(w, r)
		default:
			l.Info("Unsupported method")
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.composerGET(w, r)
	case http.MethodPut:
		h.composerPUT(w, r)
	case http.MethodDelete:
		h.composerDELETE(w, r)
	default:
		l.Info("Unsupported method")
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ComposerHandler) composerGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid composer ID", slog.String("error", err.Error()), slog.Int("composer_id", id))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	composerResponse, err := h.db.GetComposer(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Composer not found", slog.Int("id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch composer", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful composer fetch")
	respondJSON(w, r, http.StatusOK, composerResponse)
}

func (h *ComposerHandler) composersGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	composers, err := h.db.GetComposers(r.Context())
	if err != nil {

		// NOTE: Not going to happen often, but nice to be ready
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Successful fetch; empty composers table")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch composers", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful composer list fetch")
	respondJSON(w, r, http.StatusOK, composers)
}

func (h *ComposerHandler) composerPOST(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	var composerRequest models.ComposerRequest
	if err := json.NewDecoder(r.Body).Decode(&composerRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	composerResponse, err := h.db.CreateComposer(r.Context(), composerRequest)
	if err != nil {
		if errors.Is(err, db.ErrIncompleteResource) {
			l.Info("Missing field(s); all fields are required", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to create composer", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful composer creation")
	respondJSON(w, r, http.StatusCreated, composerResponse)
}

func (h *ComposerHandler) composerPUT(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid composer ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var composerRequest models.ComposerRequest
	if err := json.NewDecoder(r.Body).Decode(&composerRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	composerResponse, err := h.db.UpdateComposer(r.Context(), id, composerRequest)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Composer not found", slog.Int("composer_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to update composer", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful composer update")
	respondJSON(w, r, http.StatusOK, composerResponse)
}

func (h *ComposerHandler) composerDELETE(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid composer ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Nice to have: Protect user from unintentional cascading deletes

	if err = h.db.DeleteComposer(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Composer not found", slog.Int("id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to delete composer", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful composer deletion")
	w.WriteHeader(http.StatusNoContent)
}
