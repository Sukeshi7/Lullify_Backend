ALTER TABLE tracks
    ADD COLUMN format       TEXT        NOT NULL DEFAULT 'mp3',
    ADD COLUMN size_bytes   BIGINT      NOT NULL DEFAULT 0,
    ADD COLUMN uploaded_by  UUID,
    ADD COLUMN updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE tracks
    ADD CONSTRAINT tracks_format_check CHECK (format IN ('mp3', 'flac', 'wav')),
    ADD CONSTRAINT tracks_size_bytes_check CHECK (size_bytes >= 0),
    ADD CONSTRAINT tracks_uploaded_by_fkey FOREIGN KEY (uploaded_by) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_tracks_uploaded_by ON tracks(uploaded_by);