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

type Format string

const (
	FormatMP3  Format = "mp3"
	FormatFLAC Format = "flac"
	FormatWAV  Format = "wav"
)

func (f Format) IsValid() bool {
	switch f {
	case FormatMP3, FormatFLAC, FormatWAV:
		return true
	}
	return false
}

type Track struct {
	ID         uuid.UUID
	PlaylistID uuid.UUID
	Title      string
	Artist     string
	FilePath   string
	Format     Format
	SizeBytes  int64
	Duration   int
	Position   int
	UploadedBy uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
