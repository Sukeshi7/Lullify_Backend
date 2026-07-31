package stream

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	segmentDuration = 2 * time.Second
	windowSize      = 6
	segmentsDir     = "/tmp/lullify"
)

type HLSSegmenter struct {
	streamID uuid.UUID
	dir      string

	mu       sync.Mutex
	segments []string
	seq      int
}

func NewHLSSegmenter(streamID uuid.UUID) (*HLSSegmenter, error) {
	dir := filepath.Join(segmentsDir, streamID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating segments dir: %w", err)
	}
	return &HLSSegmenter{
		streamID: streamID,
		dir:      dir,
		segments: make([]string, 0, windowSize),
	}, nil
}

func (h *HLSSegmenter) WriteSegment(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	segName := fmt.Sprintf("segment%06d.ts", h.seq)
	segPath := filepath.Join(h.dir, segName)

	if err := os.WriteFile(segPath, data, 0644); err != nil {
		return fmt.Errorf("writing segment: %w", err)
	}

	h.segments = append(h.segments, segName)
	h.seq++

	if len(h.segments) > windowSize {
		oldest := h.segments[0]
		h.segments = h.segments[1:]
		_ = os.Remove(filepath.Join(h.dir, oldest))
	}

	return h.writePlaylist()
}

func (h *HLSSegmenter) Playlist() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	path := filepath.Join(h.dir, "playlist.m3u8")
	return os.ReadFile(path)
}

func (h *HLSSegmenter) Cleanup() error {
	return os.RemoveAll(h.dir)
}

func (h *HLSSegmenter) writePlaylist() error {
	firstSeq := h.seq - len(h.segments)

	content := "#EXTM3U\n"
	content += "#EXT-X-VERSION:3\n"
	content += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(segmentDuration.Seconds()))
	content += fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", firstSeq)

	for _, seg := range h.segments {
		content += fmt.Sprintf("#EXTINF:%.1f,\n", segmentDuration.Seconds())
		content += seg + "\n"
	}

	path := filepath.Join(h.dir, "playlist.m3u8")
	return os.WriteFile(path, []byte(content), 0644)
}
