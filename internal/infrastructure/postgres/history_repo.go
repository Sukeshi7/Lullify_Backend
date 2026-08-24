package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/internal/domain/history"
)

type HistoryRepository struct {
	db *pgxpool.Pool
}

func NewHistoryRepository(db *pgxpool.Pool) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Save(ctx context.Context, e *history.Entry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO listening_history (id, user_id, track_title, artist, stream_id, played_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, e.ID, e.UserID, e.TrackTitle, e.Artist, e.StreamID, e.PlayedAt)
	return err
}

func (r *HistoryRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*history.Entry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, track_title, artist, stream_id, played_at
		FROM listening_history WHERE user_id = $1
		ORDER BY played_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*history.Entry
	for rows.Next() {
		e := &history.Entry{}
		if err := rows.Scan(&e.ID, &e.UserID, &e.TrackTitle, &e.Artist, &e.StreamID, &e.PlayedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}
