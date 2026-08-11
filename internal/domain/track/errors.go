package track

import "errors"

var (
	ErrEmptyTitle       = errors.New("title is required")
	ErrInvalidFormat    = errors.New("audio format not supported")
	ErrFileTooLarge     = errors.New("file exceeds maximum allowed size")
	ErrEmptyFile        = errors.New("uploaded file is empty")
	ErrInvalidAudioFile = errors.New("file content does not match declared format")
	ErrTrackNotFound    = errors.New("track not found")
	ErrStorageFailure   = errors.New("failed to store audio file")
)