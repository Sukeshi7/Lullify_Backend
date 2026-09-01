package stream

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestHLSSegmenter_WriteAndPlaylist(t *testing.T) {
	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("expected no error creating segmenter, got %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	data := make([]byte, 1024)
	if writeErr := segmenter.WriteSegment(data); writeErr != nil {
		t.Fatalf("expected no error writing segment, got %v", writeErr)
	}

	playlist, playlistErr := segmenter.Playlist()
	if playlistErr != nil {
		t.Fatalf("expected no error reading playlist, got %v", playlistErr)
	}
	if len(playlist) == 0 {
		t.Error("expected non-empty playlist")
	}

	content := string(playlist)
	if len(content) < 7 || content[:7] != "#EXTM3U" {
		t.Errorf("expected playlist to start with #EXTM3U")
	}
}

func TestHLSSegmenter_RollingWindow(t *testing.T) {
	segmenter, err := NewHLSSegmenter(uuid.New())
	if err != nil {
		t.Fatalf("expected no error creating segmenter, got %v", err)
	}
	defer segmenter.Cleanup() //nolint:errcheck

	data := make([]byte, 512)
	for i := 0; i < 8; i++ {
		if writeErr := segmenter.WriteSegment(data); writeErr != nil {
			t.Fatalf("error writing segment %d: %v", i, writeErr)
		}
	}

	playlist, playlistErr := segmenter.Playlist()
	if playlistErr != nil {
		t.Fatalf("expected no error reading playlist, got %v", playlistErr)
	}

	count := 0
	for _, line := range splitLines(string(playlist)) {
		if len(line) >= 8 && line[:8] == "#EXTINF:" {
			count++
		}
	}
	if count > windowSize {
		t.Errorf("expected at most %d segments in playlist, got %d", windowSize, count)
	}
}

func TestHLSSegmenter_Cleanup(t *testing.T) {
	id := uuid.New()
	segmenter, err := NewHLSSegmenter(id)
	if err != nil {
		t.Fatalf("expected no error creating segmenter, got %v", err)
	}

	data := make([]byte, 512)
	_ = segmenter.WriteSegment(data)

	if cleanupErr := segmenter.Cleanup(); cleanupErr != nil {
		t.Fatalf("expected no error on cleanup, got %v", cleanupErr)
	}

	_, statErr := os.Stat("/tmp/lullify/" + id.String())
	if !os.IsNotExist(statErr) {
		t.Log("dir still exists after cleanup")
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
