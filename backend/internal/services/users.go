package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jul2264/Flock/backend/internal/models"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// Upsert creates the user row if it doesn't exist, or updates profile fields
// and refreshes last_seen_at if it does. Called on every sign-in.
func (s *UserService) Upsert(clerkID, email, fullName, avatarURL string) (*models.User, error) {
	query := `
		INSERT INTO users (clerk_id, email, full_name, avatar_url, last_seen_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (clerk_id) DO UPDATE SET
			email        = EXCLUDED.email,
			full_name    = EXCLUDED.full_name,
			avatar_url   = EXCLUDED.avatar_url,
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at   = NOW()
		RETURNING
			id, clerk_id, email, full_name, username, avatar_url,
			date_of_birth, city, neighborhood, latitude, longitude,
			search_radius, role, last_seen_at, created_at, updated_at
	`
	row := s.db.QueryRow(query, clerkID, email, fullName, avatarURL, time.Now())
	return scanUser(row)
}

// GetByClerkID fetches a user by their Clerk subject ID.
func (s *UserService) GetByClerkID(clerkID string) (*models.User, error) {
	query := `
		SELECT
			id, clerk_id, email, full_name, username, avatar_url,
			date_of_birth, city, neighborhood, latitude, longitude,
			search_radius, role, last_seen_at, created_at, updated_at
		FROM users
		WHERE clerk_id = $1
	`
	row := s.db.QueryRow(query, clerkID)
	return scanUser(row)
}

// Update applies a partial update to the user row. Only non-nil fields in
// req are written; at minimum updated_at is always refreshed.
func (s *UserService) Update(clerkID string, req *models.UpdateUserRequest) (*models.User, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	idx := 1

	if req.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", idx))
		args = append(args, *req.FullName)
		idx++
	}
	if req.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", idx))
		args = append(args, *req.Username)
		idx++
	}
	if req.AvatarURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", idx))
		args = append(args, *req.AvatarURL)
		idx++
	}
	if req.City != nil {
		setClauses = append(setClauses, fmt.Sprintf("city = $%d", idx))
		args = append(args, *req.City)
		idx++
	}
	if req.Neighborhood != nil {
		setClauses = append(setClauses, fmt.Sprintf("neighborhood = $%d", idx))
		args = append(args, *req.Neighborhood)
		idx++
	}
	if req.Latitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("latitude = $%d", idx))
		args = append(args, *req.Latitude)
		idx++
	}
	if req.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", idx))
		args = append(args, *req.Longitude)
		idx++
	}
	if req.SearchRadius != nil {
		setClauses = append(setClauses, fmt.Sprintf("search_radius = $%d", idx))
		args = append(args, *req.SearchRadius)
		idx++
	}

	// $idx is now the placeholder for clerk_id
	args = append(args, clerkID)
	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE clerk_id = $%d
		RETURNING
			id, clerk_id, email, full_name, username, avatar_url,
			date_of_birth, city, neighborhood, latitude, longitude,
			search_radius, role, last_seen_at, created_at, updated_at
	`, strings.Join(setClauses, ", "), idx)

	row := s.db.QueryRow(query, args...)
	return scanUser(row)
}

// scanUser maps a single DB row into a models.User.
func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.ClerkID, &u.Email, &u.FullName, &u.Username, &u.AvatarURL,
		&u.DateOfBirth, &u.City, &u.Neighborhood, &u.Latitude, &u.Longitude,
		&u.SearchRadius, &u.Role, &u.LastSeenAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
