package playlist

import "errors"

var (
	ErrEmptyTitle       = errors.New("title is required")
	ErrTitleTooLong     = errors.New("title must be less than 255 characters")
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrTrackNotFound    = errors.New("track not found")
	ErrNotOwner         = errors.New("user is not the owner of this playlist")
	ErrInvalidFilePath  = errors.New("invalid file path")

	ErrInvalidFormat    = errors.New("audio format not supported")
	ErrFileTooLarge     = errors.New("file exceeds maximum allowed size")
	ErrEmptyFile        = errors.New("uploaded file is empty")
	ErrInvalidAudioFile = errors.New("file content does not match declared format")
	ErrStorageFailure   = errors.New("failed to store audio file")
)