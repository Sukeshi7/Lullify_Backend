package playlist

import (
	"context"

	"github.com/google/uuid"
)

type PlaylistRepository interface {
	Create(ctx context.Context, p *Playlist) error
	FindByID(ctx context.Context, id uuid.UUID) (*Playlist, error)
	FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Playlist, error)
	Update(ctx context.Context, p *Playlist) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TrackRepository interface {
	Create(ctx context.Context, t *Track) error
	FindByID(ctx context.Context, id uuid.UUID) (*Track, error)
	FindByPlaylist(ctx context.Context, playlistID uuid.UUID) ([]*Track, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
