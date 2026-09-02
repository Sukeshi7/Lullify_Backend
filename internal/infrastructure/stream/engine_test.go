package stream

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "Lullify_Backend/internal/domain/stream"
)

func TestEngine_StartAndStop(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID, ""); err != nil {
		t.Fatalf("expected no error on Start, got %v", err)
	}
	if !engine.IsRunning(streamID) {
		t.Fatal("expected stream to be running after Start")
	}
	if err := engine.Stop(streamID); err != nil {
		t.Fatalf("expected no error on Stop, got %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if engine.IsRunning(streamID) {
		t.Fatal("expected stream to be stopped after Stop")
	}
}

func TestEngine_StartAlreadyLive(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID, ""); err != nil {
		t.Fatalf("expected no error on first Start, got %v", err)
	}
	defer engine.Stop(streamID) //nolint:errcheck

	err := engine.Start(ctx, streamID, "")
	if err != domain.ErrStreamAlreadyLive {
		t.Fatalf("expected ErrStreamAlreadyLive, got %v", err)
	}
}

func TestEngine_StopNotLive(t *testing.T) {
	engine := NewStreamEngine()
	err := engine.Stop(uuid.New())
	if err != domain.ErrStreamNotLive {
		t.Fatalf("expected ErrStreamNotLive, got %v", err)
	}
}

func TestEngine_ContextCancellation(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx, cancel := context.WithCancel(context.Background())

	if err := engine.Start(ctx, streamID, ""); err != nil {
		t.Fatalf("expected no error on Start, got %v", err)
	}
	if !engine.IsRunning(streamID) {
		t.Fatal("expected stream to be running")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)

	if engine.IsRunning(streamID) {
		t.Fatal("expected stream to be stopped after context cancellation")
	}
}

func TestEngine_SubscribeUnsubscribe(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID, ""); err != nil {
		t.Fatalf("expected no error on Start, got %v", err)
	}
	defer engine.Stop(streamID) //nolint:errcheck

	ch, err := engine.Subscribe(streamID)
	if err != nil {
		t.Fatalf("expected no error on Subscribe, got %v", err)
	}
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	engine.Unsubscribe(streamID, ch)
}

func TestEngine_SubscribeToNonExistentStream(t *testing.T) {
	engine := NewStreamEngine()
	_, err := engine.Subscribe(uuid.New())
	if err == nil {
		t.Fatal("expected error when subscribing to non-existent stream, got nil")
	}
}

func TestEngine_GetSegmenter_Running(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()

	if err := engine.Start(context.Background(), streamID, ""); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer engine.Stop(streamID) //nolint:errcheck

	seg, err := engine.GetSegmenter(streamID)
	if err != nil {
		t.Fatalf("expected no error from GetSegmenter, got %v", err)
	}
	if seg == nil {
		t.Fatal("expected non-nil segmenter")
	}
}

func TestEngine_GetSegmenter_NotRunning(t *testing.T) {
	engine := NewStreamEngine()
	_, err := engine.GetSegmenter(uuid.New())
	if err == nil {
		t.Fatal("expected error for non-running stream")
	}
}

func TestEngine_MultipleStreams(t *testing.T) {
	engine := NewStreamEngine()
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	for _, id := range ids {
		if err := engine.Start(context.Background(), id, ""); err != nil {
			t.Fatalf("Start error for %s: %v", id, err)
		}
	}

	for _, id := range ids {
		if !engine.IsRunning(id) {
			t.Errorf("expected stream %s to be running", id)
		}
	}

	for _, id := range ids {
		if err := engine.Stop(id); err != nil {
			t.Fatalf("Stop error for %s: %v", id, err)
		}
	}

	time.Sleep(50 * time.Millisecond)

	for _, id := range ids {
		if engine.IsRunning(id) {
			t.Errorf("expected stream %s to be stopped", id)
		}
	}
}

func TestEngine_SubscribeAfterStop(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()

	if err := engine.Start(context.Background(), streamID, ""); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if err := engine.Stop(streamID); err != nil {
		t.Fatalf("Stop error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	_, err := engine.Subscribe(streamID)
	if err == nil {
		t.Fatal("expected error subscribing to stopped stream")
	}
}

func TestEngine_UnsubscribeNonExistentStream(t *testing.T) {
	engine := NewStreamEngine()
	ch := make(<-chan domain.Chunk)
	engine.Unsubscribe(uuid.New(), ch)
}

func TestEngine_StartWithAudioFile_NonExistent(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()

	if err := engine.Start(context.Background(), streamID, "/nonexistent/file.mp3"); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer engine.Stop(streamID) //nolint:errcheck

	if !engine.IsRunning(streamID) {
		t.Error("expected stream to be running even with bad audio file")
	}
}
