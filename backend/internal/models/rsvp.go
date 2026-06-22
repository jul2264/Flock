package models

import "time"

// RSVP mirrors the `rsvps` table.
type RSVP struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Attended  bool      `json:"attended"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRSVPRequest struct {
	Status string `json:"status"` // optional, e.g., 'confirmed', 'pending', 'cancelled', 'waitlisted'
}

type RSVPWithUser struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Attended  bool      `json:"attended"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ClerkID   string    `json:"clerk_id"`
	Email     string    `json:"email"`
	FullName  *string   `json:"full_name"`
	AvatarURL *string   `json:"avatar_url"`
}

type RSVPWithEvent struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Attended  bool      `json:"attended"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Location  *string   `json:"location"`
}
