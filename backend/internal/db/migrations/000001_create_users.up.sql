CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255),
    username VARCHAR(100) UNIQUE,
    avatar_url TEXT,
    date_of_birth DATE,
    city VARCHAR(100),
    neighborhood VARCHAR(100),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    search_radius INTEGER DEFAULT 10,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'organizer', 'admin')),
    last_seen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);