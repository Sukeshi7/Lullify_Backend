package favorite_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	appfavorite "Lullify_Backend/internal/application/favorite"
	"Lullify_Backend/internal/domain/favorite"
)

type mockFavoriteRepo struct {
	save       func(ctx context.Context, f *favorite.Favorite) error
	delete     func(ctx context.Context, userID, streamID uuid.UUID) error
	findByUser func(ctx context.Context, userID uuid.UUID) ([]*favorite.Favorite, error)
}

func (m *mockFavoriteRepo) Save(ctx context.Context, f *favorite.Favorite) error {
	return m.save(ctx, f)
}
func (m *mockFavoriteRepo) Delete(ctx context.Context, userID, streamID uuid.UUID) error {
	return m.delete(ctx, userID, streamID)
}
func (m *mockFavoriteRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]*favorite.Favorite, error) {
	return m.findByUser(ctx, userID)
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestAdd_Success(t *testing.T) {
	repo := &mockFavoriteRepo{
		save: func(_ context.Context, _ *favorite.Favorite) error { return nil },
	}
	uc := appfavorite.NewAddUseCase(repo)

	f, err := uc.Execute(context.Background(), appfavorite.AddInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if f == nil {
		t.Fatal("expected favorite, got nil")
	}
}

func TestAdd_AlreadyFavorited(t *testing.T) {
	// La DB retourne une erreur unique constraint → wrappée en ErrAlreadyFavorited
	repo := &mockFavoriteRepo{
		save: func(_ context.Context, _ *favorite.Favorite) error {
			return favorite.ErrAlreadyFavorited
		},
	}
	uc := appfavorite.NewAddUseCase(repo)

	_, err := uc.Execute(context.Background(), appfavorite.AddInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_RepoError(t *testing.T) {
	repo := &mockFavoriteRepo{
		save: func(_ context.Context, _ *favorite.Favorite) error {
			return errors.New("db error")
		},
	}
	uc := appfavorite.NewAddUseCase(repo)

	_, err := uc.Execute(context.Background(), appfavorite.AddInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

// ── Remove ────────────────────────────────────────────────────────────────────

func TestRemove_Success(t *testing.T) {
	repo := &mockFavoriteRepo{
		delete: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}
	uc := appfavorite.NewRemoveUseCase(repo)

	err := uc.Execute(context.Background(), appfavorite.RemoveInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRemove_NotFound(t *testing.T) {
	repo := &mockFavoriteRepo{
		delete: func(_ context.Context, _, _ uuid.UUID) error {
			return favorite.ErrFavoriteNotFound
		},
	}
	uc := appfavorite.NewRemoveUseCase(repo)

	err := uc.Execute(context.Background(), appfavorite.RemoveInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if !errors.Is(err, favorite.ErrFavoriteNotFound) {
		t.Errorf("expected ErrFavoriteNotFound, got %v", err)
	}
}

func TestRemove_RepoError(t *testing.T) {
	repo := &mockFavoriteRepo{
		delete: func(_ context.Context, _, _ uuid.UUID) error {
			return errors.New("db error")
		},
	}
	uc := appfavorite.NewRemoveUseCase(repo)

	err := uc.Execute(context.Background(), appfavorite.RemoveInput{
		UserID:   uuid.New(),
		StreamID: uuid.New(),
	})

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestList_Success(t *testing.T) {
	favorites := []*favorite.Favorite{
		{ID: uuid.New(), UserID: uuid.New(), StreamID: uuid.New(), CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: uuid.New(), StreamID: uuid.New(), CreatedAt: time.Now()},
	}
	repo := &mockFavoriteRepo{
		findByUser: func(_ context.Context, _ uuid.UUID) ([]*favorite.Favorite, error) {
			return favorites, nil
		},
	}
	uc := appfavorite.NewListUseCase(repo)

	result, err := uc.Execute(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 favorites, got %d", len(result))
	}
}

func TestList_Empty(t *testing.T) {
	repo := &mockFavoriteRepo{
		findByUser: func(_ context.Context, _ uuid.UUID) ([]*favorite.Favorite, error) {
			return []*favorite.Favorite{}, nil
		},
	}
	uc := appfavorite.NewListUseCase(repo)

	result, err := uc.Execute(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 favorites, got %d", len(result))
	}
}

func TestList_RepoError(t *testing.T) {
	repo := &mockFavoriteRepo{
		findByUser: func(_ context.Context, _ uuid.UUID) ([]*favorite.Favorite, error) {
			return nil, errors.New("db error")
		},
	}
	uc := appfavorite.NewListUseCase(repo)

	_, err := uc.Execute(context.Background(), uuid.New())

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
