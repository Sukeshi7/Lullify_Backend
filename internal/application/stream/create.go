package stream

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Lullify_Backend/internal/domain/stream"

	"github.com/google/uuid"
)

type CreateInput struct {
	OwnerID     uuid.UUID
	Title       string
	Description string
	MountPoint  string
}

type CreateUseCase struct {
	repo stream.Repository
}

func NewCreateUseCase(repo stream.Repository) *CreateUseCase {
	return &CreateUseCase{repo: repo}
}

func (uc *CreateUseCase) validate(input CreateInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return stream.ErrEmptyTitle
	}
	if strings.TrimSpace(input.MountPoint) == "" {
		return fmt.Errorf("mount point is required")
	}
	return nil
}

func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (*stream.Stream, error) {
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	s := &stream.Stream{
		ID:          uuid.New(),
		OwnerID:     input.OwnerID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		MountPoint:  strings.TrimSpace(input.MountPoint),
		Status:      stream.StatusOffline,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, s); err != nil {
		return nil, fmt.Errorf("creating stream: %w", err)
	}

	return s, nil
}
