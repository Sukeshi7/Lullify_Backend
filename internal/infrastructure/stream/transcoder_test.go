package stream

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestTranscodeFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "test.mp3")
	data := make([]byte, chunkSize*3)
	if err := os.WriteFile(audioPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("failed to create segmenter: %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	transcoder := NewTranscoder(segmenter)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- transcoder.TranscodeFile(ctx, audioPath)
	}()

	cancel()
	if transcodeErr := <-done; transcodeErr != nil {
		t.Fatalf("expected no error on canceled context, got %v", transcodeErr)
	}
}

func TestTranscodeFile_FileNotFound(t *testing.T) {
	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("failed to create segmenter: %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	transcoder := NewTranscoder(segmenter)
	if transcodeErr := transcoder.TranscodeFile(context.Background(), "/nonexistent/file.mp3"); transcodeErr == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestNewTranscoder(t *testing.T) {
	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("failed to create segmenter: %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	if transcoder := NewTranscoder(segmenter); transcoder == nil {
		t.Fatal("expected non-nil transcoder")
	}
}

func TestTranscodeFile_MultipleChunks(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "multi.mp3")
	data := make([]byte, chunkSize*5)
	if err := os.WriteFile(audioPath, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("failed to create segmenter: %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	transcoder := NewTranscoder(segmenter)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- transcoder.TranscodeFile(ctx, audioPath)
	}()

	cancel()
	if transcodeErr := <-done; transcodeErr != nil {
		t.Fatalf("expected no error, got %v", transcodeErr)
	}
}

func TestTranscodeFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "empty.mp3")
	if err := os.WriteFile(audioPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("failed to create segmenter: %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	transcoder := NewTranscoder(segmenter)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if transcodeErr := transcoder.TranscodeFile(ctx, audioPath); transcodeErr != nil {
		t.Fatalf("expected no error on empty file with canceled ctx, got %v", transcodeErr)
	}
}
