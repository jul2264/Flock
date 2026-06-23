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

type CommunityHandler struct {
	communities *services.CommunityService
	users       *services.UserService
}

func NewCommunityHandler(communities *services.CommunityService, users *services.UserService) *CommunityHandler {
	return &CommunityHandler{communities: communities, users: users}
}

// CreateCommunity POST /communities
func (h *CommunityHandler) CreateCommunity(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateCommunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	community, err := h.communities.Create(clerkID, &req)
	if err != nil {
		if err.Error() == "owner not found" {
			respondError(w, http.StatusForbidden, "user profile not synced")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create community")
		return
	}

	respondJSON(w, http.StatusCreated, community)
}

// GetCommunity GET /communities/{id}
func (h *CommunityHandler) GetCommunity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	community, err := h.communities.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch community")
		return
	}
	if community == nil {
		respondError(w, http.StatusNotFound, "community not found")
		return
	}

	respondJSON(w, http.StatusOK, community)
}

// ListCommunities GET /communities
func (h *CommunityHandler) ListCommunities(w http.ResponseWriter, r *http.Request) {
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

	filter := models.CommunityFilter{
		Lat:      pLat,
		Lng:      pLng,
		RadiusKm: pRadius,
		Sort:     r.URL.Query().Get("sort"),
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

	communities, err := h.communities.List(&filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list communities")
		return
	}

	if communities == nil {
		communities = []models.Community{}
	}

	respondJSON(w, http.StatusOK, communities)
}

// UpdateCommunity PATCH /communities/{id}
func (h *CommunityHandler) UpdateCommunity(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateCommunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	community, err := h.communities.Update(id, clerkID, role, &req)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden: insufficient permissions")
			return
		}
		if err.Error() == "community not found" {
			respondError(w, http.StatusNotFound, "community not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update community")
		return
	}

	respondJSON(w, http.StatusOK, community)
}

// DeleteCommunity DELETE /communities/{id}
func (h *CommunityHandler) DeleteCommunity(w http.ResponseWriter, r *http.Request) {
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

	err := h.communities.Delete(id, clerkID, role)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden: insufficient permissions")
			return
		}
		if err.Error() == "community not found" {
			respondError(w, http.StatusNotFound, "community not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete community")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// JoinCommunity POST /communities/{id}/join
func (h *CommunityHandler) JoinCommunity(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	err := h.communities.Join(id, clerkID)
	if err != nil {
		if err.Error() == "community not found" {
			respondError(w, http.StatusNotFound, "community not found")
			return
		}
		if err.Error() == "community is full" || err.Error() == "already a member" || err.Error() == "cannot join a deactivated community" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to join community")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

// LeaveCommunity DELETE /communities/{id}/leave
func (h *CommunityHandler) LeaveCommunity(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	err := h.communities.Leave(id, clerkID)
	if err != nil {
		if err.Error() == "not a member of this community" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to leave community")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

// ListCommunityMembers GET /communities/{id}/members
func (h *CommunityHandler) ListCommunityMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	members, err := h.communities.ListMembers(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list community members")
		return
	}

	respondJSON(w, http.StatusOK, members)
}
