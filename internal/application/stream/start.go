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

	if err := uc.engine.Start(context.Background(), input.StreamID); err != nil {
		return fmt.Errorf("starting engine: %w", err)
	}

	if err := uc.repo.UpdateStatus(ctx, input.StreamID, stream.StatusLive); err != nil {
		_ = uc.engine.Stop(input.StreamID)
		return fmt.Errorf("updating stream status: %w", err)
	}

	// Récupère la première track depuis la file Redis et associe le fichier audio
	job, err := uc.queue.Pop(ctx, input.StreamID.String())
	if err != nil {
		observability.Logger.Warn().
			Err(err).
			Str("stream_id", input.StreamID.String()).
			Msg("failed to pop track from queue")
	} else if job != nil {
		if setErr := uc.engine.SetAudioFile(input.StreamID, job.FilePath); setErr != nil {
			observability.Logger.Warn().
				Err(setErr).
				Str("stream_id", input.StreamID.String()).
				Msg("failed to set audio file")
		} else {
			observability.Logger.Info().
				Str("stream_id", input.StreamID.String()).
				Str("file", job.FilePath).
				Msg("audio file set for stream")
		}
	} else {
		observability.Logger.Info().
			Str("stream_id", input.StreamID.String()).
			Msg("no track in queue — stream started without audio file")
	}

	observability.ActiveStreams.Inc()

	return nil
}
