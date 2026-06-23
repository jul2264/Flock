package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jul2264/Flock/backend/internal/models"
)

type EventService struct {
	db           *sql.DB
	search       *SearchService
	notification *NotificationService
}

func NewEventService(db *sql.DB, search *SearchService, notification *NotificationService) *EventService {
	return &EventService{db: db, search: search, notification: notification}
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
			starts_at, ends_at, max_participants, age_min, age_max, status, banner_url
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16
		)
		RETURNING
			id, organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, rsvp_count,
			age_min, age_max, status, banner_url, created_at, updated_at
	`
	row := s.db.QueryRow(query,
		organizerID, req.CommunityID, req.Title, req.Description, req.VenueName,
		req.VenueAddress, req.GooglePlaceID, req.Latitude, req.Longitude,
		req.StartsAt, req.EndsAt, req.MaxParticipants, req.AgeMin, req.AgeMax, status, req.BannerURL,
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

	if event.Status == "published" {
		go s.sendNearbyAlerts(event)
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
			age_min, age_max, status, banner_url, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	return scanEvent(row)
}

// List returns upcoming events filtered, sorted, and paginated based on the provided parameters.
func (s *EventService) List(filter *models.EventFilter) ([]models.Event, error) {
	var selectCols = `id, organizer_id, community_id, title, description, venue_name,
		venue_address, google_place_id, latitude, longitude,
		starts_at, ends_at, max_participants, rsvp_count,
		age_min, age_max, status, banner_url, created_at, updated_at`

	args := []interface{}{}
	idx := 1

	// Base conditions
	whereClauses := []string{"status = 'published'", "starts_at > NOW()"}

	// Geo-filtering
	var selectQuery string
	if filter.Lat != nil && filter.Lng != nil && filter.RadiusKm != nil {
		selectQuery = fmt.Sprintf(`
			SELECT %s FROM (
				SELECT *, (6371 * acos(
					cos(radians($%d)) * cos(radians(latitude)) * cos(radians(longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(latitude))
				)) AS distance
				FROM events
			) sub
		`, selectCols, idx, idx+1, idx)
		args = append(args, *filter.Lat, *filter.Lng)
		idx += 2

		whereClauses = append(whereClauses, fmt.Sprintf("distance <= $%d", idx))
		args = append(args, *filter.RadiusKm)
		idx++

		whereClauses = append(whereClauses, "latitude IS NOT NULL", "longitude IS NOT NULL")
	} else {
		selectQuery = fmt.Sprintf("SELECT %s FROM events", selectCols)
	}

	// Interest filter
	if filter.InterestID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("EXISTS (SELECT 1 FROM event_interests WHERE event_id = id AND interest_id = $%d)", idx))
		args = append(args, *filter.InterestID)
		idx++
	}

	// Age filters
	if filter.AgeMin != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("age_min >= $%d", idx))
		args = append(args, *filter.AgeMin)
		idx++
	}
	if filter.AgeMax != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("age_max <= $%d", idx))
		args = append(args, *filter.AgeMax)
		idx++
	}

	// Date range filters
	if filter.From != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("starts_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("starts_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}

	// Construct WHERE clause
	var whereClause string
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sort logic
	var orderClause string
	switch filter.Sort {
	case "distance":
		if filter.Lat != nil && filter.Lng != nil && filter.RadiusKm != nil {
			orderClause = " ORDER BY distance ASC"
		} else {
			orderClause = " ORDER BY starts_at ASC"
		}
	case "trending":
		orderClause = " ORDER BY rsvp_count DESC, starts_at ASC"
	case "upcoming":
		fallthrough
	default:
		orderClause = " ORDER BY starts_at ASC"
	}

	// Pagination limits
	limit := 20
	if filter.Limit > 0 {
		limit = filter.Limit
		if limit > 50 {
			limit = 50
		}
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	query := selectQuery + whereClause + orderClause + paginationClause

	rows, err := s.db.Query(query, args...)
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
			&e.AgeMin, &e.AgeMax, &e.Status, &e.BannerURL, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

// GetEventOwner fetches the organizer's database UUID for the given event ID.
func (s *EventService) GetEventOwner(eventID string) (string, error) {
	var organizerID string
	err := s.db.QueryRow(`SELECT organizer_id FROM events WHERE id = $1`, eventID).Scan(&organizerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("event not found")
		}
		return "", err
	}
	return organizerID, nil
}

// Update partially updates an event after enforcing ownership/admin checks.
func (s *EventService) Update(eventID string, userClerkID string, userRole string, req *models.UpdateEventRequest) (*models.Event, error) {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, userClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 2. Fetch the event organizer to check ownership
	organizerID, err := s.GetEventOwner(eventID)
	if err != nil {
		return nil, err
	}

	var prevStatus string
	_ = s.db.QueryRow(`SELECT status FROM events WHERE id = $1`, eventID).Scan(&prevStatus)

	// 3. Ownership check: organizer or admin
	if organizerID != userID && userRole != "admin" {
		return nil, fmt.Errorf("forbidden")
	}

	// 4. Construct dynamic UPDATE query
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	idx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", idx))
		args = append(args, *req.Title)
		idx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", idx))
		args = append(args, req.Description)
		idx++
	}
	if req.VenueName != nil {
		setClauses = append(setClauses, fmt.Sprintf("venue_name = $%d", idx))
		args = append(args, req.VenueName)
		idx++
	}
	if req.VenueAddress != nil {
		setClauses = append(setClauses, fmt.Sprintf("venue_address = $%d", idx))
		args = append(args, req.VenueAddress)
		idx++
	}
	if req.GooglePlaceID != nil {
		setClauses = append(setClauses, fmt.Sprintf("google_place_id = $%d", idx))
		args = append(args, req.GooglePlaceID)
		idx++
	}
	if req.Latitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("latitude = $%d", idx))
		args = append(args, req.Latitude)
		idx++
	}
	if req.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", idx))
		args = append(args, req.Longitude)
		idx++
	}
	if req.StartsAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("starts_at = $%d", idx))
		args = append(args, *req.StartsAt)
		idx++
	}
	if req.EndsAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("ends_at = $%d", idx))
		args = append(args, req.EndsAt)
		idx++
	}
	if req.MaxParticipants != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_participants = $%d", idx))
		args = append(args, req.MaxParticipants)
		idx++
	}
	if req.AgeMin != nil {
		setClauses = append(setClauses, fmt.Sprintf("age_min = $%d", idx))
		args = append(args, req.AgeMin)
		idx++
	}
	if req.AgeMax != nil {
		setClauses = append(setClauses, fmt.Sprintf("age_max = $%d", idx))
		args = append(args, req.AgeMax)
		idx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, *req.Status)
		idx++
	}
	if req.BannerURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("banner_url = $%d", idx))
		args = append(args, req.BannerURL)
		idx++
	}

	// Add eventID as the final argument
	args = append(args, eventID)
	query := fmt.Sprintf(`
		UPDATE events
		SET %s
		WHERE id = $%d
		RETURNING
			id, organizer_id, community_id, title, description, venue_name,
			venue_address, google_place_id, latitude, longitude,
			starts_at, ends_at, max_participants, rsvp_count,
			age_min, age_max, status, banner_url, created_at, updated_at
	`, strings.Join(setClauses, ", "), idx)

	row := s.db.QueryRow(query, args...)
	updatedEvent, err := scanEvent(row)
	if err != nil {
		return nil, err
	}

	if s.search != nil && updatedEvent != nil {
		go func(e *models.Event) {
			if err := s.search.SyncEvent(e); err != nil {
				log.Printf("Error syncing event to Meilisearch: %v", err)
			}
		}(updatedEvent)
	}

	if updatedEvent != nil && updatedEvent.Status == "published" && prevStatus != "published" {
		go s.sendNearbyAlerts(updatedEvent)
	}

	return updatedEvent, nil
}

// Delete soft-deletes an event by updating its status to 'cancelled' after enforcing ownership/admin checks.
func (s *EventService) Delete(eventID string, userClerkID string, userRole string) error {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, userClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// 2. Fetch the event organizer to check ownership
	organizerID, err := s.GetEventOwner(eventID)
	if err != nil {
		return err
	}

	// 3. Ownership check: organizer or admin
	if organizerID != userID && userRole != "admin" {
		return fmt.Errorf("forbidden")
	}

	// 4. Soft delete by updating status to 'cancelled'
	_, err = s.db.Exec(`UPDATE events SET status = 'cancelled', updated_at = NOW() WHERE id = $1`, eventID)
	if err != nil {
		return err
	}

	// 5. Update Meilisearch asynchronously
	if s.search != nil {
		go func() {
			updatedEvent, err := s.GetByID(eventID)
			if err != nil {
				log.Printf("Error fetching event for Meilisearch sync: %v", err)
				return
			}
			if updatedEvent != nil {
				if err := s.search.SyncEvent(updatedEvent); err != nil {
					log.Printf("Error syncing event to Meilisearch: %v", err)
				}
			}
		}()
	}

	return nil
}

func scanEvent(row *sql.Row) (*models.Event, error) {
	var e models.Event
	err := row.Scan(
		&e.ID, &e.OrganizerID, &e.CommunityID, &e.Title, &e.Description, &e.VenueName,
		&e.VenueAddress, &e.GooglePlaceID, &e.Latitude, &e.Longitude,
		&e.StartsAt, &e.EndsAt, &e.MaxParticipants, &e.RSVPCount,
		&e.AgeMin, &e.AgeMax, &e.Status, &e.BannerURL, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EventService) sendNearbyAlerts(event *models.Event) {
	if s.notification == nil {
		return
	}

	if event.Latitude == nil || event.Longitude == nil {
		return
	}

	query := `
		SELECT DISTINCT t.token 
		FROM user_fcm_tokens t
		JOIN users u ON t.user_id = u.id
		JOIN user_interests ui ON u.id = ui.user_id
		JOIN event_interests ei ON ui.interest_id = ei.interest_id
		WHERE ei.event_id = $1
		  AND u.latitude IS NOT NULL 
		  AND u.longitude IS NOT NULL
		  AND (6371 * acos(
		      cos(radians($2)) * cos(radians(u.latitude)) * cos(radians(u.longitude) - radians($3)) +
		      sin(radians($2)) * sin(radians(u.latitude))
		  )) <= u.search_radius
	`
	rows, err := s.db.Query(query, event.ID, *event.Latitude, *event.Longitude)
	if err != nil {
		log.Printf("Error querying nearby FCM tokens: %v", err)
		return
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err == nil {
			tokens = append(tokens, token)
		}
	}

	if len(tokens) == 0 {
		return
	}

	title := "New Event Nearby"
	body := fmt.Sprintf("A new event matching your interests is happening nearby: '%s'", event.Title)
	data := map[string]string{
		"type":     "nearby_event",
		"event_id": event.ID,
	}

	_ = s.notification.SendToTokens(context.Background(), tokens, title, body, data)
}
