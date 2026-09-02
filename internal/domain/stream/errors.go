package stream

import "errors"

var (
	ErrStreamNotFound    = errors.New("stream not found")
	ErrStreamAlreadyLive = errors.New("stream is already live")
	ErrStreamNotLive     = errors.New("stream is not live")
	ErrNotStreamOwner    = errors.New("not the stream owner")
	ErrEmptyTitle        = errors.New("title is required")
	ErrMountPointTaken   = errors.New("mount point already in use")
)
