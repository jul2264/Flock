package services

import (
	"database/sql"
	"fmt"

	"github.com/jul2264/Flock/backend/internal/models"
)

type InterestService struct {
	db *sql.DB
}

func NewInterestService(db *sql.DB) *InterestService {
	return &InterestService{db: db}
}

func (s *InterestService) ListAll() ([]models.Interest, error) {
	query := `SELECT id, name, parent_id, created_at FROM interests ORDER BY name ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []models.Interest
	for rows.Next() {
		var i models.Interest
		err := rows.Scan(&i.ID, &i.Name, &i.ParentID, &i.CreatedAt)
		if err != nil {
			return nil, err
		}
		interests = append(interests, i)
	}

	return interests, nil
}

func (s *InterestService) Create(req *models.CreateInterestRequest) (*models.Interest, error) {
	query := `
		INSERT INTO interests (name, parent_id)
		VALUES ($1, $2)
		RETURNING id, name, parent_id, created_at
	`
	var i models.Interest
	err := s.db.QueryRow(query, req.Name, req.ParentID).Scan(&i.ID, &i.Name, &i.ParentID, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (s *InterestService) SetUserInterests(clerkID string, inputs []models.UserInterestInput) error {
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user profile not synced")
		}
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing interests
	_, err = tx.Exec(`DELETE FROM user_interests WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Insert new interests
	for _, input := range inputs {
		proficiency := "beginner"
		if input.ProficiencyLevel != "" {
			proficiency = input.ProficiencyLevel
		}

		// Validate proficiency
		switch proficiency {
		case "beginner", "intermediate", "regular", "expert":
			// valid
		default:
			return fmt.Errorf("invalid proficiency level: %s", proficiency)
		}

		_, err = tx.Exec(
			`INSERT INTO user_interests (user_id, interest_id, proficiency_level) VALUES ($1, $2, $3)`,
			userID, input.InterestID, proficiency,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *InterestService) GetUserInterests(clerkID string) ([]models.UserInterestDetail, error) {
	query := `
		SELECT ui.interest_id, ui.proficiency_level, i.name, i.parent_id
		FROM user_interests ui
		JOIN users u ON ui.user_id = u.id
		JOIN interests i ON ui.interest_id = i.id
		WHERE u.clerk_id = $1
		ORDER BY i.name ASC
	`
	rows, err := s.db.Query(query, clerkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []models.UserInterestDetail
	for rows.Next() {
		var d models.UserInterestDetail
		err := rows.Scan(&d.InterestID, &d.ProficiencyLevel, &d.Name, &d.ParentID)
		if err != nil {
			return nil, err
		}
		details = append(details, d)
	}

	return details, nil
}

func (s *InterestService) SetEventInterests(eventID string, interestIDs []string) error {
	// Verify event exists
	var dummy string
	err := s.db.QueryRow(`SELECT id FROM events WHERE id = $1`, eventID).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("event not found")
		}
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing event interests
	_, err = tx.Exec(`DELETE FROM event_interests WHERE event_id = $1`, eventID)
	if err != nil {
		return err
	}

	// Insert new ones
	for _, interestID := range interestIDs {
		_, err = tx.Exec(
			`INSERT INTO event_interests (event_id, interest_id) VALUES ($1, $2)`,
			eventID, interestID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *InterestService) GetEventInterests(eventID string) ([]models.Interest, error) {
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
		SELECT i.id, i.name, i.parent_id, i.created_at
		FROM event_interests ei
		JOIN interests i ON ei.interest_id = i.id
		WHERE ei.event_id = $1
		ORDER BY i.name ASC
	`
	rows, err := s.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []models.Interest
	for rows.Next() {
		var i models.Interest
		err := rows.Scan(&i.ID, &i.Name, &i.ParentID, &i.CreatedAt)
		if err != nil {
			return nil, err
		}
		interests = append(interests, i)
	}

	return interests, nil
}
