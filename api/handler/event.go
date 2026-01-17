package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/models"
)

type EventHandler struct {
	db *db.DB
}

func NewEventHandler(db *db.DB) *EventHandler {
	return &EventHandler{db: db}
}

func (h *EventHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /events/{id}", h.get)
	mux.HandleFunc("GET /events", h.list)
	mux.HandleFunc("POST /events", h.create)
	mux.HandleFunc("PUT /events/{id}", h.update)
	mux.HandleFunc("PATCH /events/{id}/draft", h.draft)
	mux.HandleFunc("PATCH /events/{id}/publish", h.publish)
	mux.HandleFunc("PATCH /events/{id}/archive", h.archive)
	mux.HandleFunc("DELETE /events/{id}", h.delete)
}

func (h *EventHandler) get(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := h.db.GetEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch event", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful event fetch")
	respondJSON(w, r, http.StatusOK, event)
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	query := r.URL.Query()

	timeframe := query.Get("timeframe")
	status := query.Get("status")

	// NOTE: Only a single value per filter is supported
	if err := validateEventsFilters(timeframe, status); err != nil {
		l.Info("Failed events fetch", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	events, err := h.db.ListEvents(r.Context(), timeframe, status)
	if err != nil {
		l.Error("Failed to fetch events", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful event list fetch")
	respondJSON(w, r, http.StatusOK, events)
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	var eventRequest models.EventRequest
	if err := json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventResponse, err := h.db.CreateEvent(r.Context(), eventRequest)
	if err != nil {

		if errors.Is(err, db.ErrTitleNotFound) {
			l.Info("Missing field; title field required", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to create event", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful event creation")
	respondJSON(w, r, http.StatusCreated, eventResponse)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var eventRequest models.EventRequest
	if err := json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventResponse, err := h.db.UpdateEvent(r.Context(), id, eventRequest)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrResourceNotFound):
			l.Info("Event not found", slog.Int("event_id", id))
			w.WriteHeader(http.StatusNotFound)
			return

		case errors.Is(err, db.ErrForeignKeyViolation):
			l.Info("Invalid referenced resource, see downstream logs", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		case errors.Is(err, db.ErrImmutableState):
			l.Info("Event is in immutable state", slog.Int("event_id", id), slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		case errors.Is(err, db.ErrEmptyRequest):
			l.Info("Received empty request body", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		default:
			l.Error("Failed to update event", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	l.Info("Successful event update")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.db.DeleteEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to delete event", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Event deleted successfully", slog.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}

var allowedTimeframes = map[string]bool{
	"upcoming": true,
	"past":     true,
}

var allowedStatuses = map[string]bool{
	"draft":     true,
	"published": true,
	"archived":  true,
}

func validateEventsFilters(timeframe, status string) error {
	if timeframe != "" {
		if !allowedTimeframes[timeframe] {
			return fmt.Errorf("invalid timeframe: %s", timeframe)
		}
	}

	if status != "" {
		if !allowedStatuses[status] {
			return fmt.Errorf("invalid status: %s", status)
		}
	}

	return nil
}

func (h *EventHandler) draft(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventResponse, err := h.db.DraftEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrStatusUnchanged) {
			l.Info("Success; event already draft")
			respondJSON(w, r, http.StatusOK, eventResponse)
		}

		l.Error("Failed to draft event", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.Info("Event drafted")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

func (h *EventHandler) publish(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventResponse, err := h.db.PublishEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrStatusUnchanged) {
			l.Info("Success; event already published")
			respondJSON(w, r, http.StatusOK, eventResponse)
		}

		l.Error("Failed to publish event", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.Info("Event published")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

func (h *EventHandler) archive(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid event ID", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventResponse, err := h.db.ArchiveEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrStatusUnchanged) {
			l.Info("Success; event already archived")
			respondJSON(w, r, http.StatusOK, eventResponse)
		}

		l.Error("Failed to archive event", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.Info("Event archived")
	respondJSON(w, r, http.StatusOK, eventResponse)
}
