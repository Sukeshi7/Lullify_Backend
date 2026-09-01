package track_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	apptrack "Lullify_Backend/internal/application/track"
	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/infrastructure/storage"
)

type mockTrackRepo struct {
	create func(ctx context.Context, t *playlist.Track) error
}

func (m *mockTrackRepo) Create(ctx context.Context, t *playlist.Track) error {
	return m.create(ctx, t)
}
func (m *mockTrackRepo) FindByID(_ context.Context, _ uuid.UUID) (*playlist.Track, error) {
	return nil, nil
}
func (m *mockTrackRepo) FindByPlaylist(_ context.Context, _ uuid.UUID) ([]*playlist.Track, error) {
	return nil, nil
}
func (m *mockTrackRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type mockPlaylistRepo struct {
	findByID func(ctx context.Context, id uuid.UUID) (*playlist.Playlist, error)
}

func (m *mockPlaylistRepo) FindByID(ctx context.Context, id uuid.UUID) (*playlist.Playlist, error) {
	return m.findByID(ctx, id)
}
func (m *mockPlaylistRepo) Create(_ context.Context, _ *playlist.Playlist) error { return nil }
func (m *mockPlaylistRepo) FindByOwner(_ context.Context, _ uuid.UUID) ([]*playlist.Playlist, error) {
	return nil, nil
}
func (m *mockPlaylistRepo) Update(_ context.Context, _ *playlist.Playlist) error { return nil }
func (m *mockPlaylistRepo) Delete(_ context.Context, _ uuid.UUID) error          { return nil }

func newOwnerPlaylist(ownerID uuid.UUID) *playlist.Playlist {
	return &playlist.Playlist{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Title:   "My Playlist",
	}
}

func mp3Header() []byte {
	header := make([]byte, 512)
	header[0] = 0xFF
	header[1] = 0xFB
	return header
}

func TestUpload_Success(t *testing.T) {
	ownerID := uuid.New()
	p := newOwnerPlaylist(ownerID)

	localDir := t.TempDir()
	s, _ := storage.New(storage.Options{Provider: "local", LocalPath: localDir})

	trackRepo := &mockTrackRepo{
		create: func(_ context.Context, _ *playlist.Track) error { return nil },
	}
	playlistRepo := &mockPlaylistRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*playlist.Playlist, error) {
			return p, nil
		},
	}

	uc := apptrack.NewUploadUseCase(trackRepo, playlistRepo, s, 10*1024*1024)

	data := mp3Header()
	data = append(data, make([]byte, 100)...)

	track, err := uc.Execute(context.Background(), apptrack.UploadInput{
		PlaylistID: p.ID,
		UploaderID: ownerID,
		Title:      "Chill Track",
		Artist:     "DJ Lo",
		Format:     playlist.FormatMP3,
		SizeBytes:  int64(len(data)),
		Reader:     bytes.NewReader(data),
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if track == nil {
		t.Fatal("expected track, got nil")
	}
}

func TestUpload_EmptyTitle(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "", Format: playlist.FormatMP3, SizeBytes: 100,
	})
	if !errors.Is(err, playlist.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestUpload_InvalidFormat(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.Format("ogg"), SizeBytes: 100,
	})
	if !errors.Is(err, playlist.ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestUpload_EmptyFile(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.FormatMP3, SizeBytes: 0,
	})
	if !errors.Is(err, playlist.ErrEmptyFile) {
		t.Errorf("expected ErrEmptyFile, got %v", err)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.FormatMP3, SizeBytes: 2048,
	})
	if !errors.Is(err, playlist.ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestUpload_PlaylistNotFound(t *testing.T) {
	playlistRepo := &mockPlaylistRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*playlist.Playlist, error) {
			return nil, nil
		},
	}
	uc := apptrack.NewUploadUseCase(nil, playlistRepo, nil, 1024*1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.FormatMP3, SizeBytes: 100,
	})
	if !errors.Is(err, playlist.ErrPlaylistNotFound) {
		t.Errorf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestUpload_NotOwner(t *testing.T) {
	p := newOwnerPlaylist(uuid.New())
	playlistRepo := &mockPlaylistRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*playlist.Playlist, error) {
			return p, nil
		},
	}
	uc := apptrack.NewUploadUseCase(nil, playlistRepo, nil, 1024*1024)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.FormatMP3, SizeBytes: 100, UploaderID: uuid.New(),
	})
	if !errors.Is(err, playlist.ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestUpload_InvalidAudioFile(t *testing.T) {
	ownerID := uuid.New()
	p := newOwnerPlaylist(ownerID)
	playlistRepo := &mockPlaylistRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*playlist.Playlist, error) {
			return p, nil
		},
	}
	uc := apptrack.NewUploadUseCase(nil, playlistRepo, nil, 1024*1024)
	data := make([]byte, 512)
	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title: "Track", Format: playlist.FormatMP3, SizeBytes: int64(len(data)),
		UploaderID: ownerID, Reader: bytes.NewReader(data),
	})
	if !errors.Is(err, playlist.ErrInvalidAudioFile) {
		t.Errorf("expected ErrInvalidAudioFile, got %v", err)
	}
}
