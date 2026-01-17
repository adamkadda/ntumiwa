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

type PieceHandler struct {
	db *db.DB
}

func NewPieceHandler(db *db.DB) *PieceHandler {
	return &PieceHandler{db: db}
}

func (h *PieceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /pieces/{id}", h.get)
	mux.HandleFunc("GET /pieces", h.list)
	mux.HandleFunc("POST /pieces", h.create)
	mux.HandleFunc("PUT /pieces/{id}", h.update)
	mux.HandleFunc("DELETE /pieces/{id}", h.delete)
}

func (h *PieceHandler) get(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid piece ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pieceResponse, err := h.db.GetPiece(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Piece not found", slog.Int("piece_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to fetch piece", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful piece fetch")
	respondJSON(w, r, http.StatusOK, pieceResponse)
}

func (h *PieceHandler) list(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	pieces, err := h.db.ListPieces(r.Context())
	if err != nil {
		l.Error("Failed to fetch pieces", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful piece list fetch")
	respondJSON(w, r, http.StatusOK, pieces)
}

func (h *PieceHandler) create(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	var pieceRequest models.PieceRequest
	if err := json.NewDecoder(r.Body).Decode(&pieceRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pieceResponse, err := h.db.CreatePiece(r.Context(), pieceRequest)
	if err != nil {

		if errors.Is(err, db.ErrTitleNotFound) {
			l.Info("Missing field; title field required", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to create piece", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Sucessful piece creation")
	respondJSON(w, r, http.StatusCreated, pieceResponse)
}

func (h *PieceHandler) update(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid piece ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var pieceRequest models.PieceRequest
	if err := json.NewDecoder(r.Body).Decode(&pieceRequest); err != nil {
		l.Info("Failed to decode body", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pieceResponse, err := h.db.UpdatePiece(r.Context(), id, pieceRequest)
	if err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Piece not found", slog.Int("piece_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, db.ErrForeignKeyViolation) {
			l.Info("Invalid composer ID", slog.String("error", err.Error()), slog.Int("composer_id", pieceRequest.ComposerID))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		l.Error("Failed to update piece", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Successful piece update")
	respondJSON(w, r, http.StatusOK, pieceResponse)
}

func (h *PieceHandler) delete(w http.ResponseWriter, r *http.Request) {
	l := logging.GetLogger(r)

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Info("Invalid piece ID", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Nice to have: Protect user from unintentional cascading deletes

	if err = h.db.DeletePiece(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrResourceNotFound) {
			l.Info("Piece not found", slog.Int("piece_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		l.Error("Failed to delete piece", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	l.Info("Sucessful piece deletion")
	w.WriteHeader(http.StatusNoContent)
}
