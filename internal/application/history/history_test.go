package history_test

import (
	"testing"

	"Lullify_Backend/internal/domain/history"
)

func TestHistoryError(t *testing.T) {
	if history.ErrEmptyTitle == nil {
		t.Error("expected ErrEmptyTitle to be non-nil")
	}
	if history.ErrEmptyTitle.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestHistoryError_Message(t *testing.T) {
	if history.ErrEmptyTitle.Error() != "track title is required" {
		t.Errorf("unexpected error message: %s", history.ErrEmptyTitle.Error())
	}
}
