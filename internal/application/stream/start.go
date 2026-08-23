package stream

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/stream"

	"github.com/google/uuid"
)

type StartInput struct {
	StreamID uuid.UUID
	OwnerID  uuid.UUID
}

type StartUseCase struct {
	repo   stream.Repository
	engine stream.Engine
}

func NewStartUseCase(repo stream.Repository, engine stream.Engine) *StartUseCase {
	return &StartUseCase{repo: repo, engine: engine}
}

func (uc *StartUseCase) Execute(ctx context.Context, input StartInput) error {
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
	if s.IsLive() {
		return stream.ErrStreamAlreadyLive
	}

	if err := uc.engine.Start(context.Background(), input.StreamID); err != nil {
		return fmt.Errorf("starting engine: %w", err)
	}

	if err := uc.repo.UpdateStatus(ctx, input.StreamID, stream.StatusLive); err != nil {
		_ = uc.engine.Stop(input.StreamID)
		return fmt.Errorf("updating stream status: %w", err)
	}

	return nil
}
