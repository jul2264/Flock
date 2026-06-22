package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type CommunityHandler struct {
	communities *services.CommunityService
}

func NewCommunityHandler(communities *services.CommunityService) *CommunityHandler {
	return &CommunityHandler{communities: communities}
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
	communities, err := h.communities.List()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list communities")
		return
	}

	if communities == nil {
		communities = []models.Community{}
	}

	respondJSON(w, http.StatusOK, communities)
}
