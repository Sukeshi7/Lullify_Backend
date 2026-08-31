package favorite_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/favorite"
)

func TestFavorite_Fields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	streamID := uuid.New()
	now := time.Now()

	f := &favorite.Favorite{
		ID:        id,
		UserID:    userID,
		StreamID:  streamID,
		CreatedAt: now,
	}

	if f.ID != id {
		t.Errorf("expected ID %s, got %s", id, f.ID)
	}
	if f.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, f.UserID)
	}
	if f.StreamID != streamID {
		t.Errorf("expected StreamID %s, got %s", streamID, f.StreamID)
	}
}

func TestFavoriteErrors(t *testing.T) {
	if favorite.ErrAlreadyFavorited == nil {
		t.Error("expected ErrAlreadyFavorited to be non-nil")
	}
	if favorite.ErrFavoriteNotFound == nil {
		t.Error("expected ErrFavoriteNotFound to be non-nil")
	}
	if favorite.ErrAlreadyFavorited.Error() == "" {
		t.Error("expected non-empty error message")
	}
	if favorite.ErrFavoriteNotFound.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
