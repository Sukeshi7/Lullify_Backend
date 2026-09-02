package stream

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, s *Stream) error
	FindByID(ctx context.Context, id uuid.UUID) (*Stream, error)
	FindActive(ctx context.Context) ([]*Stream, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	IncrementListeners(ctx context.Context, id uuid.UUID) error
	DecrementListeners(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
