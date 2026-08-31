package playlist_test

import (
	"testing"

	"Lullify_Backend/internal/domain/playlist"
)

func TestFormat_IsValid_MP3(t *testing.T) {
	if !playlist.FormatMP3.IsValid() {
		t.Error("expected MP3 to be valid")
	}
}

func TestFormat_IsValid_FLAC(t *testing.T) {
	if !playlist.FormatFLAC.IsValid() {
		t.Error("expected FLAC to be valid")
	}
}

func TestFormat_IsValid_WAV(t *testing.T) {
	if !playlist.FormatWAV.IsValid() {
		t.Error("expected WAV to be valid")
	}
}

func TestFormat_IsValid_Unknown(t *testing.T) {
	f := playlist.Format("ogg")
	if f.IsValid() {
		t.Error("expected unknown format to be invalid")
	}
}
