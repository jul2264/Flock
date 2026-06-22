package models

import "time"

// User mirrors the `users` table.
type User struct {
	ID           string     `json:"id"`
	ClerkID      string     `json:"clerk_id"`
	Email        string     `json:"email"`
	FullName     *string    `json:"full_name"`
	Username     *string    `json:"username"`
	AvatarURL    *string    `json:"avatar_url"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	City         *string    `json:"city"`
	Neighborhood *string    `json:"neighborhood"`
	Latitude     *float64   `json:"latitude"`
	Longitude    *float64   `json:"longitude"`
	SearchRadius int        `json:"search_radius"`
	Role         string     `json:"role"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SyncUserRequest is the body for POST /users/sync.
// The clerk_id comes from the JWT; only profile data is sent here.
type SyncUserRequest struct {
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateUserRequest is the body for PATCH /users/me.
// All fields are optional — only non-nil fields are applied.
type UpdateUserRequest struct {
	FullName     *string  `json:"full_name"`
	Username     *string  `json:"username"`
	AvatarURL    *string  `json:"avatar_url"`
	City         *string  `json:"city"`
	Neighborhood *string  `json:"neighborhood"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	SearchRadius *int     `json:"search_radius"`
}
