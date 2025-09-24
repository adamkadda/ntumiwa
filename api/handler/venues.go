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

type VenueHandler struct {
	db *db.DB
}

func NewVenueHandler(db *db.DB) *VenueHandler {
	return &VenueHandler{db: db}
}

func (h *VenueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	if idStr == "" {
		switch r.Method {
		case http.MethodGet:
			h.venuesGET(w, r)
		case http.MethodPost:
			h.venuePOST(w, r)
		default:
			l.Info("Unsupported method")
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.venueGET(w, r)
	case http.MethodPut:
		h.venuePUT(w, r)
	case http.MethodDelete:
		h.venueDELETE(w, r)
	default:
		l.Info("Unsupported method")
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *VenueHandler) venueGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid venue ID", slog.String("error", err.Error()), slog.Int("venue_id", id))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	venueResponse, err := h.db.GetVenue(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("venue not found", slog.Int("id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch venue", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful venue fetch")
	respondJSON(w, r, http.StatusOK, venueResponse)
}

func (h *VenueHandler) venuesGET(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	venues, err := h.db.GetVenues(r.Context())
	if err != nil {
		l.Warn("Failed to fetch venues", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful venue list fetch")
	respondJSON(w, r, http.StatusOK, venues)
}

func (h *VenueHandler) venuePOST(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	var venueRequest models.VenueRequest
	if err := json.NewDecoder(r.Body).Decode(&venueRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	venueResponse, err := h.db.CreateVenue(r.Context(), venueRequest)
	if err != nil {
		if errors.Is(err, db.ErrIncompleteResource) {
			l.Info("Address field required", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to create venue", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful venue creation")
	respondJSON(w, r, http.StatusCreated, venueResponse)
}

func (h *VenueHandler) venuePUT(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid venue ID", slog.String("error", err.Error()), slog.Int("venue_id", id))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var venueRequest models.VenueRequest
	if err := json.NewDecoder(r.Body).Decode(&venueRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	venueResponse, err := h.db.UpdateVenue(r.Context(), id, venueRequest)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Venue not found", slog.Int("venue_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to update venue", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Succesful venue update")
	respondJSON(w, r, http.StatusOK, venueResponse)
}

func (h *VenueHandler) venueDELETE(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid venue ID", slog.String("error", err.Error()), slog.Int("venue_id", id))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Nice to have: Protect user from unintentional cascading deletes

	if err = h.db.DeleteVenue(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Venue not found", slog.Int("venue_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to delete venue", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful venue deletion")
	w.WriteHeader(http.StatusNoContent)
}
