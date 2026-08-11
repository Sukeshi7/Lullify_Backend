package track

import (
	"time"

	"github.com/google/uuid"
)

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
	ID          uuid.UUID
	PlaylistID  uuid.UUID
	Title       string
	StorageKey  string
	Format      Format
	SizeBytes   int64
	DurationSec int
	UploadedBy  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}