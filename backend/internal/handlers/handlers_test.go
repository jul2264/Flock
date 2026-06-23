package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/models"
	"github.com/jul2264/Flock/backend/internal/services"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Load .env
	_ = godotenv.Load("../../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping database integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Failed to ping database (skipping database integration tests): %v", err)
	}

	// Clean up tables (using CASCADE to handle foreign key dependencies)
	tables := []string{"event_interests", "user_interests", "rsvps", "events", "communities", "interests", "users"}
	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			t.Fatalf("Failed to clean table %s: %v", table, err)
		}
	}

	return db
}

func TestBackendFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Instantiate services
	userService := services.NewUserService(db)
	eventService := services.NewEventService(db, nil) // pass nil search client for basic DB testing
	communityService := services.NewCommunityService(db, nil)
	rsvpService := services.NewRSVPService(db)
	interestService := services.NewInterestService(db)

	// Instantiate handlers
	userHandler := NewUserHandler(userService)
	eventHandler := NewEventHandler(eventService, userService)
	communityHandler := NewCommunityHandler(communityService, userService)
	rsvpHandler := NewRSVPHandler(rsvpService)
	interestHandler := NewInterestHandler(interestService)

	var userUUID string
	clerkID := "clerk_test_user_123"

	// 1. Test SyncUser
	t.Run("SyncUser", func(t *testing.T) {
		body := `{"email":"test@example.com","full_name":"Test User","avatar_url":"http://example.com/avatar.jpg"}`
		req := httptest.NewRequest("POST", "/users/sync", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.SyncUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp models.User
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", resp.Email)
		}
		userUUID = resp.ID
	})

	// 2. Test GetMe
	t.Run("GetMe", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/me", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.GetMe(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	var communityID string
	// 3. Test CreateCommunity
	t.Run("CreateCommunity", func(t *testing.T) {
		body := `{"name":"Golang Flock","description":"Go enthusiasts","city":"Berlin","visibility":"public"}`
		req := httptest.NewRequest("POST", "/communities", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		communityHandler.CreateCommunity(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp models.Community
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Name != "Golang Flock" {
			t.Errorf("expected community name Golang Flock, got %s", resp.Name)
		}
		communityID = resp.ID
	})

	// 4. Test ListCommunities
	t.Run("ListCommunities", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/communities", nil)
		w := httptest.NewRecorder()
		communityHandler.ListCommunities(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp []models.Community
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp) != 1 {
			t.Errorf("expected 1 community, got %d", len(resp))
		}
	})

	var eventID string
	// 5. Test CreateEvent
	t.Run("CreateEvent", func(t *testing.T) {
		// CreateEvent expects a proper starts_at and ends_at. Let's make a JSON body directly
		body := `{"community_id":"` + communityID + `","title":"Go Meetup #1","starts_at":"2027-06-22T15:00:00Z","status":"published"}`

		req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		eventHandler.CreateEvent(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp models.Event
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Title != "Go Meetup #1" {
			t.Errorf("expected event title Go Meetup #1, got %s", resp.Title)
		}
		eventID = resp.ID
	})

	// 6. Test CreateRSVP
	t.Run("CreateRSVP", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/events/"+eventID+"/rsvp", bytes.NewBufferString(`{"status":"confirmed"}`))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		// Set chi URLParam for event ID
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		rsvpHandler.CreateRSVP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var resp models.RSVP
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "confirmed" {
			t.Errorf("expected status confirmed, got %s", resp.Status)
		}
		if resp.UserID != userUUID {
			t.Errorf("expected user id %s, got %s", userUUID, resp.UserID)
		}
	})

	// 7. Test ListEventRSVPs
	t.Run("ListEventRSVPs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/events/"+eventID+"/rsvps", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		rsvpHandler.ListEventRSVPs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp []models.RSVPWithUser
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp) != 1 {
			t.Errorf("expected 1 RSVP, got %d", len(resp))
		}
	})

	var interestID string
	// 8. Test CreateInterest
	t.Run("CreateInterest", func(t *testing.T) {
		body := `{"name":"Programming"}`
		req := httptest.NewRequest("POST", "/interests", strings.NewReader(body))
		w := httptest.NewRecorder()
		interestHandler.CreateInterest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var resp models.Interest
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		interestID = resp.ID
	})

	// 9. Test SetUserInterests
	t.Run("SetUserInterests", func(t *testing.T) {
		body := `{"interests":[{"interest_id":"` + interestID + `","proficiency_level":"expert"}]}`
		req := httptest.NewRequest("POST", "/users/me/interests", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		interestHandler.SetUserInterests(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	// 10. Test GetUserInterests
	t.Run("GetUserInterests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/me/interests", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, clerkID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		interestHandler.GetUserInterests(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp []models.UserInterestDetail
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp) != 1 {
			t.Errorf("expected 1 user interest, got %d", len(resp))
		}
		if resp[0].ProficiencyLevel != "expert" {
			t.Errorf("expected proficiency 'expert', got %s", resp[0].ProficiencyLevel)
		}
	})
}
