package models

type RegisterFCMTokenRequest struct {
	Token string `json:"token"`
}

type CommunityAnnouncementRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
