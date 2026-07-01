package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jul2264/Flock/backend/internal/models"
)

type SearchService struct {
	url       string
	masterKey string
	client    *http.Client
}

func NewSearchService() *SearchService {
	url := os.Getenv("MEILI_URL")
	if url == "" {
		url = "http://localhost:7700"
	}
	key := os.Getenv("MEILI_MASTER_KEY")
	if key == "" {
		key = "masterKey"
	}
	return &SearchService{
		url:       url,
		masterKey: key,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ConfigureSettings updates filterable and sortable attributes on indexes.
func (s *SearchService) ConfigureSettings() {
	// Configure events index settings
	eventSettings := map[string]interface{}{
		"filterableAttributes": []string{"interests", "status", "_geo"},
		"sortableAttributes":   []string{"start_date", "_geo"},
	}
	if err := s.updateIndexSettings("events", eventSettings); err != nil {
		log.Printf("Warning: failed to update Meilisearch events settings: %v", err)
	}

	// Configure communities index settings
	communitySettings := map[string]interface{}{
		"filterableAttributes": []string{"visibility", "status", "_geo"},
		"sortableAttributes":   []string{"_geo"},
	}
	if err := s.updateIndexSettings("communities", communitySettings); err != nil {
		log.Printf("Warning: failed to update Meilisearch communities settings: %v", err)
	}
}

func (s *SearchService) updateIndexSettings(indexName string, settings map[string]interface{}) error {
	body, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/indexes/%s/settings", s.url, indexName)
	req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.masterKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("meilisearch settings returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SyncEvent indexes or updates an event document in Meilisearch
func (s *SearchService) SyncEvent(event *models.Event, interestIDs []string) error {
	doc := map[string]interface{}{
		"id":          event.ID,
		"type":        "event",
		"title":       event.Title,
		"description": event.Description,
		"venue_name":  event.VenueName,
		"start_date":  event.StartsAt.Unix(),
		"status":      event.Status,
		"interests":   interestIDs,
		"created_at":  event.CreatedAt.Unix(),
	}

	if event.Latitude != nil && event.Longitude != nil {
		doc["_geo"] = map[string]float64{
			"lat": *event.Latitude,
			"lng": *event.Longitude,
		}
	}

	return s.indexDocument("events", doc)
}

// SyncCommunity indexes or updates a community document in Meilisearch
func (s *SearchService) SyncCommunity(community *models.Community) error {
	doc := map[string]interface{}{
		"id":          community.ID,
		"type":        "community",
		"name":        community.Name,
		"description": community.Description,
		"city":        community.City,
		"visibility":  community.Visibility,
		"status":      community.Status,
		"created_at":  community.CreatedAt.Unix(),
	}

	if community.Latitude != nil && community.Longitude != nil {
		doc["_geo"] = map[string]float64{
			"lat": *community.Latitude,
			"lng": *community.Longitude,
		}
	}

	return s.indexDocument("communities", doc)
}

// DeleteEvent removes an event document from Meilisearch
func (s *SearchService) DeleteEvent(eventID string) error {
	return s.deleteDocument("events", eventID)
}

// DeleteCommunity removes a community document from Meilisearch
func (s *SearchService) DeleteCommunity(communityID string) error {
	return s.deleteDocument("communities", communityID)
}

func (s *SearchService) deleteDocument(indexName string, docID string) error {
	reqURL := fmt.Sprintf("%s/indexes/%s/documents/%s", s.url, indexName, docID)
	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	if s.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.masterKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("meilisearch delete returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *SearchService) indexDocument(indexName string, doc map[string]interface{}) error {
	docs := []map[string]interface{}{doc}
	body, err := json.Marshal(docs)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/indexes/%s/documents", s.url, indexName)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.masterKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("meilisearch returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

type SearchResult struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	VenueName   *string `json:"venue_name,omitempty"`
	City        *string `json:"city,omitempty"`
	StartDate   *int64  `json:"start_date,omitempty"`
}

type SearchParams struct {
	Query      string
	Type       string
	InterestID string
	Lat        *float64
	Lng        *float64
	RadiusKm   *float64
}

// Search queries Meilisearch for matching records across indexes with filters.
func (s *SearchService) Search(params SearchParams) ([]SearchResult, error) {
	var results []SearchResult

	var indexes []string
	if params.Type == "events" {
		indexes = []string{"events"}
	} else if params.Type == "communities" {
		indexes = []string{"communities"}
	} else {
		indexes = []string{"events", "communities"}
	}

	for _, index := range indexes {
		idxResults, err := s.queryIndex(index, params)
		if err != nil {
			log.Printf("Warning: failed to query Meilisearch index %s: %v", index, err)
			continue
		}
		results = append(results, idxResults...)
	}

	return results, nil
}

func (s *SearchService) queryIndex(indexName string, params SearchParams) ([]SearchResult, error) {
	var filters []string
	if indexName == "events" {
		filters = append(filters, "status = 'published'")
		if params.InterestID != "" {
			filters = append(filters, fmt.Sprintf("interests = '%s'", params.InterestID))
		}
	} else if indexName == "communities" {
		filters = append(filters, "status = 'active'", "visibility = 'public'")
	}

	if params.Lat != nil && params.Lng != nil && params.RadiusKm != nil {
		radiusMeters := int(*params.RadiusKm * 1000)
		filters = append(filters, fmt.Sprintf("_geoRadius(%f, %f, %d)", *params.Lat, *params.Lng, radiusMeters))
	}

	searchPayload := map[string]interface{}{
		"q":     params.Query,
		"limit": 20,
	}

	if len(filters) > 0 {
		searchPayload["filter"] = strings.Join(filters, " AND ")
	}

	if params.Lat != nil && params.Lng != nil {
		searchPayload["sort"] = []string{fmt.Sprintf("_geoPoint(%f, %f):asc", *params.Lat, *params.Lng)}
	}

	body, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/indexes/%s/search", s.url, indexName)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.masterKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("meilisearch search error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var meiliResp struct {
		Hits []map[string]interface{} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meiliResp); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, hit := range meiliResp.Hits {
		id, _ := hit["id"].(string)
		itemType, _ := hit["type"].(string)

		var title string
		if itemType == "community" {
			title, _ = hit["name"].(string)
		} else {
			title, _ = hit["title"].(string)
		}

		var desc *string
		if val, ok := hit["description"].(string); ok {
			desc = &val
		}

		var venue *string
		if val, ok := hit["venue_name"].(string); ok {
			venue = &val
		}

		var city *string
		if val, ok := hit["city"].(string); ok {
			city = &val
		}

		var startDate *int64
		if val, ok := hit["start_date"].(float64); ok {
			ts := int64(val)
			startDate = &ts
		}

		results = append(results, SearchResult{
			ID:          id,
			Type:        itemType,
			Title:       title,
			Description: desc,
			VenueName:   venue,
			City:        city,
			StartDate:   startDate,
		})
	}

	return results, nil
}

type AutocompleteResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// Autocomplete retrieves suggestion matches restricted to top 5 recommendations.
func (s *SearchService) Autocomplete(query string, indexFilter string) ([]AutocompleteResult, error) {
	var results []AutocompleteResult

	var indexes []string
	if indexFilter == "events" {
		indexes = []string{"events"}
	} else if indexFilter == "communities" {
		indexes = []string{"communities"}
	} else {
		indexes = []string{"events", "communities"}
	}

	for _, index := range indexes {
		idxResults, err := s.queryAutocompleteIndex(index, query)
		if err != nil {
			log.Printf("Warning: failed to query Meilisearch autocomplete for index %s: %v", index, err)
			continue
		}
		results = append(results, idxResults...)
	}

	return results, nil
}

func (s *SearchService) queryAutocompleteIndex(indexName string, query string) ([]AutocompleteResult, error) {
	var filters []string
	if indexName == "events" {
		filters = append(filters, "status = 'published'")
	} else if indexName == "communities" {
		filters = append(filters, "status = 'active'", "visibility = 'public'")
	}

	searchPayload := map[string]interface{}{
		"q":                    query,
		"limit":                5,
		"attributesToRetrieve": []string{"id", "title", "name", "type"},
	}

	if len(filters) > 0 {
		searchPayload["filter"] = strings.Join(filters, " AND ")
	}

	body, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/indexes/%s/search", s.url, indexName)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.masterKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.masterKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("meilisearch search error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var meiliResp struct {
		Hits []map[string]interface{} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meiliResp); err != nil {
		return nil, err
	}

	var results []AutocompleteResult
	for _, hit := range meiliResp.Hits {
		id, _ := hit["id"].(string)
		itemType, _ := hit["type"].(string)

		var title string
		if itemType == "community" {
			title, _ = hit["name"].(string)
		} else {
			title, _ = hit["title"].(string)
		}

		results = append(results, AutocompleteResult{
			ID:    id,
			Title: title,
			Type:  itemType,
		})
	}

	return results, nil
}
