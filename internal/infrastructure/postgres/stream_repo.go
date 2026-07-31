package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/internal/domain/stream"
)

type StreamRepository struct {
	db *pgxpool.Pool
}

func NewStreamRepository(db *pgxpool.Pool) *StreamRepository {
	return &StreamRepository{db: db}
}

func (r *StreamRepository) Create(ctx context.Context, s *stream.Stream) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO streams (id, owner_id, title, description, mount_point, status, listener_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, s.ID, s.OwnerID, s.Title, s.Description, s.MountPoint, s.Status, s.ListenerCount, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *StreamRepository) FindByID(ctx context.Context, id uuid.UUID) (*stream.Stream, error) {
	s := &stream.Stream{}
	var startedAt, endedAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT id, owner_id, title, description, mount_point, status, listener_count, started_at, ended_at, created_at, updated_at
		FROM streams WHERE id = $1
	`, id).Scan(
		&s.ID, &s.OwnerID, &s.Title, &s.Description, &s.MountPoint,
		&s.Status, &s.ListenerCount, &startedAt, &endedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.StartedAt = startedAt
	s.EndedAt = endedAt
	return s, nil
}

func (r *StreamRepository) FindActive(ctx context.Context) ([]*stream.Stream, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, owner_id, title, description, mount_point, status, listener_count, started_at, ended_at, created_at, updated_at
		FROM streams WHERE status = 'live'
		ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []*stream.Stream
	for rows.Next() {
		s := &stream.Stream{}
		var startedAt, endedAt *time.Time
		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.Title, &s.Description, &s.MountPoint,
			&s.Status, &s.ListenerCount, &startedAt, &endedAt,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.StartedAt = startedAt
		s.EndedAt = endedAt
		streams = append(streams, s)
	}
	return streams, rows.Err()
}

func (r *StreamRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status stream.Status) error {
	now := time.Now().UTC()
	var err error

	switch status {
	case stream.StatusLive:
		_, err = r.db.Exec(ctx, `
			UPDATE streams SET status = $1, started_at = $2, updated_at = $2 WHERE id = $3
		`, status, now, id)
	case stream.StatusEnded:
		_, err = r.db.Exec(ctx, `
			UPDATE streams SET status = $1, ended_at = $2, updated_at = $2 WHERE id = $3
		`, status, now, id)
	default:
		_, err = r.db.Exec(ctx, `
			UPDATE streams SET status = $1, updated_at = $2 WHERE id = $3
		`, status, now, id)
	}
	return err
}

func (r *StreamRepository) IncrementListeners(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE streams SET listener_count = listener_count + 1, updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *StreamRepository) DecrementListeners(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE streams SET listener_count = GREATEST(listener_count - 1, 0), updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *StreamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM streams WHERE id = $1`, id)
	return err
}
