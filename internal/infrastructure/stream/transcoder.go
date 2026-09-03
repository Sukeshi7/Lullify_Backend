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

	// Vérifie si le contexte est déjà annulé avant de lancer ffmpeg
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
	defer pipeR.Close()
	defer pipeW.Close()

	// ── Producteur : lit les fichiers en boucle et sort du PCM brut ──
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

	// ── Consommateur : lit le PCM et génère le HLS ──
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

	if err := producer.Start(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("starting producer: %w", err)
	}
	if err := consumer.Start(); err != nil {
		producer.Process.Kill() //nolint:errcheck
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("starting consumer: %w", err)
	}

	go func() {
		<-ctx.Done()
		producer.Process.Kill() //nolint:errcheck
		pipeW.Close()
	}()

	producerErr := producer.Wait()
	pipeW.Close()
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
