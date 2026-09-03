package stream

import (
	"context"
	"fmt"
	"os/exec"
)

type Transcoder struct {
	segmenter    *HLSSegmenter
	segmentCount int
}

func NewTranscoder(segmenter *HLSSegmenter) *Transcoder {
	return &Transcoder{segmenter: segmenter}
}

func (t *Transcoder) TranscodeFile(ctx context.Context, filePath string) error {
	return t.transcodeOne(ctx, filePath)
}

func (t *Transcoder) transcodeOne(ctx context.Context, filePath string) error {
	dir := t.segmenter.dir
	playlistPath := dir + "/playlist.m3u8"
	segmentPattern := dir + "/segment%06d.ts"

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-re",
		"-i", filePath,
		"-vn",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "6",
		"-hls_flags", "delete_segments+omit_endlist+append_list",
		"-start_number", fmt.Sprintf("%d", t.segmentCount),
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	t.segmentCount += 300

	return nil
}
