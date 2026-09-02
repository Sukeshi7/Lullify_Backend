package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/playlist"
)

func TestStatusForUploadError(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{playlist.ErrEmptyTitle, http.StatusBadRequest},
		{playlist.ErrInvalidFormat, http.StatusBadRequest},
		{playlist.ErrEmptyFile, http.StatusBadRequest},
		{playlist.ErrInvalidAudioFile, http.StatusBadRequest},
		{playlist.ErrFileTooLarge, http.StatusRequestEntityTooLarge},
		{playlist.ErrPlaylistNotFound, http.StatusNotFound},
		{playlist.ErrNotOwner, http.StatusForbidden},
		{playlist.ErrStorageFailure, http.StatusBadGateway},
	}

	for _, c := range cases {
		got := statusForUploadError(c.err)
		if got != c.expected {
			t.Errorf("statusForUploadError(%v) = %d, want %d", c.err, got, c.expected)
		}
	}
}

func TestTrackToJSON(t *testing.T) {
	track := &playlist.Track{
		ID:         uuid.New(),
		PlaylistID: uuid.New(),
		Title:      "Chill Track",
		Artist:     "DJ Lo",
		FilePath:   "/audio/track.mp3",
		Format:     playlist.FormatMP3,
		SizeBytes:  1024,
		Duration:   180,
		Position:   1,
		UploadedBy: uuid.New(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result := trackToJSON(track)

	if result["title"] != track.Title {
		t.Errorf("expected title %s, got %s", track.Title, result["title"])
	}
	if result["format"] != string(track.Format) {
		t.Errorf("expected format %s, got %s", track.Format, result["format"])
	}
	if result["duration"] != track.Duration {
		t.Errorf("expected duration %d, got %v", track.Duration, result["duration"])
	}
}

func TestWriteJSON_StatusCodes(t *testing.T) {
	cases := []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}
	for _, code := range cases {
		w := httptest.NewRecorder()
		writeJSON(w, code, map[string]string{"status": "ok"})
		if w.Code != code {
			t.Errorf("expected %d, got %d", code, w.Code)
		}
	}
}
