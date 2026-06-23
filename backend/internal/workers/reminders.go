package workers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jul2264/Flock/backend/internal/services"
)

type ReminderWorker struct {
	db           *sql.DB
	notification *services.NotificationService
}

func NewReminderWorker(db *sql.DB, notification *services.NotificationService) *ReminderWorker {
	return &ReminderWorker{
		db:           db,
		notification: notification,
	}
}

// Start runs the periodic ticker checks in a blocking loop (should be run in a goroutine).
func (w *ReminderWorker) Start(ctx context.Context) {
	log.Println("⏰ Event Reminder background worker started!")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run initial check immediately
	w.runChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Event Reminder background worker stopping...")
			return
		case <-ticker.C:
			w.runChecks(ctx)
		}
	}
}

func (w *ReminderWorker) runChecks(ctx context.Context) {
	// 1. Process 24-hour reminders
	w.processReminders(ctx, "24h", "23 hours 55 minutes", "24 hours 5 minutes", "starts in 24 hours")

	// 2. Process 1-hour reminders
	w.processReminders(ctx, "1h", "55 minutes", "1 hour 5 minutes", "starts in 1 hour")
}

func (w *ReminderWorker) processReminders(ctx context.Context, reminderType, minInterval, maxInterval, timeText string) {
	query := fmt.Sprintf(`
		SELECT id, title, COALESCE(venue_name, 'the venue') 
		FROM events
		WHERE status = 'published'
		  AND starts_at >= NOW() + INTERVAL '%s'
		  AND starts_at <= NOW() + INTERVAL '%s'
		  AND NOT EXISTS (
		      SELECT 1 FROM event_reminders_sent 
		      WHERE event_id = id AND reminder_type = $1
		  )
	`, minInterval, maxInterval)

	rows, err := w.db.QueryContext(ctx, query, reminderType)
	if err != nil {
		log.Printf("ReminderWorker error querying events for %s: %v", reminderType, err)
		return
	}
	defer rows.Close()

	type eventInfo struct {
		id    string
		title string
		venue string
	}
	var events []eventInfo
	for rows.Next() {
		var ev eventInfo
		if err := rows.Scan(&ev.id, &ev.title, &ev.venue); err == nil {
			events = append(events, ev)
		}
	}

	for _, ev := range events {
		tokenQuery := `
			SELECT DISTINCT t.token 
			FROM user_fcm_tokens t
			JOIN rsvps r ON t.user_id = r.user_id
			WHERE r.event_id = $1 AND r.status = 'confirmed'
		`
		tRows, err := w.db.QueryContext(ctx, tokenQuery, ev.id)
		if err != nil {
			log.Printf("ReminderWorker error querying tokens for event %s: %v", ev.id, err)
			continue
		}

		var tokens []string
		for tRows.Next() {
			var token string
			if err := tRows.Scan(&token); err == nil {
				tokens = append(tokens, token)
			}
		}
		tRows.Close()

		// Record the reminder as sent first to prevent duplicate delivery
		_, err = w.db.ExecContext(ctx, `
			INSERT INTO event_reminders_sent (event_id, reminder_type)
			VALUES ($1, $2)
			ON CONFLICT (event_id, reminder_type) DO NOTHING
		`, ev.id, reminderType)
		if err != nil {
			log.Printf("ReminderWorker error marking reminder %s sent for event %s: %v", reminderType, ev.id, err)
			continue
		}

		if len(tokens) == 0 {
			continue
		}

		title := "Upcoming Event Reminder"
		body := fmt.Sprintf("Reminder: Event '%s' %s at %s!", ev.title, timeText, ev.venue)
		data := map[string]string{
			"type":     "event_reminder",
			"event_id": ev.id,
			"reminder": reminderType,
		}

		go func(tkns []string, t, b string, d map[string]string) {
			_ = w.notification.SendToTokens(context.Background(), tkns, t, b, d)
		}(tokens, title, body, data)
	}
}
