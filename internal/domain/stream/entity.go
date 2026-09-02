package stream

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOffline Status = "offline"
	StatusLive    Status = "live"
	StatusEnded   Status = "ended"
)

type Stream struct {
	ID            uuid.UUID
	OwnerID       uuid.UUID
	Title         string
	Description   string
	MountPoint    string
	Status        Status
	ListenerCount int
	StartedAt     *time.Time
	EndedAt       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (s *Stream) IsLive() bool {
	return s.Status == StatusLive
}
