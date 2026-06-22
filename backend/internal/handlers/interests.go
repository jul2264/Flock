package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type InterestHandler struct {
	interests *services.InterestService
}

func NewInterestHandler(interests *services.InterestService) *InterestHandler {
	return &InterestHandler{interests: interests}
}

// ListInterests GET /interests
func (h *InterestHandler) ListInterests(w http.ResponseWriter, r *http.Request) {
	list, err := h.interests.ListAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list interests")
		return
	}

	if list == nil {
		list = []models.Interest{}
	}

	respondJSON(w, http.StatusOK, list)
}

// CreateInterest POST /interests
func (h *InterestHandler) CreateInterest(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	interest, err := h.interests.Create(&req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create interest")
		return
	}

	respondJSON(w, http.StatusCreated, interest)
}

// GetUserInterests GET /users/me/interests
func (h *InterestHandler) GetUserInterests(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userInterests, err := h.interests.GetUserInterests(clerkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user interests")
		return
	}

	if userInterests == nil {
		userInterests = []models.UserInterestDetail{}
	}

	respondJSON(w, http.StatusOK, userInterests)
}

// SetUserInterests POST /users/me/interests
func (h *InterestHandler) SetUserInterests(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.SetUserInterestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	err := h.interests.SetUserInterests(clerkID, req.Interests)
	if err != nil {
		if err.Error() == "user profile not synced" {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "interests updated"})
}

// GetEventInterests GET /events/{id}/interests
func (h *InterestHandler) GetEventInterests(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	eventInterests, err := h.interests.GetEventInterests(eventID)
	if err != nil {
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch event interests")
		return
	}

	if eventInterests == nil {
		eventInterests = []models.Interest{}
	}

	respondJSON(w, http.StatusOK, eventInterests)
}

// SetEventInterests POST /events/{id}/interests
func (h *InterestHandler) SetEventInterests(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	var req models.SetEventInterestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	err := h.interests.SetEventInterests(eventID, req.InterestIDs)
	if err != nil {
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to set event interests")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "event interests updated"})
}
