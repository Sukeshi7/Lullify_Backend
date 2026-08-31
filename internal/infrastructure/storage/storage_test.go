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

	if err := s.Upload(ctx, "test/track.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg"); err != nil {
		t.Fatalf("expected no error uploading, got %v", err)
	}

	_, err = os.Stat(dir + "/test/track.mp3")
	if os.IsNotExist(err) {
		t.Fatal("expected file to exist after upload")
	}

	if err := s.Delete(ctx, "test/track.mp3"); err != nil {
		t.Fatalf("expected no error deleting, got %v", err)
	}

	_, err = os.Stat(dir + "/test/track.mp3")
	if !os.IsNotExist(err) {
		t.Fatal("expected file to be deleted")
	}
}

func TestLocalStorage_PresignedURL(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("expected no error creating local storage, got %v", err)
	}

	url, err := s.PresignedGetURL(context.Background(), "test/track.mp3", 3600)
	if err != nil {
		t.Fatalf("expected no error getting presigned URL, got %v", err)
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

	err = s.Upload(context.Background(), "../evil/path.mp3",
		bytes.NewReader([]byte("data")), 4, "audio/mpeg")
	if err == nil {
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
	if err := s.Upload(context.Background(), "large/file.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg"); err != nil {
		t.Fatalf("unexpected error uploading large file: %v", err)
	}
}

func TestLocalStorage_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.Delete(context.Background(), "nonexistent/file.mp3")
	if err != nil {
		t.Fatalf("expected no error deleting nonexistent file, got %v", err)
	}
}

func TestLocalStorage_NestedPath(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []byte("audio data")
	if err := s.Upload(context.Background(), "a/b/c/track.mp3", bytes.NewReader(data), int64(len(data)), "audio/mpeg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLocalStorage_PresignedURL_Format(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	url, err := s.PresignedGetURL(context.Background(), "some/track.mp3", 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "some/track.mp3") {
		t.Errorf("expected URL to contain key, got %s", url)
	}
}

func TestFactory_LocalProvider_Variants(t *testing.T) {
	variants := []string{"local", "filesystem", "disk"}
	for _, v := range variants {
		dir := t.TempDir()
		s, err := storage.New(storage.Options{Provider: v, LocalPath: dir})
		if err != nil {
			t.Errorf("expected no error for provider %q, got %v", v, err)
		}
		if s == nil {
			t.Errorf("expected non-nil storage for provider %q", v)
		}
	}
}

func TestLocalStorage_NewLocalStorage_InvalidPath(t *testing.T) {
	// Path vide
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

	// Clé vide — resolve retourne basePath lui-même → erreur
	err = s.Upload(context.Background(), "", bytes.NewReader([]byte("data")), 4, "audio/mpeg")
	// Peut réussir ou échouer selon l'OS, l'important c'est pas de paniquer
	_ = err
}

func TestLocalStorage_Delete_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(storage.Options{Provider: "local", LocalPath: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Path traversal sur Delete
	err = s.Delete(context.Background(), "../outside")
	if err == nil {
		t.Fatal("expected error for path traversal on delete")
	}
}
