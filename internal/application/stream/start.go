package stream

import (
	"context"
	"fmt"

	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/infrastructure/observability"
	"Lullify_Backend/internal/infrastructure/redis"

	"github.com/google/uuid"
)

type Queue interface {
	Pop(ctx context.Context, streamID string) (*redis.TrackJob, error)
	Push(ctx context.Context, streamID string, job redis.TrackJob) error
}

type TrackFinder interface {
	FindLatestByUploader(ctx context.Context, uploaderID uuid.UUID) (*playlist.Track, error)
	FindByPlaylist(ctx context.Context, playlistID uuid.UUID) ([]*playlist.Track, error)
}

type StartInput struct {
	StreamID uuid.UUID
	OwnerID  uuid.UUID
}

type StartUseCase struct {
	repo        stream.Repository
	engine      stream.Engine
	queue       Queue
	tracks      TrackFinder
	storagePath string
}

func NewStartUseCase(repo stream.Repository, engine stream.Engine, queue Queue, tracks TrackFinder, storagePath string) *StartUseCase {
	return &StartUseCase{repo: repo, engine: engine, queue: queue, tracks: tracks, storagePath: storagePath}
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

	if uc.tracks != nil {
		latestTrack, terr := uc.tracks.FindLatestByUploader(ctx, input.OwnerID)
		if terr == nil && latestTrack != nil {
			allTracks, perr := uc.tracks.FindByPlaylist(ctx, latestTrack.PlaylistID)
			if perr == nil && len(allTracks) > 0 {
				// Push toutes les tracks dans Redis — l'engine les collectera toutes
				for _, track := range allTracks {
					_ = uc.queue.Push(ctx, input.StreamID.String(), redis.TrackJob{
						TrackID:  track.ID.String(),
						FilePath: uc.storagePath + "/" + track.FilePath,
						Title:    track.Title,
						Artist:   track.Artist,
					})
				}

				observability.Logger.Info().
					Str("stream_id", input.StreamID.String()).
					Int("tracks_queued", len(allTracks)).
					Msg("playlist loaded into queue")
			}
		}
	}

	// L'engine collecte lui-même toutes les tracks depuis Redis
	err = uc.engine.Start(context.Background(), input.StreamID, "")
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
