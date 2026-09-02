package history_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	apphistory "Lullify_Backend/internal/application/history"
	"Lullify_Backend/internal/domain/history"
)

type mockHistoryRepo struct {
	save       func(ctx context.Context, e *history.Entry) error
	findByUser func(ctx context.Context, userID uuid.UUID, limit int) ([]*history.Entry, error)
}

func (m *mockHistoryRepo) Save(ctx context.Context, e *history.Entry) error {
	return m.save(ctx, e)
}

func (m *mockHistoryRepo) FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*history.Entry, error) {
	return m.findByUser(ctx, userID, limit)
}

func TestRecord_Success(t *testing.T) {
	repo := &mockHistoryRepo{
		save: func(_ context.Context, _ *history.Entry) error { return nil },
	}
	uc := apphistory.NewRecordUseCase(repo)

	entry, err := uc.Execute(context.Background(), apphistory.RecordInput{
		UserID:     uuid.New(),
		TrackTitle: "Chill Lofi",
		Artist:     "DJ Sleepy",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
		return
	}
	if entry.TrackTitle != "Chill Lofi" {
		t.Errorf("expected title 'Chill Lofi', got %s", entry.TrackTitle)
	}
}

func TestRecord_EmptyTitle(t *testing.T) {
	repo := &mockHistoryRepo{}
	uc := apphistory.NewRecordUseCase(repo)

	_, err := uc.Execute(context.Background(), apphistory.RecordInput{
		UserID:     uuid.New(),
		TrackTitle: "",
	})

	if !errors.Is(err, history.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestRecord_WithStreamID(t *testing.T) {
	repo := &mockHistoryRepo{
		save: func(_ context.Context, _ *history.Entry) error { return nil },
	}
	uc := apphistory.NewRecordUseCase(repo)
	streamID := uuid.New()

	entry, err := uc.Execute(context.Background(), apphistory.RecordInput{
		UserID:     uuid.New(),
		TrackTitle: "Vaporwave Dreams",
		StreamID:   &streamID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.StreamID == nil || *entry.StreamID != streamID {
		t.Error("expected stream ID to be set")
	}
}

func TestRecord_RepoError(t *testing.T) {
	repo := &mockHistoryRepo{
		save: func(_ context.Context, _ *history.Entry) error {
			return errors.New("db error")
		},
	}
	uc := apphistory.NewRecordUseCase(repo)

	_, err := uc.Execute(context.Background(), apphistory.RecordInput{
		UserID:     uuid.New(),
		TrackTitle: "Test Track",
	})

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestList_Success(t *testing.T) {
	entries := []*history.Entry{
		{ID: uuid.New(), UserID: uuid.New(), TrackTitle: "Track 1", PlayedAt: time.Now()},
		{ID: uuid.New(), UserID: uuid.New(), TrackTitle: "Track 2", PlayedAt: time.Now()},
	}
	repo := &mockHistoryRepo{
		findByUser: func(_ context.Context, _ uuid.UUID, _ int) ([]*history.Entry, error) {
			return entries, nil
		},
	}
	uc := apphistory.NewListUseCase(repo)

	result, err := uc.Execute(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
	if result[0].TrackTitle != "Track 1" {
		t.Errorf("expected TrackTitle 'Track 1', got %s", result[0].TrackTitle)
	}
}

func TestList_Empty(t *testing.T) {
	repo := &mockHistoryRepo{
		findByUser: func(_ context.Context, _ uuid.UUID, _ int) ([]*history.Entry, error) {
			return []*history.Entry{}, nil
		},
	}
	uc := apphistory.NewListUseCase(repo)

	result, err := uc.Execute(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestList_RepoError(t *testing.T) {
	repo := &mockHistoryRepo{
		findByUser: func(_ context.Context, _ uuid.UUID, _ int) ([]*history.Entry, error) {
			return nil, errors.New("db error")
		},
	}
	uc := apphistory.NewListUseCase(repo)

	_, err := uc.Execute(context.Background(), uuid.New())

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
