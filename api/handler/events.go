package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adamkadda/ntumiwa-site/internal/db"
	"github.com/adamkadda/ntumiwa-site/shared/logging"
	"github.com/adamkadda/ntumiwa-site/shared/models"
)

type EventHandler struct {
	db *db.DB
}

func NewEventHandler(db *db.DB) *EventHandler {
	return &EventHandler{db: db}
}

// Matched to the /events/{id} pattern.
// Exposed only to authenticated users.
// Don't bother authenticating, delegate to middleware.
// Public users access /performances
func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	if idStr == "" {
		switch r.Method {
		case http.MethodGet:
			h.eventsGET(w, r)
		case http.MethodPost:
			h.eventPOST(w, r)
		default:
			l.Info("Unsupported method")
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	} else {
		switch r.Method {
		case http.MethodGet:
			h.eventGET(w, r)
		case http.MethodPut:
			h.eventPUT(w, r)
		case http.MethodDelete:
			h.eventDELETE(w, r)
		default:
			l.Info("Unsupported method")
			w.Header().Set("Allow", "GET, PUT, DELETE")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (h *EventHandler) eventGET(w http.ResponseWriter, r *http.Request) {
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

func (h *EventHandler) eventsGET(w http.ResponseWriter, r *http.Request) {
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

	events, err := h.db.GetEvents(r.Context(), timeframe, status)
	if err != nil {
		l.Error("Failed to fetch events", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful event list fetch")
	respondJSON(w, r, http.StatusOK, events)
}

func (h *EventHandler) eventPOST(w http.ResponseWriter, r *http.Request) {
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

func (h *EventHandler) eventPUT(w http.ResponseWriter, r *http.Request) {
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

		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Event not found", slog.Int("event_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrForeignKeyViolation) {
			l.Info("Invalid referenced resource, see downstream logs", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to update event", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful event update")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

func (h *EventHandler) eventDELETE(w http.ResponseWriter, r *http.Request) {
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

// Matched to the /events/{id}/draft pattern
func (h *EventHandler) DraftEvent(w http.ResponseWriter, r *http.Request) {
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

	l.Info("Success: event drafted")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

// Matched to the /events/{id}/publish pattern
func (h *EventHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
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

	l.Info("Success: event published")
	respondJSON(w, r, http.StatusOK, eventResponse)
}

// Matched to the /events/{id}/archive pattern
func (h *EventHandler) ArchiveEvent(w http.ResponseWriter, r *http.Request) {
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

	l.Info("Success: event archived")
	respondJSON(w, r, http.StatusOK, eventResponse)
}
