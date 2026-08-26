package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/internal/domain/favorite"
)

type FavoriteRepository struct {
	db *pgxpool.Pool
}

func NewFavoriteRepository(db *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Save(ctx context.Context, f *favorite.Favorite) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO favorites (id, user_id, stream_id, created_at)
		VALUES ($1, $2, $3, $4)
	`, f.ID, f.UserID, f.StreamID, f.CreatedAt)

	// 23505 = unique_violation Postgres → déjà en favoris
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return favorite.ErrAlreadyFavorited
	}
	return err
}

func (r *FavoriteRepository) Delete(ctx context.Context, userID, streamID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM favorites WHERE user_id = $1 AND stream_id = $2
	`, userID, streamID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return favorite.ErrFavoriteNotFound
	}
	return nil
}

func (r *FavoriteRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*favorite.Favorite, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, stream_id, created_at
		FROM favorites WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*favorite.Favorite
	for rows.Next() {
		f := &favorite.Favorite{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.StreamID, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return list, nil
}