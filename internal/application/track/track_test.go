package track_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	apptrack "Lullify_Backend/internal/application/track"
	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/infrastructure/storage"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockTrackRepo struct {
	create func(ctx context.Context, t *playlist.Track) error
}

func (m *mockTrackRepo) Create(ctx context.Context, t *playlist.Track) error {
	return m.create(ctx, t)
}
func (m *mockTrackRepo) FindByID(ctx context.Context, id uuid.UUID) (*playlist.Track, error) {
	return nil, nil
}
func (m *mockTrackRepo) FindByPlaylist(ctx context.Context, playlistID uuid.UUID) ([]*playlist.Track, error) {
	return nil, nil
}
func (m *mockTrackRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockPlaylistRepo struct {
	findByID func(ctx context.Context, id uuid.UUID) (*playlist.Playlist, error)
}

func (m *mockPlaylistRepo) FindByID(ctx context.Context, id uuid.UUID) (*playlist.Playlist, error) {
	return m.findByID(ctx, id)
}
func (m *mockPlaylistRepo) Create(ctx context.Context, p *playlist.Playlist) error { return nil }
func (m *mockPlaylistRepo) FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]*playlist.Playlist, error) {
	return nil, nil
}
func (m *mockPlaylistRepo) Update(ctx context.Context, p *playlist.Playlist) error { return nil }
func (m *mockPlaylistRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }

type mockStorage struct {
	upload func(ctx context.Context, key string, r interface{}, size int64, ct string) error
	delete func(ctx context.Context, key string) error
}

func (m *mockStorage) Upload(ctx context.Context, key string, r interface{ Read([]byte) (int, error) }, size int64, ct string) error {
	return m.upload(ctx, key, r, size, ct)
}
func (m *mockStorage) Delete(ctx context.Context, key string) error { return m.delete(ctx, key) }
func (m *mockStorage) PresignedGetURL(ctx context.Context, key string, expiry int) (string, error) {
	return "", nil
}

func newOwnerPlaylist(ownerID uuid.UUID) *playlist.Playlist {
	return &playlist.Playlist{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Title:   "My Playlist",
	}
}

// mp3 magic bytes
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
	if track.Title != "Chill Track" {
		t.Errorf("expected title 'Chill Track', got %s", track.Title)
	}
}

func TestUpload_EmptyTitle(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)

	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title:     "",
		Format:    playlist.FormatMP3,
		SizeBytes: 100,
	})

	if !errors.Is(err, playlist.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestUpload_InvalidFormat(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)

	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title:     "Track",
		Format:    playlist.Format("ogg"),
		SizeBytes: 100,
	})

	if !errors.Is(err, playlist.ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestUpload_EmptyFile(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)

	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title:     "Track",
		Format:    playlist.FormatMP3,
		SizeBytes: 0,
	})

	if !errors.Is(err, playlist.ErrEmptyFile) {
		t.Errorf("expected ErrEmptyFile, got %v", err)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	uc := apptrack.NewUploadUseCase(nil, nil, nil, 1024)

	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title:     "Track",
		Format:    playlist.FormatMP3,
		SizeBytes: 2048,
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
		Title:     "Track",
		Format:    playlist.FormatMP3,
		SizeBytes: 100,
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
		Title:      "Track",
		Format:     playlist.FormatMP3,
		SizeBytes:  100,
		UploaderID: uuid.New(), // different owner
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

	// Données qui ne correspondent pas au format MP3
	data := make([]byte, 512)

	_, err := uc.Execute(context.Background(), apptrack.UploadInput{
		Title:      "Track",
		Format:     playlist.FormatMP3,
		SizeBytes:  int64(len(data)),
		UploaderID: ownerID,
		Reader:     bytes.NewReader(data),
	})

	if !errors.Is(err, playlist.ErrInvalidAudioFile) {
		t.Errorf("expected ErrInvalidAudioFile, got %v", err)
	}
}

// Test des fonctions pures matchesFormat et contentTypeFor via les erreurs
func TestContentTypeFor(t *testing.T) {
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

	formats := []struct {
		format playlist.Format
		header []byte
	}{
		{playlist.FormatFLAC, append([]byte("fLaC"), make([]byte, 508)...)},
		{playlist.FormatWAV, append([]byte("RIFF"), append(make([]byte, 4), append([]byte("WAVE"), make([]byte, 500)...)...)...)},
	}

	for _, f := range formats {
		_, err := uc.Execute(context.Background(), apptrack.UploadInput{
			Title:      "Track",
			Format:     f.format,
			SizeBytes:  int64(len(f.header)),
			UploaderID: ownerID,
			PlaylistID: p.ID,
			Reader:     bytes.NewReader(f.header),
		})
		// L'erreur attendue ici est nil ou une erreur de storage — pas ErrInvalidAudioFile
		if errors.Is(err, playlist.ErrInvalidAudioFile) {
			t.Errorf("format %s: unexpected ErrInvalidAudioFile", f.format)
		}
		_ = time.Now() // évite import inutilisé
	}
}
