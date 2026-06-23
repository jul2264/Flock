package models

import "time"

// Event mirrors the `events` table.
type Event struct {
	ID             string     `json:"id"`
	OrganizerID    string     `json:"organizer_id"`
	CommunityID    *string    `json:"community_id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	VenueName      *string    `json:"venue_name"`
	VenueAddress   *string    `json:"venue_address"`
	GooglePlaceID  *string    `json:"google_place_id"`
	Latitude       *float64   `json:"latitude"`
	Longitude      *float64   `json:"longitude"`
	StartsAt       time.Time  `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	MaxParticipants *int       `json:"max_participants"`
	RSVPCount      int        `json:"rsvp_count"`
	AgeMin         *int       `json:"age_min"`
	AgeMax         *int       `json:"age_max"`
	Status         string     `json:"status"`
	BannerURL      *string    `json:"banner_url"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateEventRequest struct {
	CommunityID    *string    `json:"community_id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	VenueName      *string    `json:"venue_name"`
	VenueAddress   *string    `json:"venue_address"`
	GooglePlaceID  *string    `json:"google_place_id"`
	Latitude       *float64   `json:"latitude"`
	Longitude      *float64   `json:"longitude"`
	StartsAt       time.Time  `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	MaxParticipants *int       `json:"max_participants"`
	AgeMin         *int       `json:"age_min"`
	AgeMax         *int       `json:"age_max"`
	Status         *string    `json:"status"` // Defaults to "draft"
	BannerURL      *string    `json:"banner_url"`
}

type UpdateEventRequest struct {
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	VenueName      *string    `json:"venue_name"`
	VenueAddress   *string    `json:"venue_address"`
	GooglePlaceID  *string    `json:"google_place_id"`
	Latitude       *float64   `json:"latitude"`
	Longitude      *float64   `json:"longitude"`
	StartsAt       *time.Time `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	MaxParticipants *int       `json:"max_participants"`
	AgeMin         *int       `json:"age_min"`
	AgeMax         *int       `json:"age_max"`
	Status         *string    `json:"status"`
	BannerURL      *string    `json:"banner_url"`
}

type EventFilter struct {
	Lat        *float64
	Lng        *float64
	RadiusKm   *float64
	InterestID *string
	AgeMin     *int
	AgeMax     *int
	From       *time.Time
	To         *time.Time
	Sort       string // "upcoming", "distance", "trending"
	Limit      int
	Offset     int
}
