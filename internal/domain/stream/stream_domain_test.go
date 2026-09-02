package stream_test

import (
	"testing"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/stream"
)

func TestStream_IsLive(t *testing.T) {
	s := &stream.Stream{
		ID:     uuid.New(),
		Status: stream.StatusLive,
	}
	if !s.IsLive() {
		t.Error("expected stream to be live")
	}
}

func TestStream_IsNotLive_Offline(t *testing.T) {
	s := &stream.Stream{
		ID:     uuid.New(),
		Status: stream.StatusOffline,
	}
	if s.IsLive() {
		t.Error("expected stream to not be live when offline")
	}
}

func TestStream_IsNotLive_Ended(t *testing.T) {
	s := &stream.Stream{
		ID:     uuid.New(),
		Status: stream.StatusEnded,
	}
	if s.IsLive() {
		t.Error("expected stream to not be live when ended")
	}
}

func TestStreamStatus_Constants(t *testing.T) {
	if stream.StatusOffline != "offline" {
		t.Errorf("expected StatusOffline to be 'offline', got %s", stream.StatusOffline)
	}
	if stream.StatusLive != "live" {
		t.Errorf("expected StatusLive to be 'live', got %s", stream.StatusLive)
	}
	if stream.StatusEnded != "ended" {
		t.Errorf("expected StatusEnded to be 'ended', got %s", stream.StatusEnded)
	}
}
