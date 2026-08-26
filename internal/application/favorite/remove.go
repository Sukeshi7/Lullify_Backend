package favorite

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/favorite"

	"github.com/google/uuid"
)

type RemoveInput struct {
	UserID   uuid.UUID
	StreamID uuid.UUID
}

type RemoveUseCase struct {
	repo favorite.Repository
}

func NewRemoveUseCase(repo favorite.Repository) *RemoveUseCase {
	return &RemoveUseCase{repo: repo}
}

func (uc *RemoveUseCase) Execute(ctx context.Context, input RemoveInput) error {
	if err := uc.repo.Delete(ctx, input.UserID, input.StreamID); err != nil {
		return fmt.Errorf("removing favorite: %w", err)
	}
	return nil
}
