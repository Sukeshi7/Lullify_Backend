CREATE TABLE IF NOT EXISTS streams (
    id             UUID         PRIMARY KEY,
    owner_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title          VARCHAR(255) NOT NULL,
    description    TEXT         NOT NULL DEFAULT '',
    mount_point    VARCHAR(100) NOT NULL UNIQUE,
    status         VARCHAR(20)  NOT NULL DEFAULT 'offline',
    listener_count INT          NOT NULL DEFAULT 0,
    started_at     TIMESTAMPTZ,
    ended_at       TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_owner_id ON streams(owner_id);