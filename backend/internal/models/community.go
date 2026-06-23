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
	Status         string    `json:"status"`
	ImageURL       *string   `json:"image_url"`
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
	ImageURL       *string   `json:"image_url"`
}

type UpdateCommunityRequest struct {
	Name           *string   `json:"name"`
	Description    *string   `json:"description"`
	City           *string   `json:"city"`
	Latitude       *float64  `json:"latitude"`
	Longitude      *float64  `json:"longitude"`
	AgeMin         *int      `json:"age_min"`
	AgeMax         *int      `json:"age_max"`
	MaxMembers     *int      `json:"max_members"`
	IsRecurring    *bool     `json:"is_recurring"`
	RecurrenceRule *string   `json:"recurrence_rule"`
	Visibility     *string   `json:"visibility"`
	Status         *string   `json:"status"`
	ImageURL       *string   `json:"image_url"`
}

type CommunityMember struct {
	UserID    string    `json:"user_id"`
	ClerkID   string    `json:"clerk_id"`
	Email     string    `json:"email"`
	FullName  *string   `json:"full_name"`
	Username  *string   `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type CommunityFilter struct {
	Lat      *float64
	Lng      *float64
	RadiusKm *float64
	AgeMin   *int
	AgeMax   *int
	Sort     string // "newest", "distance", "trending"
	Limit    int
	Offset   int
}
