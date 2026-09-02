package stream

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/infrastructure/observability"

	"github.com/google/uuid"
)

type StopInput struct {
	StreamID uuid.UUID
	OwnerID  uuid.UUID
}

type StopUseCase struct {
	repo   stream.Repository
	engine stream.Engine
}

func NewStopUseCase(repo stream.Repository, engine stream.Engine) *StopUseCase {
	return &StopUseCase{repo: repo, engine: engine}
}

func (uc *StopUseCase) Execute(ctx context.Context, input StopInput) error {
	s, err := uc.repo.FindByID(ctx, input.StreamID)
	if err != nil {
		return fmt.Errorf("finding stream: %w", err)
	}
	if s == nil {
		return stream.ErrStreamNotFound
	}
	if s.OwnerID != input.OwnerID {
		return stream.ErrNotStreamOwner
	}
	if !s.IsLive() {
		return stream.ErrStreamNotLive
	}

	if err := uc.engine.Stop(input.StreamID); err != nil {
		return fmt.Errorf("stopping engine: %w", err)
	}

	if err := uc.repo.UpdateStatus(ctx, input.StreamID, stream.StatusEnded); err != nil {
		return fmt.Errorf("updating stream status: %w", err)
	}

	// Métrique : un stream de moins en live
	observability.ActiveStreams.Dec()

	return nil
}
