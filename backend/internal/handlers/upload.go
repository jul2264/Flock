package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

type UploadHandler struct {
	storage *services.StorageService
}

func NewUploadHandler(storage *services.StorageService) *UploadHandler {
	return &UploadHandler{storage: storage}
}



func generateRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	default:
		return ""
	}
}

// GenerateAvatarUploadURL generates a presigned URL to upload a profile avatar
// POST /upload/avatar
func (h *UploadHandler) GenerateAvatarUploadURL(w http.ResponseWriter, r *http.Request) {
	clerkID, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.storage == nil {
		respondError(w, http.StatusNotImplemented, "media uploads are not configured on the server")
		return
	}

	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ext := extensionFromContentType(req.ContentType)
	if ext == "" {
		respondError(w, http.StatusBadRequest, "unsupported content type. Only jpeg, png, gif, and webp are allowed.")
		return
	}

	sanitizedClerkID := strings.ReplaceAll(clerkID, "/", "_")
	key := fmt.Sprintf("avatars/%s_%s.%s", sanitizedClerkID, generateRandomHex(8), ext)

	uploadURL, publicURL, err := h.storage.GeneratePresignedUploadURL(r.Context(), key, req.ContentType, 15*time.Minute)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	respondJSON(w, http.StatusOK, models.UploadResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
	})
}

// GenerateEventBannerUploadURL generates a presigned URL to upload an event banner
// POST /upload/event-banner
func (h *UploadHandler) GenerateEventBannerUploadURL(w http.ResponseWriter, r *http.Request) {
	_, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.storage == nil {
		respondError(w, http.StatusNotImplemented, "media uploads are not configured on the server")
		return
	}

	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ext := extensionFromContentType(req.ContentType)
	if ext == "" {
		respondError(w, http.StatusBadRequest, "unsupported content type. Only jpeg, png, gif, and webp are allowed.")
		return
	}

	key := fmt.Sprintf("events/banners/%s.%s", generateRandomHex(16), ext)

	uploadURL, publicURL, err := h.storage.GeneratePresignedUploadURL(r.Context(), key, req.ContentType, 15*time.Minute)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	respondJSON(w, http.StatusOK, models.UploadResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
	})
}

// GenerateCommunityImageUploadURL generates a presigned URL to upload a community cover image
// POST /upload/community-image
func (h *UploadHandler) GenerateCommunityImageUploadURL(w http.ResponseWriter, r *http.Request) {
	_, ok := clerkIDFromCtx(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.storage == nil {
		respondError(w, http.StatusNotImplemented, "media uploads are not configured on the server")
		return
	}

	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ext := extensionFromContentType(req.ContentType)
	if ext == "" {
		respondError(w, http.StatusBadRequest, "unsupported content type. Only jpeg, png, gif, and webp are allowed.")
		return
	}

	key := fmt.Sprintf("communities/images/%s.%s", generateRandomHex(16), ext)

	uploadURL, publicURL, err := h.storage.GeneratePresignedUploadURL(r.Context(), key, req.ContentType, 15*time.Minute)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	respondJSON(w, http.StatusOK, models.UploadResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
	})
}
