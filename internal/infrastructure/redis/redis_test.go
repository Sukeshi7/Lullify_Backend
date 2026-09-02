package redis_test

import (
	"context"
	"testing"
	"time"

	"Lullify_Backend/internal/infrastructure/redis"
)

func newTestClient(t *testing.T) *redis.Client {
	t.Helper()
	client, err := redis.NewClient("redis://localhost:6379")
	if err != nil {
		t.Skip("Redis not available")
	}
	return client
}

func TestClient_Ping(t *testing.T) {
	client := newTestClient(t)
	defer client.Close() //nolint:errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("expected ping to succeed, got %v", err)
	}
}

func TestQueue_PushAndLen(t *testing.T) {
	client := newTestClient(t)
	defer client.Close() //nolint:errcheck

	ctx := context.Background()
	streamID := "test-stream-len-" + time.Now().Format("20060102150405")

	// Nettoie avant le test
	_ = client.Clear(ctx, streamID)

	job := redis.TrackJob{
		TrackID:  "track-1",
		FilePath: "/audio/track.mp3",
		Title:    "Test Track",
		Artist:   "DJ Test",
		Duration: 180,
	}

	if err := client.Push(ctx, streamID, job); err != nil {
		t.Fatalf("Push error: %v", err)
	}

	length, err := client.Len(ctx, streamID)
	if err != nil {
		t.Fatalf("Len error: %v", err)
	}
	if length != 1 {
		t.Errorf("expected length 1, got %d", length)
	}

	// Nettoie après
	_ = client.Clear(ctx, streamID)
}

func TestQueue_Clear(t *testing.T) {
	client := newTestClient(t)
	defer client.Close() //nolint:errcheck

	ctx := context.Background()
	streamID := "test-stream-clear-" + time.Now().Format("20060102150405")

	job := redis.TrackJob{TrackID: "t1", FilePath: "/f", Title: "T", Artist: "A", Duration: 1}
	_ = client.Push(ctx, streamID, job)
	_ = client.Push(ctx, streamID, job)

	if err := client.Clear(ctx, streamID); err != nil {
		t.Fatalf("Clear error: %v", err)
	}

	length, _ := client.Len(ctx, streamID)
	if length != 0 {
		t.Errorf("expected length 0 after clear, got %d", length)
	}
}
