package models

import "time"

// Community mirrors the `communities` table.
type Community struct {
	ID             string    `json:"id"`
	OwnerID        string    `json:"owner_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	City           *string   `json:"city"`
	Latitude       *float64  `json:"latitude"`
	Longitude      *float64  `json:"longitude"`
	AgeMin         *int      `json:"age_min"`
	AgeMax         *int      `json:"age_max"`
	MaxMembers     *int      `json:"max_members"`
	IsRecurring    bool      `json:"is_recurring"`
	RecurrenceRule *string   `json:"recurrence_rule"`
	Visibility     string    `json:"visibility"`
	MemberCount    int       `json:"member_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateCommunityRequest struct {
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	City           *string   `json:"city"`
	Latitude       *float64  `json:"latitude"`
	Longitude      *float64  `json:"longitude"`
	AgeMin         *int      `json:"age_min"`
	AgeMax         *int      `json:"age_max"`
	MaxMembers     *int      `json:"max_members"`
	IsRecurring    *bool     `json:"is_recurring"`
	RecurrenceRule *string   `json:"recurrence_rule"`
	Visibility     *string   `json:"visibility"` // default 'public'
}
