package track

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/infrastructure/storage"
)

type UploadInput struct {
	PlaylistID uuid.UUID
	UploaderID uuid.UUID
	Title      string
	Artist     string
	Format     playlist.Format
	SizeBytes  int64
	Reader     io.Reader
}

type UploadUseCase struct {
	tracks    playlist.TrackRepository
	playlists playlist.Repository
	storage   storage.Storage
	maxSize   int64
}

func NewUploadUseCase(
	tracks playlist.TrackRepository,
	playlists playlist.Repository,
	storage storage.Storage,
	maxSize int64,
) *UploadUseCase {
	return &UploadUseCase{
		tracks:    tracks,
		playlists: playlists,
		storage:   storage,
		maxSize:   maxSize,
	}
}

func (uc *UploadUseCase) Execute(ctx context.Context, in UploadInput) (*playlist.Track, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, playlist.ErrEmptyTitle
	}
	if !in.Format.IsValid() {
		return nil, playlist.ErrInvalidFormat
	}
	if in.SizeBytes <= 0 {
		return nil, playlist.ErrEmptyFile
	}
	if in.SizeBytes > uc.maxSize {
		return nil, playlist.ErrFileTooLarge
	}

	p, err := uc.playlists.FindByID(ctx, in.PlaylistID)
	if err != nil {
		return nil, fmt.Errorf("checking playlist: %w", err)
	}
	if p == nil {
		return nil, playlist.ErrPlaylistNotFound
	}
	if p.OwnerID != in.UploaderID {
		return nil, playlist.ErrNotOwner
	}

	sniffBuf := make([]byte, 512)
	n, err := io.ReadFull(in.Reader, sniffBuf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("reading file header: %w", err)
	}
	sniffBuf = sniffBuf[:n]

	if !matchesFormat(sniffBuf, in.Format) {
		return nil, playlist.ErrInvalidAudioFile
	}

	fullReader := io.MultiReader(bytes.NewReader(sniffBuf), in.Reader)

	trackID := uuid.New()
	storageKey := fmt.Sprintf("playlists/%s/tracks/%s.%s",
		in.PlaylistID.String(), trackID.String(), in.Format)

	contentType := contentTypeFor(in.Format)
	if err := uc.storage.Upload(ctx, storageKey, fullReader, in.SizeBytes, contentType); err != nil {
		return nil, fmt.Errorf("%w: %v", playlist.ErrStorageFailure, err)
	}

	now := time.Now().UTC()
	track := &playlist.Track{
		ID:         trackID,
		PlaylistID: in.PlaylistID,
		Title:      strings.TrimSpace(in.Title),
		Artist:     strings.TrimSpace(in.Artist),
		FilePath:   storageKey,
		Format:     in.Format,
		SizeBytes:  in.SizeBytes,
		Duration:   0,
		Position:   0,
		UploadedBy: in.UploaderID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.tracks.Create(ctx, track); err != nil {
		_ = uc.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("creating track: %w", err)
	}

	return track, nil
}

func matchesFormat(header []byte, format playlist.Format) bool {
	if len(header) < 4 {
		return false
	}
	switch format {
	case playlist.FormatMP3:
		if bytes.HasPrefix(header, []byte("ID3")) {
			return true
		}
		if len(header) >= 2 && header[0] == 0xFF && (header[1]&0xE0) == 0xE0 {
			return true
		}
		return false
	case playlist.FormatFLAC:
		return bytes.HasPrefix(header, []byte("fLaC"))
	case playlist.FormatWAV:
		return bytes.HasPrefix(header, []byte("RIFF")) &&
			len(header) >= 12 && bytes.Equal(header[8:12], []byte("WAVE"))
	}
	return false
}

func contentTypeFor(format playlist.Format) string {
	switch format {
	case playlist.FormatMP3:
		return "audio/mpeg"
	case playlist.FormatFLAC:
		return "audio/flac"
	case playlist.FormatWAV:
		return "audio/wav"
	}
	return "application/octet-stream"
}
