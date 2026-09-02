package favorite

import (
	"context"
	"fmt"
	"time"

	"Lullify_Backend/internal/domain/favorite"

	"github.com/google/uuid"
)

type AddInput struct {
	UserID   uuid.UUID
	StreamID uuid.UUID
}

type AddUseCase struct {
	repo favorite.Repository
}

func NewAddUseCase(repo favorite.Repository) *AddUseCase {
	return &AddUseCase{repo: repo}
}

func (uc *AddUseCase) Execute(ctx context.Context, input AddInput) (*favorite.Favorite, error) {
	f := &favorite.Favorite{
		ID:        uuid.New(),
		UserID:    input.UserID,
		StreamID:  input.StreamID,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.repo.Save(ctx, f); err != nil {
		return nil, fmt.Errorf("saving favorite: %w", err)
	}

	return f, nil
}
