package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/internal/domain/playlist"
)

type TrackRepository struct {
	db *pgxpool.Pool
}

func NewTrackRepository(db *pgxpool.Pool) *TrackRepository {
	return &TrackRepository{db: db}
}

func (r *TrackRepository) Create(ctx context.Context, t *playlist.Track) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO tracks (
			id, playlist_id, title, artist, file_path,
			format, size_bytes, duration, position,
			uploaded_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, t.ID, t.PlaylistID, t.Title, t.Artist, t.FilePath,
		string(t.Format), t.SizeBytes, t.Duration, t.Position,
		t.UploadedBy, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TrackRepository) FindByID(ctx context.Context, id uuid.UUID) (*playlist.Track, error) {
	t := &playlist.Track{}
	var format string
	err := r.db.QueryRow(ctx, `
		SELECT id, playlist_id, title, artist, file_path,
		       format, size_bytes, duration, position,
		       uploaded_by, created_at, updated_at
		FROM tracks WHERE id = $1
	`, id).Scan(
		&t.ID, &t.PlaylistID, &t.Title, &t.Artist, &t.FilePath,
		&format, &t.SizeBytes, &t.Duration, &t.Position,
		&t.UploadedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	t.Format = playlist.Format(format)
	return t, err
}

func (r *TrackRepository) FindByPlaylist(ctx context.Context, playlistID uuid.UUID) ([]*playlist.Track, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, playlist_id, title, artist, file_path,
		       format, size_bytes, duration, position,
		       uploaded_by, created_at, updated_at
		FROM tracks WHERE playlist_id = $1
		ORDER BY position ASC
	`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*playlist.Track
	for rows.Next() {
		t := &playlist.Track{}
		var format string
		if err := rows.Scan(
			&t.ID, &t.PlaylistID, &t.Title, &t.Artist, &t.FilePath,
			&format, &t.SizeBytes, &t.Duration, &t.Position,
			&t.UploadedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		t.Format = playlist.Format(format)
		list = append(list, t)
	}
	return list, rows.Err()
}

// FindLatestByUploader retourne la dernière track uploadée par un utilisateur,
// ou (nil, nil) s'il n'en a aucune.
func (r *TrackRepository) FindLatestByUploader(ctx context.Context, uploaderID uuid.UUID) (*playlist.Track, error) {
	t := &playlist.Track{}
	var format string
	err := r.db.QueryRow(ctx, `
		SELECT id, playlist_id, title, artist, file_path,
		       format, size_bytes, duration, position,
		       uploaded_by, created_at, updated_at
		FROM tracks WHERE uploaded_by = $1
		ORDER BY created_at DESC LIMIT 1
	`, uploaderID).Scan(
		&t.ID, &t.PlaylistID, &t.Title, &t.Artist, &t.FilePath,
		&format, &t.SizeBytes, &t.Duration, &t.Position,
		&t.UploadedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	t.Format = playlist.Format(format)
	return t, err
}

func (r *TrackRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tracks WHERE id = $1`, id)
	return err
}
