package history

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TrackTitle string
	Artist     string
	StreamID   *uuid.UUID // optionnel : d'où venait l'écoute
	PlayedAt   time.Time
}
