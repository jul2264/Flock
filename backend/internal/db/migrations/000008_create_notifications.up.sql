CREATE TABLE IF NOT EXISTS user_fcm_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS event_reminders_sent (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    reminder_type VARCHAR(10) NOT NULL CHECK (reminder_type IN ('24h', '1h')),
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (event_id, reminder_type)
);
