package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// clerkIDFromCtx is a small helper that extracts the Clerk user ID stored by
// ClerkMiddleware. It returns ("", false) if the value is missing or wrong type.
func clerkIDFromCtx(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(middleware.UserIDKey).(string)
	return id, ok && id != ""
}

// SyncUser  POST /users/sync
//
// Called by the client right after a successful Clerk sign-in or sign-up.
// Upserts the user row so our DB stays in sync with Clerk.
//
// Request body:
//
//	{ "email": "...", "full_name": "...", "avatar_url": "..." }
func (h *UserHandler) SyncUser(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.SyncUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	user, err := h.users.Upsert(clerkID, req.Email, req.FullName, req.AvatarURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to sync user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// GetMe  GET /users/me
//
// Returns the current user's full profile.
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.users.GetByClerkID(clerkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found — call /users/sync first")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// UpdateMe  PATCH /users/me
//
// Partially updates the current user's profile. Only fields present in the
// JSON body are written; omitted fields are left unchanged.
//
// Example body (all optional):
//
//	{
//	  "full_name": "Jane Doe",
//	  "username": "janedoe",
//	  "city": "London",
//	  "latitude": 51.5074,
//	  "longitude": -0.1278,
//	  "search_radius": 20
//	}
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := h.users.Update(clerkID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found — call /users/sync first")
		return
	}

	respondJSON(w, http.StatusOK, user)
}
