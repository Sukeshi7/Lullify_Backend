package stream

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "Lullify_Backend/internal/domain/stream"
)

// Vérifie qu'un stream démarre et s'arrête proprement.
func TestEngine_StartAndStop(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID); err != nil {
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

// Vérifie qu'on ne peut pas démarrer deux fois le même stream.
func TestEngine_StartAlreadyLive(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID); err != nil {
		t.Fatalf("expected no error on first Start, got %v", err)
	}
	defer engine.Stop(streamID)

	err := engine.Start(ctx, streamID)
	if err != domain.ErrStreamAlreadyLive {
		t.Fatalf("expected ErrStreamAlreadyLive, got %v", err)
	}
}

// Vérifie qu'on ne peut pas arrêter un stream inexistant.
func TestEngine_StopNotLive(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()

	err := engine.Stop(streamID)
	if err != domain.ErrStreamNotLive {
		t.Fatalf("expected ErrStreamNotLive, got %v", err)
	}
}

// Vérifie que l'annulation du context arrête bien le stream.
func TestEngine_ContextCancellation(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx, cancel := context.WithCancel(context.Background())

	if err := engine.Start(ctx, streamID); err != nil {
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

// Vérifie qu'un auditeur peut s'abonner et se désabonner sans crash.
func TestEngine_SubscribeUnsubscribe(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()
	ctx := context.Background()

	if err := engine.Start(ctx, streamID); err != nil {
		t.Fatalf("expected no error on Start, got %v", err)
	}
	defer engine.Stop(streamID)

	ch, err := engine.Subscribe(streamID)
	if err != nil {
		t.Fatalf("expected no error on Subscribe, got %v", err)
	}
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	engine.Unsubscribe(streamID, ch)
}

// Vérifie qu'on ne peut pas s'abonner à un stream qui ne tourne pas.
func TestEngine_SubscribeToNonExistentStream(t *testing.T) {
	engine := NewStreamEngine()
	streamID := uuid.New()

	_, err := engine.Subscribe(streamID)
	if err == nil {
		t.Fatal("expected error when subscribing to non-existent stream, got nil")
	}
}