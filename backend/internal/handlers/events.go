package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type EventHandler struct {
	events *services.EventService
}

func NewEventHandler(events *services.EventService) *EventHandler {
	return &EventHandler{events: events}
}

// CreateEvent POST /events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Basic validation
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.StartsAt.IsZero() {
		respondError(w, http.StatusBadRequest, "starts_at is required")
		return
	}

	event, err := h.events.Create(clerkID, &req)
	if err != nil {
		if err.Error() == "organizer not found" {
			respondError(w, http.StatusForbidden, "user profile not synced")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	respondJSON(w, http.StatusCreated, event)
}

// GetEvent GET /events/{id}
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	event, err := h.events.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch event")
		return
	}
	if event == nil {
		respondError(w, http.StatusNotFound, "event not found")
		return
	}

	respondJSON(w, http.StatusOK, event)
}

// ListEvents GET /events
// Optionally filter by query parameters later.
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.events.List()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	
	// If empty array, return [] instead of null
	if events == nil {
		events = []models.Event{}
	}

	respondJSON(w, http.StatusOK, events)
}
