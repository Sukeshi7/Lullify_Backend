package stream

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	// Taille d'un chunk audio ~2s à 128kbps = 32KB
	chunkSize = 32 * 1024
)

// Transcoder lit un fichier audio depuis le filesystem local
// et écrit les segments HLS dans le segmenter.
type Transcoder struct {
	segmenter *HLSSegmenter
}

func NewTranscoder(segmenter *HLSSegmenter) *Transcoder {
	return &Transcoder{segmenter: segmenter}
}

// TranscodeFile lit le fichier audio au chemin donné et génère les segments HLS.
// Bloque jusqu'à la fin du fichier ou l'annulation du context.
func (t *Transcoder) TranscodeFile(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening audio file %q: %w", filePath, err)
	}
	defer f.Close()

	buf := make([]byte, chunkSize)

	for {
		// Vérifie l'annulation du context
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		n, err := f.Read(buf)
		if n > 0 {
			// Écrit le chunk comme segment HLS
			if writeErr := t.segmenter.WriteSegment(buf[:n]); writeErr != nil {
				return fmt.Errorf("writing segment: %w", writeErr)
			}
		}

		if err == io.EOF {
			// Fin du fichier — on attend segmentDuration avant de recommencer
			// (comportement radio : on boucle sur le fichier)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(segmentDuration):
				// Seek au début pour boucler
				if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
					return fmt.Errorf("seeking to start: %w", seekErr)
				}
			}
			continue
		}

		if err != nil {
			return fmt.Errorf("reading audio file: %w", err)
		}

		// Simule la durée réelle du segment pour ne pas saturer
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(segmentDuration):
		}
	}
}
