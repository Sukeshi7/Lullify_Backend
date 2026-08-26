package favorite

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, f *Favorite) error
	Delete(ctx context.Context, userID, streamID uuid.UUID) error
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*Favorite, error)
}