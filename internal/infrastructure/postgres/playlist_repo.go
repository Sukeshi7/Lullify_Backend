package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/internal/domain/playlist"
)

type PlaylistRepository struct {
	db *pgxpool.Pool
}

func NewPlaylistRepository(db *pgxpool.Pool) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(ctx context.Context, p *playlist.Playlist) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO playlists (id, owner_id, title, description, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, p.ID, p.OwnerID, p.Title, p.Description, p.IsPublic, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PlaylistRepository) FindByID(ctx context.Context, id uuid.UUID) (*playlist.Playlist, error) {
	p := &playlist.Playlist{}
	err := r.db.QueryRow(ctx, `
		SELECT id, owner_id, title, description, is_public, created_at, updated_at
		FROM playlists WHERE id = $1
	`, id).Scan(&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.IsPublic, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *PlaylistRepository) FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]*playlist.Playlist, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, owner_id, title, description, is_public, created_at, updated_at
		FROM playlists WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*playlist.Playlist
	for rows.Next() {
		p := &playlist.Playlist{}
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.IsPublic, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *PlaylistRepository) Update(ctx context.Context, p *playlist.Playlist) error {
	_, err := r.db.Exec(ctx, `
		UPDATE playlists
		SET title = $1, description = $2, is_public = $3, updated_at = $4
		WHERE id = $5
	`, p.Title, p.Description, p.IsPublic, p.UpdatedAt, p.ID)
	return err
}

func (r *PlaylistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM playlists WHERE id = $1`, id)
	return err
}
