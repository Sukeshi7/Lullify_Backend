package favorite

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/favorite"

	"github.com/google/uuid"
)

type ListUseCase struct {
	repo favorite.Repository
}

func NewListUseCase(repo favorite.Repository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*favorite.Favorite, error) {
	favorites, err := uc.repo.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing favorites: %w", err)
	}
	return favorites, nil
}
