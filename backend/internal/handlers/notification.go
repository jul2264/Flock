package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type NotificationHandler struct {
	notifications    *services.NotificationService
	communityService *services.CommunityService
	userService      *services.UserService
}

func NewNotificationHandler(
	notifications *services.NotificationService,
	communityService *services.CommunityService,
	userService *services.UserService,
) *NotificationHandler {
	return &NotificationHandler{
		notifications:    notifications,
		communityService: communityService,
		userService:      userService,
	}
}

// RegisterFCMToken handles POST /users/me/fcm-token
func (h *NotificationHandler) RegisterFCMToken(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.RegisterFCMTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	user, err := h.userService.GetByClerkID(clerkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user profile")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user profile not found")
		return
	}

	err = h.notifications.RegisterToken(r.Context(), user.ID, req.Token)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to register token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SendCommunityAnnouncement handles POST /communities/{id}/announcements
func (h *NotificationHandler) SendCommunityAnnouncement(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	communityID := chi.URLParam(r, "id")
	if communityID == "" {
		respondError(w, http.StatusBadRequest, "community id is required")
		return
	}

	var req models.CommunityAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title == "" || req.Body == "" {
		respondError(w, http.StatusBadRequest, "title and body are required")
		return
	}

	user, err := h.userService.GetByClerkID(clerkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user profile")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user profile not found")
		return
	}

	ownerID, err := h.communityService.GetCommunityOwner(communityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify community ownership")
		return
	}

	if ownerID != user.ID && user.Role != "admin" {
		respondError(w, http.StatusForbidden, "forbidden: only community owner or admin can send announcements")
		return
	}

	err = h.notifications.SendCommunityAnnouncement(r.Context(), communityID, req.Title, req.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to send announcement")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
