package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type EventHandler struct {
	events *services.EventService
	users  *services.UserService
}

func NewEventHandler(events *services.EventService, users *services.UserService) *EventHandler {
	return &EventHandler{events: events, users: users}
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
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	clerkID, _ := clerkIDFromCtx(r)

	var latVal, lngVal, radiusVal float64
	var hasLocation bool

	latParam := r.URL.Query().Get("lat")
	lngParam := r.URL.Query().Get("lng")
	radiusParam := r.URL.Query().Get("radius")

	if latParam != "" && lngParam != "" {
		var errLat, errLng error
		latVal, errLat = strconv.ParseFloat(latParam, 64)
		lngVal, errLng = strconv.ParseFloat(lngParam, 64)
		if errLat == nil && errLng == nil {
			hasLocation = true
		}
	}

	// Fallback to user profile location and search radius preferences
	var userSearchRadius float64 = 10.0
	if !hasLocation && clerkID != "" {
		if u, err := h.users.GetByClerkID(clerkID); err == nil && u != nil {
			if u.Latitude != nil && u.Longitude != nil {
				latVal = *u.Latitude
				lngVal = *u.Longitude
				hasLocation = true
				if u.SearchRadius > 0 {
					userSearchRadius = float64(u.SearchRadius)
				}
			}
		}
	}

	var pLat, pLng, pRadius *float64
	if hasLocation {
		pLat = &latVal
		pLng = &lngVal

		if radiusParam != "" {
			if rVal, err := strconv.ParseFloat(radiusParam, 64); err == nil {
				radiusVal = rVal
				pRadius = &radiusVal
			}
		}
		if pRadius == nil {
			pRadius = &userSearchRadius
		}
	}

	// Parse other query parameters
	filter := models.EventFilter{
		Lat:      pLat,
		Lng:      pLng,
		RadiusKm: pRadius,
		Sort:     r.URL.Query().Get("sort"),
	}

	if interestParam := r.URL.Query().Get("interest"); interestParam != "" {
		filter.InterestID = &interestParam
	}

	if ageMinParam := r.URL.Query().Get("age_min"); ageMinParam != "" {
		if val, err := strconv.Atoi(ageMinParam); err == nil {
			filter.AgeMin = &val
		}
	}
	if ageMaxParam := r.URL.Query().Get("age_max"); ageMaxParam != "" {
		if val, err := strconv.Atoi(ageMaxParam); err == nil {
			filter.AgeMax = &val
		}
	}

	if fromParam := r.URL.Query().Get("from"); fromParam != "" {
		if val, err := time.Parse(time.RFC3339, fromParam); err == nil {
			filter.From = &val
		} else if val, err := time.Parse("2006-01-02", fromParam); err == nil {
			filter.From = &val
		}
	}
	if toParam := r.URL.Query().Get("to"); toParam != "" {
		if val, err := time.Parse(time.RFC3339, toParam); err == nil {
			filter.To = &val
		} else if val, err := time.Parse("2006-01-02", toParam); err == nil {
			filter.To = &val
		}
	}

	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if val, err := strconv.Atoi(pageParam); err == nil && val > 0 {
			page = val
		}
	}
	limit := 20
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 {
			limit = val
		}
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	events, err := h.events.List(&filter)
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

// UpdateEvent PATCH /events/{id}
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role := middleware.RoleFromCtx(r)

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req models.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	event, err := h.events.Update(id, clerkID, role, &req)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden: insufficient permissions")
			return
		}
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, "event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update event")
		return
	}

	respondJSON(w, http.StatusOK, event)
}

// DeleteEvent DELETE /events/{id}
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role := middleware.RoleFromCtx(r)

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	err := h.events.Delete(id, clerkID, role)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden: insufficient permissions")
			return
		}
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, "event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
