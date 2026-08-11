package track

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, t *Track) error
	FindByID(ctx context.Context, id uuid.UUID) (*Track, error)
	ListByPlaylist(ctx context.Context, playlistID uuid.UUID) ([]*Track, error)
	Delete(ctx context.Context, id uuid.UUID) error
}