package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type RSVPHandler struct {
	rsvps *services.RSVPService
}

func NewRSVPHandler(rsvps *services.RSVPService) *RSVPHandler {
	return &RSVPHandler{rsvps: rsvps}
}

// CreateRSVP POST /events/{id}/rsvp
func (h *RSVPHandler) CreateRSVP(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	var req models.CreateRSVPRequest
	// Body is optional, default status is "confirmed"
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
	}

	if req.Status == "" {
		req.Status = "confirmed"
	}

	// Validate status
	switch req.Status {
	case "pending", "confirmed", "cancelled", "waitlisted":
		// valid
	default:
		respondError(w, http.StatusBadRequest, "invalid status; must be pending, confirmed, cancelled, or waitlisted")
		return
	}

	rsvp, err := h.rsvps.Upsert(clerkID, eventID, req.Status)
	if err != nil {
		if err.Error() == "user profile not synced" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create rsvp")
		return
	}

	respondJSON(w, http.StatusCreated, rsvp)
}

// CancelRSVP DELETE /events/{id}/rsvp
func (h *RSVPHandler) CancelRSVP(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	err := h.rsvps.Cancel(clerkID, eventID)
	if err != nil {
		if err.Error() == "user profile not synced" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "rsvp not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to cancel rsvp")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "rsvp cancelled"})
}

// ListEventRSVPs GET /events/{id}/rsvps
func (h *RSVPHandler) ListEventRSVPs(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	rsvps, err := h.rsvps.ListEventRSVPs(eventID)
	if err != nil {
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch event rsvps")
		return
	}

	if rsvps == nil {
		rsvps = []models.RSVPWithUser{}
	}

	respondJSON(w, http.StatusOK, rsvps)
}

// ListUserRSVPs GET /users/me/rsvps
func (h *RSVPHandler) ListUserRSVPs(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rsvps, err := h.rsvps.ListUserRSVPs(clerkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user rsvps")
		return
	}

	if rsvps == nil {
		rsvps = []models.RSVPWithEvent{}
	}

	respondJSON(w, http.StatusOK, rsvps)
}
