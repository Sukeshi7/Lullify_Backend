package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if strings.TrimSpace(basePath) == "" {
		return nil, fmt.Errorf("local storage path is empty")
	}
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("creating storage dir %q: %w", basePath, err)
	}
	return &LocalStorage{basePath: filepath.Clean(basePath)}, nil
}

func (s *LocalStorage) resolve(key string) (string, error) {
	dest := filepath.Join(s.basePath, filepath.FromSlash(key))
	if dest != s.basePath && !strings.HasPrefix(dest, s.basePath+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid storage key %q", key)
	}
	return dest, nil
}

func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	dest, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating dir for %q: %w", key, err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", key, err)
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err = io.Copy(f, reader); err != nil {
		return fmt.Errorf("writing file %q: %w", key, err)
	}
	return nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	dest, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err = os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting %q: %w", key, err)
	}
	return nil
}

func (s *LocalStorage) PresignedGetURL(ctx context.Context, key string, expirySeconds int) (string, error) {
	return "/files/" + key, nil
}
