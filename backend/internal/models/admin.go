package models

// UpdateUserRoleRequest holds the payload to promote/demote user roles.
type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// UsersListResponse represents the paginated response for listing users.
type UsersListResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// PlatformStats contains high-level analytics metrics across the platform.
type PlatformStats struct {
	TotalUsers          int            `json:"total_users"`
	UsersByRole         map[string]int `json:"users_by_role"`
	ActiveUsers24h      int            `json:"active_users_24h"`
	TotalEvents         int            `json:"total_events"`
	EventsByStatus      map[string]int `json:"events_by_status"`
	TotalCommunities    int            `json:"total_communities"`
	CommunitiesByStatus map[string]int `json:"communities_by_status"`
	TotalRSVPs          int            `json:"total_rsvps"`
	RSVPsByStatus       map[string]int `json:"rsvps_by_status"`
}
