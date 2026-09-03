package stream

import (
	"Lullify_Backend/internal/infrastructure/observability"
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
	defer func() { _ = pipeW.Close() }()

	producerArgs := []string{
		"-hide_banner", "-loglevel", "error",
		"-re",
	}

	for _, fp := range filePaths {
		producerArgs = append(producerArgs, "-i", fp)
	}

	if len(filePaths) > 1 {
		filterComplex := fmt.Sprintf("%s concat=n=%d:v=0:a=1[aout]",
			buildInputLabels(len(filePaths)),
			len(filePaths))
		producerArgs = append(producerArgs,
			"-filter_complex", filterComplex,
			"-map", "[aout]",
		)
	}

	producerArgs = append(producerArgs,
		"-vn",
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "2",
		"-stream_loop", "-1",
		"pipe:1",
	)

	producer := exec.CommandContext(ctx, "ffmpeg", producerArgs...)
	producer.Stdout = pipeW

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

	observability.Logger.Info().
		Strs("producer_args", producerArgs).
		Msg("ffmpeg producer command")

	if err := producer.Start(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("starting producer: %w", err)
	}
	if err := consumer.Start(); err != nil {
		_ = producer.Process.Kill()
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("starting consumer: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = producer.Process.Kill()
		_ = pipeW.Close()
	}()

	producerErr := producer.Wait()
	_ = pipeW.Close()
	consumerErr := consumer.Wait()

	if ctx.Err() != nil {
		return nil
	}

	if producerErr != nil {
		return fmt.Errorf("producer error: %w", producerErr)
	}
	if consumerErr != nil {
		return fmt.Errorf("consumer error: %w", consumerErr)
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
