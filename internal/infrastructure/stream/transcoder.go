package stream

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Transcoder struct {
	segmenter *HLSSegmenter
}

func NewTranscoder(segmenter *HLSSegmenter) *Transcoder {
	return &Transcoder{segmenter: segmenter}
}

func (t *Transcoder) TranscodeFile(ctx context.Context, filePath string) error {
	return t.TranscodeFiles(ctx, []string{filePath})
}

func (t *Transcoder) TranscodeFiles(ctx context.Context, filePaths []string) error {
	if len(filePaths) == 0 {
		<-ctx.Done()
		return nil
	}

	if ctx.Err() != nil {
		return nil
	}

	dir := t.segmenter.dir

	listFile := dir + "/playlist_input.txt"
	content := ""
	for _, f := range filePaths {
		content += "file '" + f + "'\n"
	}
	if err := os.WriteFile(listFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing concat list: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-re",
		"-f", "concat",
		"-safe", "0",
		"-stream_loop", "-1",
		"-i", listFile,
		"-vn",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "6",
		"-hls_flags", "delete_segments+omit_endlist",
		"-hls_segment_filename", dir+"/segment%06d.ts",
		dir+"/playlist.m3u8",
	)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	return nil
}

func buildInputLabels(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("[%d:a]", i)
	}
	return result
}
