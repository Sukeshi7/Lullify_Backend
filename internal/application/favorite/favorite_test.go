package favorite_test

import (
	"testing"

	"Lullify_Backend/internal/domain/favorite"
	"github.com/google/uuid"
)

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

func TestFavorite_IDs(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	streamID := uuid.New()

	if id == userID {
		t.Error("expected different UUIDs")
	}
	if userID == streamID {
		t.Error("expected different UUIDs")
	}
}
