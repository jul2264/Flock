package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// SyncEvent indexes or updates an event document in Meilisearch
func (s *SearchService) SyncEvent(event *models.Event) error {
	doc := map[string]interface{}{
		"id":          event.ID,
		"type":        "event",
		"title":       event.Title,
		"description": event.Description,
		"venue_name":  event.VenueName,
		"start_date":  event.StartsAt.Unix(),
		"status":      event.Status,
		"created_at":  event.CreatedAt.Unix(),
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
		"created_at":  community.CreatedAt.Unix(),
	}
	return s.indexDocument("communities", doc)
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

	// If index doesn't exist, Meilisearch will automatically create it.
	// We handle successful response codes (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("meilisearch returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

type SearchResult struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"` // "event" or "community"
	Title       string  `json:"title"` // mapped from title (event) or name (community)
	Description *string `json:"description"`
	VenueName   *string `json:"venue_name,omitempty"`
	City        *string `json:"city,omitempty"`
	StartDate   *int64  `json:"start_date,omitempty"`
}

// Search queries Meilisearch for matching records across indexes
func (s *SearchService) Search(query string, indexFilter string) ([]SearchResult, error) {
	var results []SearchResult

	var indexes []string
	if indexFilter == "events" {
		indexes = []string{"events"}
	} else if indexFilter == "communities" {
		indexes = []string{"communities"}
	} else {
		indexes = []string{"events", "communities"}
	}

	for _, index := range indexes {
		idxResults, err := s.queryIndex(index, query)
		if err != nil {
			// If index is empty/doesn't exist, we skip rather than failing
			log.Printf("Warning: failed to query Meilisearch index %s: %v", index, err)
			continue
		}
		results = append(results, idxResults...)
	}

	return results, nil
}

func (s *SearchService) queryIndex(indexName string, query string) ([]SearchResult, error) {
	searchPayload := map[string]interface{}{
		"q":     query,
		"limit": 20,
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
