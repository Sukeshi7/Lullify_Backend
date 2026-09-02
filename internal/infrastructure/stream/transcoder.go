package stream

import (
	"context"
	"fmt"
	"os/exec"
)

// Transcoder invoque ffmpeg pour produire un VRAI flux HLS
// (segments MPEG-TS valides + playlist.m3u8) à partir d'un fichier audio local.
// Boucle en continu sur le fichier — comportement radio.
type Transcoder struct {
	segmenter *HLSSegmenter
}

func NewTranscoder(segmenter *HLSSegmenter) *Transcoder {
	return &Transcoder{segmenter: segmenter}
}

// TranscodeFile lance ffmpeg. ffmpeg écrit lui-même playlist.m3u8 et
// segmentNNNNNN.ts dans le dossier du segmenter. Bloque jusqu'à
// l'annulation du context (fin du direct).
func (t *Transcoder) TranscodeFile(ctx context.Context, filePath string) error {
	dir := t.segmenter.dir
	playlistPath := dir + "/playlist.m3u8"
	segmentPattern := dir + "/segment%06d.ts"

	// -re            : lecture au débit réel (comportement radio/live)
	// -stream_loop -1: boucle le fichier à l'infini
	// -c:a aac       : transcode en AAC, le codec standard pour HLS
	// -f hls         : génère de VRAIS segments MPEG-TS + m3u8 valides
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-re",
		"-stream_loop", "-1",
		"-i", filePath,
		"-vn",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "6",
		"-hls_flags", "delete_segments+omit_endlist",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	)

	if err := cmd.Run(); err != nil {
		// Si le context a été annulé (Stop du direct), c'est un arrêt normal.
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}
	return nil
}
