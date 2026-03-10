CREATE TABLE IF NOT EXISTS competition_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    host_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    mode TEXT NOT NULL,
    question_count INT NOT NULL CHECK (question_count > 0 AND question_count <= 100),
    difficulty_policy TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'active', 'completed', 'closed')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS competition_room_members (
    room_id UUID NOT NULL REFERENCES competition_rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_host BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_competition_rooms_host_created
    ON competition_rooms (host_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_competition_room_members_user_joined
    ON competition_room_members (user_id, joined_at DESC);

