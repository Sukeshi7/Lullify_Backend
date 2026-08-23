package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	playlistQueuePrefix = "queue:stream:"
	blockTimeout        = 2 * time.Second
)

// TrackJob représente une track dans la file de lecture
type TrackJob struct {
	TrackID  string `json:"track_id"`
	FilePath string `json:"file_path"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Duration int    `json:"duration"`
}

// Push ajoute une track en fin de file (LPUSH)
func (c *Client) Push(ctx context.Context, streamID string, job TrackJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshaling track job: %w", err)
	}

	key := playlistQueuePrefix + streamID
	if err := c.rdb.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("pushing to queue %s: %w", key, err)
	}

	return nil
}

// Pop retire et retourne la prochaine track (BRPOP — bloquant)
// Retourne nil, nil si le context est annulé ou timeout
func (c *Client) Pop(ctx context.Context, streamID string) (*TrackJob, error) {
	key := playlistQueuePrefix + streamID

	result, err := c.rdb.BRPop(ctx, blockTimeout, key).Result()
	if err != nil {
		// Timeout ou context annulé — pas une erreur critique
		return nil, nil
	}

	// BRPop retourne [key, value]
	if len(result) < 2 {
		return nil, nil
	}

	var job TrackJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshalling track job: %w", err)
	}

	return &job, nil
}

// Len retourne le nombre de tracks en attente dans la file
func (c *Client) Len(ctx context.Context, streamID string) (int64, error) {
	key := playlistQueuePrefix + streamID
	return c.rdb.LLen(ctx, key).Result()
}

// Clear vide la file d'un stream
func (c *Client) Clear(ctx context.Context, streamID string) error {
	key := playlistQueuePrefix + streamID
	return c.rdb.Del(ctx, key).Err()
}
