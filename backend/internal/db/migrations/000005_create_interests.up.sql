CREATE TABLE IF NOT EXISTS interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES interests(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_interests (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    proficiency_level VARCHAR(20) DEFAULT 'beginner' CHECK (proficiency_level IN ('beginner', 'intermediate', 'regular', 'expert')),
    PRIMARY KEY (user_id, interest_id)
);

CREATE TABLE IF NOT EXISTS event_interests (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, interest_id)
);