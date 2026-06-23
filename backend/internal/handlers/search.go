package handlers

import (
	"net/http"
	"strconv"

	"github.com/jul2264/Flock/backend/internal/services"
)

type SearchHandler struct {
	search *services.SearchService
}

func NewSearchHandler(search *services.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

// Search GET /search?q=query&type=events|communities|all&interest_id=uuid&lat=x&lng=y&radius=km
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	var lat *float64
	var lng *float64
	var radiusKm *float64

	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if val, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &val
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		if val, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &val
		}
	}
	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		if val, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radiusKm = &val
		}
	}

	params := services.SearchParams{
		Query:      r.URL.Query().Get("q"),
		Type:       r.URL.Query().Get("type"),
		InterestID: r.URL.Query().Get("interest_id"),
		Lat:        lat,
		Lng:        lng,
		RadiusKm:   radiusKm,
	}

	results, err := h.search.Search(params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search operation failed")
		return
	}

	if results == nil {
		results = []services.SearchResult{}
	}

	respondJSON(w, http.StatusOK, results)
}

// Autocomplete GET /search/autocomplete?q=query&type=events|communities|all
func (h *SearchHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	indexFilter := r.URL.Query().Get("type")

	results, err := h.search.Autocomplete(query, indexFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "autocomplete search failed")
		return
	}

	if results == nil {
		results = []services.AutocompleteResult{}
	}

	respondJSON(w, http.StatusOK, results)
}
