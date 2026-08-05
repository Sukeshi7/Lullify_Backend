package playlist

import "errors"

var (
	ErrEmptyTitle       = errors.New("title is required")
	ErrTitleTooLong     = errors.New("title must be less than 255 characters")
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrTrackNotFound    = errors.New("track not found")
	ErrNotOwner         = errors.New("user is not the owner of this playlist")
	ErrInvalidFilePath  = errors.New("invalid file path")
)
