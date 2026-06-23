package models

type UploadRequest struct {
	ContentType string `json:"content_type"`
}

type UploadResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}
