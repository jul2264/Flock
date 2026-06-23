package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jul2264/Flock/backend/internal/models"
)

type AdminService struct {
	db *sql.DB
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

// ListUsers retrieves a paginated list of users along with the total count.
func (s *AdminService) ListUsers(page, limit int) (*models.UsersListResponse, error) {
	offset := (page - 1) * limit

	var total int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT
			id, clerk_id, email, full_name, username, avatar_url,
			date_of_birth, city, neighborhood, latitude, longitude,
			search_radius, role, last_seen_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}

	if users == nil {
		users = []models.User{}
	}

	return &models.UsersListResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// UpdateUserRole updates the role of a user identified by their internal database ID.
func (s *AdminService) UpdateUserRole(userID string, role string) (*models.User, error) {
	if role != "user" && role != "organizer" && role != "admin" {
		return nil, fmt.Errorf("invalid role: must be user, organizer, or admin")
	}

	query := `
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING
			id, clerk_id, email, full_name, username, avatar_url,
			date_of_birth, city, neighborhood, latitude, longitude,
			search_radius, role, last_seen_at, created_at, updated_at
	`
	row := s.db.QueryRow(query, role, userID)
	u, err := scanUserFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return u, nil
}

// GetStats compiles platform-wide analytics statistics.
func (s *AdminService) GetStats() (*models.PlatformStats, error) {
	stats := &models.PlatformStats{
		UsersByRole:         make(map[string]int),
		EventsByStatus:      make(map[string]int),
		CommunitiesByStatus: make(map[string]int),
		RSVPsByStatus:       make(map[string]int),
	}

	// 1. Total users
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// 2. Users by role
	uRows, err := s.db.Query(`SELECT role, COUNT(*) FROM users GROUP BY role`)
	if err != nil {
		return nil, err
	}
	defer uRows.Close()
	for uRows.Next() {
		var role string
		var count int
		if err := uRows.Scan(&role, &count); err == nil {
			stats.UsersByRole[role] = count
		}
	}

	// 3. Active users (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	err = s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE last_seen_at >= $1`, yesterday).Scan(&stats.ActiveUsers24h)
	if err != nil {
		return nil, err
	}

	// 4. Total events
	err = s.db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&stats.TotalEvents)
	if err != nil {
		return nil, err
	}

	// 5. Events by status
	eRows, err := s.db.Query(`SELECT status, COUNT(*) FROM events GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer eRows.Close()
	for eRows.Next() {
		var status string
		var count int
		if err := eRows.Scan(&status, &count); err == nil {
			stats.EventsByStatus[status] = count
		}
	}

	// 6. Total communities
	err = s.db.QueryRow(`SELECT COUNT(*) FROM communities`).Scan(&stats.TotalCommunities)
	if err != nil {
		return nil, err
	}

	// 7. Communities by status
	cRows, err := s.db.Query(`SELECT status, COUNT(*) FROM communities GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()
	for cRows.Next() {
		var status string
		var count int
		if err := cRows.Scan(&status, &count); err == nil {
			stats.CommunitiesByStatus[status] = count
		}
	}

	// 8. Total RSVPs
	err = s.db.QueryRow(`SELECT COUNT(*) FROM rsvps`).Scan(&stats.TotalRSVPs)
	if err != nil {
		return nil, err
	}

	// 9. RSVPs by status
	rRows, err := s.db.Query(`SELECT status, COUNT(*) FROM rsvps GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rRows.Close()
	for rRows.Next() {
		var status string
		var count int
		if err := rRows.Scan(&status, &count); err == nil {
			stats.RSVPsByStatus[status] = count
		}
	}

	return stats, nil
}

// Helpers for scanning users
func scanUserFromRow(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.ClerkID, &u.Email, &u.FullName, &u.Username, &u.AvatarURL,
		&u.DateOfBirth, &u.City, &u.Neighborhood, &u.Latitude, &u.Longitude,
		&u.SearchRadius, &u.Role, &u.LastSeenAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func scanUserFromRows(rows *sql.Rows) (*models.User, error) {
	var u models.User
	err := rows.Scan(
		&u.ID, &u.ClerkID, &u.Email, &u.FullName, &u.Username, &u.AvatarURL,
		&u.DateOfBirth, &u.City, &u.Neighborhood, &u.Latitude, &u.Longitude,
		&u.SearchRadius, &u.Role, &u.LastSeenAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
