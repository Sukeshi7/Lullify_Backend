package history

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Lullify_Backend/internal/domain/history"

	"github.com/google/uuid"
)

type RecordInput struct {
	UserID     uuid.UUID
	TrackTitle string
	Artist     string
	StreamID   *uuid.UUID
}

type RecordUseCase struct {
	repo history.Repository
}

func NewRecordUseCase(repo history.Repository) *RecordUseCase {
	return &RecordUseCase{repo: repo}
}

func (uc *RecordUseCase) Execute(ctx context.Context, input RecordInput) (*history.Entry, error) {
	if strings.TrimSpace(input.TrackTitle) == "" {
		return nil, history.ErrEmptyTitle
	}

	e := &history.Entry{
		ID:         uuid.New(),
		UserID:     input.UserID,
		TrackTitle: strings.TrimSpace(input.TrackTitle),
		Artist:     strings.TrimSpace(input.Artist),
		StreamID:   input.StreamID,
		PlayedAt:   time.Now().UTC(),
	}

	if err := uc.repo.Save(ctx, e); err != nil {
		return nil, fmt.Errorf("saving history entry: %w", err)
	}

	return e, nil
}
