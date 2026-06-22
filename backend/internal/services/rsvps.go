package services

import (
	"database/sql"
	"fmt"

	"github.com/jul2264/Flock/backend/internal/models"
)

type RSVPService struct {
	db *sql.DB
}

func NewRSVPService(db *sql.DB) *RSVPService {
	return &RSVPService{db: db}
}

func (s *RSVPService) Upsert(clerkID string, eventID string, status string) (*models.RSVP, error) {
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user profile not synced")
		}
		return nil, err
	}

	// Verify event exists
	var dummy string
	err = s.db.QueryRow(`SELECT id FROM events WHERE id = $1`, eventID).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, err
	}

	if status == "" {
		status = "confirmed"
	}

	query := `
		INSERT INTO rsvps (event_id, user_id, status, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id, user_id)
		DO UPDATE SET status = EXCLUDED.status, updated_at = NOW()
		RETURNING id, event_id, user_id, status, attended, created_at, updated_at
	`

	var r models.RSVP
	err = s.db.QueryRow(query, eventID, userID, status).Scan(
		&r.ID, &r.EventID, &r.UserID, &r.Status, &r.Attended, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *RSVPService) Cancel(clerkID string, eventID string) error {
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user profile not synced")
		}
		return err
	}

	res, err := s.db.Exec(
		`UPDATE rsvps SET status = 'cancelled', updated_at = NOW() WHERE event_id = $1 AND user_id = $2`,
		eventID, userID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("rsvp not found")
	}

	return nil
}

func (s *RSVPService) ListEventRSVPs(eventID string) ([]models.RSVPWithUser, error) {
	// Verify event exists
	var dummy string
	err := s.db.QueryRow(`SELECT id FROM events WHERE id = $1`, eventID).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, err
	}

	query := `
		SELECT r.id, r.event_id, r.user_id, r.status, r.attended, r.created_at, r.updated_at,
		       u.clerk_id, u.email, u.full_name, u.avatar_url
		FROM rsvps r
		JOIN users u ON r.user_id = u.id
		WHERE r.event_id = $1 AND r.status != 'cancelled'
		ORDER BY r.created_at ASC
	`

	rows, err := s.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rsvps []models.RSVPWithUser
	for rows.Next() {
		var ru models.RSVPWithUser
		err := rows.Scan(
			&ru.ID, &ru.EventID, &ru.UserID, &ru.Status, &ru.Attended, &ru.CreatedAt, &ru.UpdatedAt,
			&ru.ClerkID, &ru.Email, &ru.FullName, &ru.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		rsvps = append(rsvps, ru)
	}

	return rsvps, nil
}

func (s *RSVPService) ListUserRSVPs(clerkID string) ([]models.RSVPWithEvent, error) {
	query := `
		SELECT r.id, r.event_id, r.user_id, r.status, r.attended, r.created_at, r.updated_at,
		       e.title, e.start_date, e.end_date, e.location
		FROM rsvps r
		JOIN users u ON r.user_id = u.id
		JOIN events e ON r.event_id = e.id
		WHERE u.clerk_id = $1
		ORDER BY e.start_date ASC
	`

	rows, err := s.db.Query(query, clerkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rsvps []models.RSVPWithEvent
	for rows.Next() {
		var re models.RSVPWithEvent
		err := rows.Scan(
			&re.ID, &re.EventID, &re.UserID, &re.Status, &re.Attended, &re.CreatedAt, &re.UpdatedAt,
			&re.Title, &re.StartDate, &re.EndDate, &re.Location,
		)
		if err != nil {
			return nil, err
		}
		rsvps = append(rsvps, re)
	}

	return rsvps, nil
}
