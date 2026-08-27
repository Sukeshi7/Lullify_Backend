package favorite

import "errors"

var (
	ErrAlreadyFavorited = errors.New("stream already in favorites")
	ErrFavoriteNotFound = errors.New("favorite not found")
)
