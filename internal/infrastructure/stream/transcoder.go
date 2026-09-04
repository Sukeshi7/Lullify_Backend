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
	playlistPath := dir + "/playlist.m3u8"
	segmentPattern := dir + "/segment%06d.ts"

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating pipe: %w", err)
	}
	defer func() { _ = pipeR.Close() }()

	// ── Args producteur ──────────────────────────────────────────────────────
	baseArgs := []string{
		"-hide_banner", "-loglevel", "error",
		"-re",
	}

	for _, fp := range filePaths {
		baseArgs = append(baseArgs, "-i", fp)
	}

	if len(filePaths) > 1 {
		filterComplex := fmt.Sprintf("%s concat=n=%d:v=0:a=1[aout]",
			buildInputLabels(len(filePaths)),
			len(filePaths))
		baseArgs = append(baseArgs,
			"-filter_complex", filterComplex,
			"-map", "[aout]",
		)
	}

	baseArgs = append(baseArgs,
		"-vn",
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "2",
		"pipe:1",
	)

	// ── Producteur en boucle ─────────────────────────────────────────────────
	go func() {
		defer func() { _ = pipeW.Close() }()
		for {
			if ctx.Err() != nil {
				return
			}
			cmd := exec.CommandContext(ctx, "ffmpeg", baseArgs...)
			cmd.Stdout = pipeW
			if err := cmd.Run(); err != nil {
				if ctx.Err() != nil {
					return
				}
				// Erreur ffmpeg — on continue à boucler
			}
		}
	}()

	// ── Consommateur : PCM → HLS ─────────────────────────────────────────────
	consumer := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "2",
		"-i", "pipe:0",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "6",
		"-hls_flags", "delete_segments+omit_endlist",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	)
	consumer.Stdin = pipeR

	if err := consumer.Start(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("starting consumer: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = consumer.Process.Kill()
	}()

	if err := consumer.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("consumer error: %w", err)
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
