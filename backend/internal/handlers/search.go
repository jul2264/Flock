package handlers

import (
	"net/http"

	"github.com/jul2264/Flock/backend/internal/services"
)

type SearchHandler struct {
	search *services.SearchService
}

func NewSearchHandler(search *services.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

// Search GET /search?q=query&type=events|communities|all
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	indexFilter := r.URL.Query().Get("type")

	// Even if query is empty, we allow searching to list documents or return empty list
	results, err := h.search.Search(query, indexFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search operation failed")
		return
	}

	if results == nil {
		results = []services.SearchResult{}
	}

	respondJSON(w, http.StatusOK, results)
}
