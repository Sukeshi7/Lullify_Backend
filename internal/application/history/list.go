package history

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/history"

	"github.com/google/uuid"
)

const defaultHistoryLimit = 50

type ListUseCase struct {
	repo history.Repository
}

func NewListUseCase(repo history.Repository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*history.Entry, error) {
	entries, err := uc.repo.FindByUser(ctx, userID, defaultHistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("listing history: %w", err)
	}
	return entries, nil
}
