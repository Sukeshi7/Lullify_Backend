package history

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, e *Entry) error
	FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*Entry, error)
}
