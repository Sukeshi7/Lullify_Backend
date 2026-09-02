package stream

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// Taille arbitraire pour les fichiers de test (l'ancien transcodeur utilisait
// une constante chunkSize, supprimée avec le passage à ffmpeg).
const testFileSize = 96 * 1024

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

// Context annulé avant le lancement de ffmpeg : le transcodeur doit retourner
// nil (arrêt propre), même si ffmpeg n'est pas disponible sur le runner CI.
func TestTranscodeFile_CanceledContext(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "test.mp3")
	data := make([]byte, testFileSize)
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
	cancel() // annulé immédiatement

	if transcodeErr := transcoder.TranscodeFile(ctx, audioPath); transcodeErr != nil {
		t.Fatalf("expected no error on canceled context, got %v", transcodeErr)
	}
}

// Idem avec un fichier vide.
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
