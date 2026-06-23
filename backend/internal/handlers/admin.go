package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type AdminHandler struct {
	admin  *services.AdminService
	events *services.EventService
}

func NewAdminHandler(admin *services.AdminService, events *services.EventService) *AdminHandler {
	return &AdminHandler{
		admin:  admin,
		events: events,
	}
}

// ListUsers GET /admin/users?page=1&limit=20
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	if limit > 50 {
		limit = 50
	}

	res, err := h.admin.ListUsers(page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	respondJSON(w, http.StatusOK, res)
}

// UpdateUserRole PATCH /admin/users/{id}/role
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user id is required")
		return
	}

	var req models.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Role != "user" && req.Role != "organizer" && req.Role != "admin" {
		respondError(w, http.StatusBadRequest, "invalid role: must be user, organizer, or admin")
		return
	}

	u, err := h.admin.UpdateUserRole(userID, req.Role)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update user role")
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// DeleteEvent DELETE /admin/events/{id}
func (h *AdminHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		respondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	adminClerkID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || adminClerkID == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.events.Delete(eventID, adminClerkID, "admin")
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		if err.Error() == "event not found" {
			respondError(w, http.StatusNotFound, "event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to moderate event")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetStats GET /admin/stats
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.admin.GetStats()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to compile stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
