// Package handler exposes the HTTP API for ride hailing data. It is API-only
// (JSON in, JSON out) and built on the standard library net/http.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/testaccount4535/ai_agentic_workshop/internal/model"
	"github.com/testaccount4535/ai_agentic_workshop/internal/store"
)

// Handler wires HTTP requests to the persistent store.
type Handler struct {
	store *store.Store
	log   *slog.Logger
}

// New builds a Handler. A nil logger falls back to the default slog logger.
func New(st *store.Store, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{store: st, log: logger}
}

// Routes returns the API router. Method-specific patterns mean unsupported
// methods automatically yield 405 Method Not Allowed.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rides", h.startRide)
	return mux
}

// errorResponse is the uniform JSON error envelope.
type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) startRide(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()

	var ride model.RideStart
	if err := dec.Decode(&ride); err != nil {
		h.log.Warn("decode ride start", "remote", r.RemoteAddr, "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	// Reject trailing data / multiple JSON values in the body.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.log.Warn("unexpected trailing data in ride start body", "remote", r.RemoteAddr)
		h.writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	if err := ride.Validate(); err != nil {
		h.log.Warn("ride start validation failed", "ride_id", ride.ID, "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch err := h.store.SaveRideStart(ride); {
	case err == nil:
		// success
	case errors.Is(err, store.ErrDuplicateRide):
		h.writeError(w, http.StatusConflict, err.Error())
		return
	default:
		h.log.Error("persist ride start", "ride_id", ride.ID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "could not save ride")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ride); err != nil {
		h.log.Error("encode ride start response", "ride_id", ride.ID, "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: msg}); err != nil {
		h.log.Error("encode error response", "error", err)
	}
}
