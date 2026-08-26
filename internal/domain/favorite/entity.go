package favorite

import (
	"time"

	"github.com/google/uuid"
)

type Favorite struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	StreamID  uuid.UUID
	CreatedAt time.Time
}
