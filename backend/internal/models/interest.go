package models

import "time"

// Interest mirrors the `interests` table.
type Interest struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserInterest mirrors the `user_interests` table.
type UserInterest struct {
	UserID           string `json:"user_id"`
	InterestID       string `json:"interest_id"`
	ProficiencyLevel string `json:"proficiency_level"`
}

// UserInterestInput represents a single interest in the request body.
type UserInterestInput struct {
	InterestID       string `json:"interest_id"`
	ProficiencyLevel string `json:"proficiency_level"` // e.g., 'beginner', 'intermediate', 'regular', 'expert'
}

// UserInterestDetail returns user interest metadata.
type UserInterestDetail struct {
	InterestID       string  `json:"interest_id"`
	Name             string  `json:"name"`
	ParentID         *string `json:"parent_id"`
	ProficiencyLevel string  `json:"proficiency_level"`
}

type CreateInterestRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"` // optional
}

type SetUserInterestsRequest struct {
	Interests []UserInterestInput `json:"interests"`
}

type SetEventInterestsRequest struct {
	InterestIDs []string `json:"interest_ids"`
}
