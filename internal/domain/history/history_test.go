package history_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/history"
)

func TestEntry_Fields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	streamID := uuid.New()
	now := time.Now()

	e := &history.Entry{
		ID:         id,
		UserID:     userID,
		TrackTitle: "Chill Track",
		Artist:     "DJ Lo",
		StreamID:   &streamID,
		PlayedAt:   now,
	}

	if e.ID != id {
		t.Errorf("expected ID %s, got %s", id, e.ID)
	}
	if e.TrackTitle != "Chill Track" {
		t.Errorf("expected title 'Chill Track', got %s", e.TrackTitle)
	}
	if e.StreamID == nil || *e.StreamID != streamID {
		t.Error("expected StreamID to be set")
	}
}

func TestEntry_NoStreamID(t *testing.T) {
	e := &history.Entry{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		TrackTitle: "Track",
		PlayedAt:   time.Now(),
	}

	if e.StreamID != nil {
		t.Error("expected StreamID to be nil")
	}
}

func TestHistoryError(t *testing.T) {
	if history.ErrEmptyTitle == nil {
		t.Error("expected ErrEmptyTitle to be non-nil")
	}
	if history.ErrEmptyTitle.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
