package playlist

import (
	"time"

	"github.com/google/uuid"
)

type Playlist struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Track struct {
	ID         uuid.UUID
	PlaylistID uuid.UUID
	Title      string
	Artist     string
	FilePath   string
	Duration   int
	Position   int
	CreatedAt  time.Time
}
