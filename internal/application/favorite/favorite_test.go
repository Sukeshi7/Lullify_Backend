package favorite_test

import (
	"testing"

	"Lullify_Backend/internal/domain/favorite"
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

func TestFavoriteErrors_Distinct(t *testing.T) {
	if favorite.ErrAlreadyFavorited == favorite.ErrFavoriteNotFound {
		t.Error("expected distinct errors")
	}
}