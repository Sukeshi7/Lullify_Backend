package stream

import (
	"context"

	"github.com/google/uuid"
)

type Chunk []byte

type Engine interface {
	Start(ctx context.Context, streamID uuid.UUID) error
	Stop(streamID uuid.UUID) error
	Subscribe(streamID uuid.UUID) (<-chan Chunk, error)
	Unsubscribe(streamID uuid.UUID, ch <-chan Chunk)
	IsRunning(streamID uuid.UUID) bool
}
