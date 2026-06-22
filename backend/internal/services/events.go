package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jul2264/Flock/backend/internal/models"
)

type EventService struct {
	db     *sql.DB
	search *SearchService
}

func NewEventService(db *sql.DB, search *SearchService) *EventService {
	return &EventService{db: db, search: search}
}

// Create inserts a new event. The user identified by clerkID becomes the organizer.
func (s *EventService) Create(organizerClerkID string, req *models.CreateEventRequest) (*models.Event, error) {
	// 1. Get user UUID from Clerk ID
	var organizerID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, organizerClerkID).Scan(&organizerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organizer not found")
		}
		return nil, err
	}

	status := "draft"
	if req.Status != nil {
		status = *req.Status
	}

	query := `
		INSERT INTO events (
			organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, age_min, age_max, status
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15
		)
		RETURNING
			id, organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, rsvp_count,
			age_min, age_max, status, created_at, updated_at
	`
	row := s.db.QueryRow(query,
		organizerID, req.CommunityID, req.Title, req.Description, req.VenueName,
		req.VenueAddress, req.GooglePlaceID, req.Latitude, req.Longitude,
		req.StartsAt, req.EndsAt, req.MaxParticipants, req.AgeMin, req.AgeMax, status,
	)
	event, err := scanEvent(row)
	if err != nil {
		return nil, err
	}

	if s.search != nil {
		go func(e *models.Event) {
			if err := s.search.SyncEvent(e); err != nil {
				log.Printf("Error syncing event to Meilisearch: %v", err)
			}
		}(event)
	}

	return event, nil
}

// GetByID fetches a single event.
func (s *EventService) GetByID(id string) (*models.Event, error) {
	query := `
		SELECT
			id, organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, rsvp_count,
			age_min, age_max, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	return scanEvent(row)
}

// List returns upcoming events (basic implementation, you'd add pagination and filters later)
func (s *EventService) List() ([]models.Event, error) {
	query := `
		SELECT
			id, organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, rsvp_count,
			age_min, age_max, status, created_at, updated_at
		FROM events
		WHERE status = 'published' AND starts_at > NOW()
		ORDER BY starts_at ASC
		LIMIT 50
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		err := rows.Scan(
			&e.ID, &e.OrganizerID, &e.CommunityID, &e.Title, &e.Description, &e.VenueName,
			&e.VenueAddress, &e.GooglePlaceID, &e.Latitude, &e.Longitude,
			&e.StartsAt, &e.EndsAt, &e.MaxParticipants, &e.RSVPCount,
			&e.AgeMin, &e.AgeMax, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func scanEvent(row *sql.Row) (*models.Event, error) {
	var e models.Event
	err := row.Scan(
		&e.ID, &e.OrganizerID, &e.CommunityID, &e.Title, &e.Description, &e.VenueName,
		&e.VenueAddress, &e.GooglePlaceID, &e.Latitude, &e.Longitude,
		&e.StartsAt, &e.EndsAt, &e.MaxParticipants, &e.RSVPCount,
		&e.AgeMin, &e.AgeMax, &e.Status, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
