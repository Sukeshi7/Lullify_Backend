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

	var audioFilePath string

	if uc.tracks != nil {
		// Trouve la dernière track uploadée pour récupérer la playlist
		latestTrack, terr := uc.tracks.FindLatestByUploader(ctx, input.OwnerID)
		if terr == nil && latestTrack != nil {
			// Récupère toutes les tracks de cette playlist
			allTracks, perr := uc.tracks.FindByPlaylist(ctx, latestTrack.PlaylistID)
			if perr == nil && len(allTracks) > 0 {
				// Première track — passée directement à l'engine
				audioFilePath = uc.storagePath + "/" + allTracks[0].FilePath

				// Reste des tracks — pushées dans Redis
				for i := 1; i < len(allTracks); i++ {
					_ = uc.queue.Push(ctx, input.StreamID.String(), redis.TrackJob{
						TrackID:  allTracks[i].ID.String(),
						FilePath: uc.storagePath + "/" + allTracks[i].FilePath,
						Title:    allTracks[i].Title,
						Artist:   allTracks[i].Artist,
					})
				}

				observability.Logger.Info().
					Str("stream_id", input.StreamID.String()).
					Int("tracks_queued", len(allTracks)-1).
					Str("first_file", audioFilePath).
					Msg("playlist loaded into queue")
			}
		}
	}

	if audioFilePath == "" {
		// Fallback — pop depuis Redis si déjà quelque chose dedans
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
				Msg("audio file popped from queue")
		} else {
			observability.Logger.Info().
				Str("stream_id", input.StreamID.String()).
				Msg("no track in queue — stream started without audio")
		}
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
