package stream

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/infrastructure/observability"
	"Lullify_Backend/internal/infrastructure/redis"

	"github.com/google/uuid"
)

type StartInput struct {
	StreamID uuid.UUID
	OwnerID  uuid.UUID
}

type StartUseCase struct {
	repo   stream.Repository
	engine stream.Engine
	queue  *redis.Client
}

func NewStartUseCase(repo stream.Repository, engine stream.Engine, queue *redis.Client) *StartUseCase {
	return &StartUseCase{repo: repo, engine: engine, queue: queue}
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

	// Pop la track AVANT de démarrer — zéro race condition
	var audioFilePath string
	job, popErr := uc.queue.Pop(ctx, input.StreamID.String())
	if popErr != nil {
		observability.Logger.Warn().
			Err(popErr).
			Str("stream_id", input.StreamID.String()).
			Msg("failed to pop track from queue")
	} else if job != nil {
		audioFilePath = job.FilePath
		observability.Logger.Info().
			Str("stream_id", input.StreamID.String()).
			Str("file", job.FilePath).
			Msg("audio file queued for stream")
	} else {
		observability.Logger.Info().
			Str("stream_id", input.StreamID.String()).
			Msg("no track in queue — stream started without audio file")
	}

	err = uc.engine.Start(context.Background(), input.StreamID, audioFilePath)
	if err != nil {
		return fmt.Errorf("starting engine: %w", err)
	}

	err = uc.repo.UpdateStatus(ctx, input.StreamID, stream.StatusLive)
	if err != nil {
		_ = uc.engine.Stop(input.StreamID)
		return fmt.Errorf("updating stream status: %w", err)
	}

	observability.ActiveStreams.Inc()

	return nil
}
