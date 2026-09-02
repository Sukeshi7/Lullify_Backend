package storage_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"Lullify_Backend/internal/infrastructure/storage"
)

func TestLocalStorage_UploadAndDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("expected no error creating local storage, got %v", err)
	}

	ctx := context.Background()
	data := []byte("test audio data")

	uploadErr := s.Upload(ctx, "test/track.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg")
	if uploadErr != nil {
		t.Fatalf("expected no error uploading, got %v", uploadErr)
	}

	_, statErr := os.Stat(dir + "/test/track.mp3")
	if os.IsNotExist(statErr) {
		t.Fatal("expected file to exist after upload")
	}

	deleteErr := s.Delete(ctx, "test/track.mp3")
	if deleteErr != nil {
		t.Fatalf("expected no error deleting, got %v", deleteErr)
	}

	_, statErr2 := os.Stat(dir + "/test/track.mp3")
	if !os.IsNotExist(statErr2) {
		t.Fatal("expected file to be deleted")
	}
}

func TestLocalStorage_PresignedURL(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("expected no error creating local storage, got %v", err)
	}

	url, urlErr := s.PresignedGetURL(context.Background(), "test/track.mp3", 3600)
	if urlErr != nil {
		t.Fatalf("expected no error getting presigned URL, got %v", urlErr)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestLocalStorage_InvalidPath(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("expected no error creating local storage, got %v", err)
	}

	pathErr := s.Upload(context.Background(), "../evil/path.mp3",
		bytes.NewReader([]byte("data")), 4, "audio/mpeg")
	if pathErr == nil {
		t.Fatal("expected error for path traversal attempt, got nil")
	}
}

func TestFactory_UnknownProvider(t *testing.T) {
	_, err := storage.New(storage.Options{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestFactory_EmptyLocalPath(t *testing.T) {
	_, err := storage.New(storage.Options{Provider: "local", LocalPath: ""})
	if err == nil {
		t.Fatal("expected error for empty local path, got nil")
	}
}

func TestLocalStorage_UploadLargeFile(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := bytes.Repeat([]byte("a"), 1024*1024)
	if uploadErr := s.Upload(context.Background(), "large/file.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg"); uploadErr != nil {
		t.Fatalf("unexpected error uploading large file: %v", uploadErr)
	}
}

func TestLocalStorage_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deleteErr := s.Delete(context.Background(), "nonexistent/file.mp3"); deleteErr != nil {
		t.Fatalf("expected no error deleting nonexistent file, got %v", deleteErr)
	}
}

func TestLocalStorage_NestedPath(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []byte("audio data")
	if uploadErr := s.Upload(context.Background(), "a/b/c/track.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg"); uploadErr != nil {
		t.Fatalf("unexpected error: %v", uploadErr)
	}
}

func TestLocalStorage_PresignedURL_Format(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	url, urlErr := s.PresignedGetURL(context.Background(), "some/track.mp3", 3600)
	if urlErr != nil {
		t.Fatalf("unexpected error: %v", urlErr)
	}
	if !strings.Contains(url, "some/track.mp3") {
		t.Errorf("expected URL to contain key, got %s", url)
	}
}

func TestFactory_LocalProvider_Variants(t *testing.T) {
	variants := []string{"local", "filesystem", "disk"}
	for _, v := range variants {
		dir := t.TempDir()
		s, newErr := storage.New(storage.Options{Provider: v, LocalPath: dir})
		if newErr != nil {
			t.Errorf("expected no error for provider %q, got %v", v, newErr)
		}
		if s == nil {
			t.Errorf("expected non-nil storage for provider %q", v)
		}
	}
}

func TestLocalStorage_NewLocalStorage_InvalidPath(t *testing.T) {
	_, err := storage.New(storage.Options{Provider: "local", LocalPath: "   "})
	if err == nil {
		t.Fatal("expected error for whitespace-only path")
	}
}

func TestLocalStorage_Upload_ExactBasePath(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = s.Upload(context.Background(), "", bytes.NewReader([]byte("data")), 4, "audio/mpeg")
}

func TestLocalStorage_Delete_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deleteErr := s.Delete(context.Background(), "../outside"); deleteErr == nil {
		t.Fatal("expected error for path traversal on delete")
	}
}
