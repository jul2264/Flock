package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type NotificationService struct {
	db        *sql.DB
	fcmClient *messaging.Client
}

func NewNotificationService(db *sql.DB) *NotificationService {
	serviceAccountJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
	var fcmClient *messaging.Client

	if serviceAccountJSON != "" {
		ctx := context.Background()
		opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Printf("Warning: error initializing Firebase App: %v. Falling back to Mock Logger.", err)
		} else {
			client, err := app.Messaging(ctx)
			if err != nil {
				log.Printf("Warning: error initializing FCM messaging client: %v. Falling back to Mock Logger.", err)
			} else {
				fcmClient = client
				log.Println("Firebase Cloud Messaging (FCM) client initialized successfully!")
			}
		}
	} else {
		log.Println("FIREBASE_SERVICE_ACCOUNT_JSON not set. NotificationService running in Mock Mode.")
	}

	return &NotificationService{
		db:        db,
		fcmClient: fcmClient,
	}
}

// RegisterToken stores or updates an FCM device token for a user.
func (s *NotificationService) RegisterToken(ctx context.Context, userID, token string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_fcm_tokens (user_id, token)
		VALUES ($1, $2)
		ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id, created_at = NOW()
	`, userID, token)
	return err
}

// UnregisterToken removes a specific token.
func (s *NotificationService) UnregisterToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_fcm_tokens WHERE token = $1`, token)
	return err
}

// SendToUser sends a notification payload to all active tokens of a specific user.
func (s *NotificationService) SendToUser(ctx context.Context, userID string, title, body string, data map[string]string) error {
	rows, err := s.db.QueryContext(ctx, `SELECT token FROM user_fcm_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tokens = append(tokens, t)
		}
	}

	if len(tokens) == 0 {
		if s.fcmClient == nil {
			log.Printf("[NOTIFICATION MOCK] (No tokens registered) User %s: Title: %s, Body: %s, Data: %v", userID, title, body, data)
		}
		return nil
	}

	return s.SendToTokens(ctx, tokens, title, body, data)
}

// SendToTokens sends notifications to a slice of raw FCM tokens.
func (s *NotificationService) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	if s.fcmClient == nil {
		log.Printf("[NOTIFICATION MOCK] To %d tokens: Title: %s, Body: %s, Data: %v", len(tokens), title, body, data)
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	br, err := s.fcmClient.SendEachForMulticast(ctx, message)
	if err != nil {
		return fmt.Errorf("FCM send multicast error: %w", err)
	}

	if br.FailureCount > 0 {
		for idx, resp := range br.Responses {
			if !resp.Success {
				log.Printf("FCM sending failed for token index %d: %v", idx, resp.Error)
				// Delete invalid/expired token from database
				_, _ = s.db.ExecContext(ctx, `DELETE FROM user_fcm_tokens WHERE token = $1`, tokens[idx])
			}
		}
	}

	return nil
}

// SendCommunityAnnouncement broadcasts an announcement to all members of a community.
func (s *NotificationService) SendCommunityAnnouncement(ctx context.Context, communityID string, title, body string) error {
	query := `
		SELECT t.token 
		FROM user_fcm_tokens t
		JOIN community_members m ON t.user_id = m.user_id
		WHERE m.community_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, communityID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tokens = append(tokens, t)
		}
	}

	if len(tokens) == 0 {
		if s.fcmClient == nil {
			log.Printf("[NOTIFICATION MOCK] Community %s announcement: Title: %s, Body: %s", communityID, title, body)
		}
		return nil
	}

	data := map[string]string{
		"type":         "community_announcement",
		"community_id": communityID,
	}

	return s.SendToTokens(ctx, tokens, title, body, data)
}
