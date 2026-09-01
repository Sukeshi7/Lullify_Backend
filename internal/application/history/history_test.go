package history_test

import (
	"testing"

	"Lullify_Backend/internal/domain/history"
	"github.com/google/uuid"
)

func TestHistoryError(t *testing.T) {
	if history.ErrEmptyTitle == nil {
		t.Error("expected ErrEmptyTitle to be non-nil")
	}
	if history.ErrEmptyTitle.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestEntry_StreamID_Optional(t *testing.T) {
	e := &history.Entry{}
	if e.StreamID != nil {
		t.Error("expected StreamID to be nil by default")
	}

	streamID := uuid.New()
	e.StreamID = &streamID
	if e.StreamID == nil {
		t.Error("expected StreamID to be set")
	}
}
