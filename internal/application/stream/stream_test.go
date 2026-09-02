package stream_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	appstream "Lullify_Backend/internal/application/stream"
	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/infrastructure/observability"
	"Lullify_Backend/internal/infrastructure/redis"
)

func init() {
	observability.InitLogger("test")
}

// ── Mocks partagés ────────────────────────────────────────────────────────────

type mockStreamRepo struct {
	findByID     func(ctx context.Context, id uuid.UUID) (*stream.Stream, error)
	create       func(ctx context.Context, s *stream.Stream) error
	updateStatus func(ctx context.Context, id uuid.UUID, status stream.Status) error
	findActive   func(ctx context.Context) ([]*stream.Stream, error)
}

func (m *mockStreamRepo) FindByID(ctx context.Context, id uuid.UUID) (*stream.Stream, error) {
	return m.findByID(ctx, id)
}
func (m *mockStreamRepo) Create(ctx context.Context, s *stream.Stream) error {
	return m.create(ctx, s)
}
func (m *mockStreamRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status stream.Status) error {
	return m.updateStatus(ctx, id, status)
}
func (m *mockStreamRepo) FindActive(ctx context.Context) ([]*stream.Stream, error) {
	return m.findActive(ctx)
}
func (m *mockStreamRepo) IncrementListeners(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockStreamRepo) DecrementListeners(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockStreamRepo) Delete(ctx context.Context, id uuid.UUID) error             { return nil }

type mockEngine struct {
	start   func(ctx context.Context, id uuid.UUID, filePath string) error
	stop    func(id uuid.UUID) error
	running map[uuid.UUID]bool
}

func (m *mockEngine) Start(ctx context.Context, id uuid.UUID, filePath string) error {
	return m.start(ctx, id, filePath)
}
func (m *mockEngine) Stop(id uuid.UUID) error { return m.stop(id) }
func (m *mockEngine) Subscribe(id uuid.UUID) (<-chan stream.Chunk, error) {
	return make(chan stream.Chunk), nil
}
func (m *mockEngine) Unsubscribe(id uuid.UUID, ch <-chan stream.Chunk) {}
func (m *mockEngine) IsRunning(id uuid.UUID) bool                      { return m.running[id] }

type mockQueue struct {
	pop func(ctx context.Context, streamID string) (*redis.TrackJob, error)
}

func (m *mockQueue) Pop(ctx context.Context, streamID string) (*redis.TrackJob, error) {
	return m.pop(ctx, streamID)
}

func newLiveStream(ownerID uuid.UUID) *stream.Stream {
	return &stream.Stream{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Title:   "Test Stream",
		Status:  stream.StatusLive,
	}
}

func newOfflineStream(ownerID uuid.UUID) *stream.Stream {
	return &stream.Stream{
		ID:         uuid.New(),
		OwnerID:    ownerID,
		Title:      "Test Stream",
		MountPoint: "test-mount",
		Status:     stream.StatusOffline,
	}
}

var _ = time.Now

// ── CreateUseCase ─────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	repo := &mockStreamRepo{
		create: func(_ context.Context, s *stream.Stream) error { return nil },
	}
	uc := appstream.NewCreateUseCase(repo)

	s, err := uc.Execute(context.Background(), appstream.CreateInput{
		OwnerID:    uuid.New(),
		Title:      "My Stream",
		MountPoint: "my-stream",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s == nil {
		t.Fatal("expected stream, got nil")
	}
	if s.Title != "My Stream" {
		t.Errorf("expected title 'My Stream', got %s", s.Title)
	}
	if s.Status != stream.StatusOffline {
		t.Errorf("expected status offline, got %s", s.Status)
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	repo := &mockStreamRepo{}
	uc := appstream.NewCreateUseCase(repo)

	_, err := uc.Execute(context.Background(), appstream.CreateInput{
		OwnerID:    uuid.New(),
		Title:      "",
		MountPoint: "mount",
	})

	if !errors.Is(err, stream.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreate_EmptyMountPoint(t *testing.T) {
	repo := &mockStreamRepo{}
	uc := appstream.NewCreateUseCase(repo)

	_, err := uc.Execute(context.Background(), appstream.CreateInput{
		OwnerID:    uuid.New(),
		Title:      "My Stream",
		MountPoint: "",
	})

	if err == nil {
		t.Fatal("expected error for empty mount point, got nil")
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := &mockStreamRepo{
		create: func(_ context.Context, _ *stream.Stream) error {
			return errors.New("db error")
		},
	}
	uc := appstream.NewCreateUseCase(repo)

	_, err := uc.Execute(context.Background(), appstream.CreateInput{
		OwnerID:    uuid.New(),
		Title:      "My Stream",
		MountPoint: "mount",
	})

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

// ── StopUseCase ───────────────────────────────────────────────────────────────

func TestStop_Success(t *testing.T) {
	ownerID := uuid.New()
	s := newLiveStream(ownerID)

	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{stop: func(_ uuid.UUID) error { return nil }}

	uc := appstream.NewStopUseCase(repo, engine)
	if err := uc.Execute(context.Background(), appstream.StopInput{StreamID: s.ID, OwnerID: ownerID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStop_NotFound(t *testing.T) {
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return nil, nil },
	}
	uc := appstream.NewStopUseCase(repo, &mockEngine{})

	err := uc.Execute(context.Background(), appstream.StopInput{StreamID: uuid.New(), OwnerID: uuid.New()})
	if !errors.Is(err, stream.ErrStreamNotFound) {
		t.Errorf("expected ErrStreamNotFound, got %v", err)
	}
}

func TestStop_NotOwner(t *testing.T) {
	s := newLiveStream(uuid.New())
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
	}
	uc := appstream.NewStopUseCase(repo, &mockEngine{})

	err := uc.Execute(context.Background(), appstream.StopInput{StreamID: s.ID, OwnerID: uuid.New()})
	if !errors.Is(err, stream.ErrNotStreamOwner) {
		t.Errorf("expected ErrNotStreamOwner, got %v", err)
	}
}

func TestStop_NotLive(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
	}
	uc := appstream.NewStopUseCase(repo, &mockEngine{})

	err := uc.Execute(context.Background(), appstream.StopInput{StreamID: s.ID, OwnerID: ownerID})
	if !errors.Is(err, stream.ErrStreamNotLive) {
		t.Errorf("expected ErrStreamNotLive, got %v", err)
	}
}

func TestStop_EngineError(t *testing.T) {
	ownerID := uuid.New()
	s := newLiveStream(ownerID)
	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{stop: func(_ uuid.UUID) error { return errors.New("engine error") }}
	uc := appstream.NewStopUseCase(repo, engine)

	err := uc.Execute(context.Background(), appstream.StopInput{StreamID: s.ID, OwnerID: ownerID})
	if err == nil {
		t.Fatal("expected error from engine, got nil")
	}
}

// ── StartUseCase ──────────────────────────────────────────────────────────────

func TestStart_Success(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)

	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{
		start: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
		stop:  func(_ uuid.UUID) error { return nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}

	uc := appstream.NewStartUseCase(repo, engine, queue)
	if err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStart_Success_WithTrack(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)

	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{
		start: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
		stop:  func(_ uuid.UUID) error { return nil },
	}
	queue := &mockQueue{
		pop: func(_ context.Context, _ string) (*redis.TrackJob, error) {
			return &redis.TrackJob{FilePath: "/audio/track.mp3"}, nil
		},
	}

	uc := appstream.NewStartUseCase(repo, engine, queue)
	if err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStart_QueueError(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)

	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{
		start: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
		stop:  func(_ uuid.UUID) error { return nil },
	}
	queue := &mockQueue{
		pop: func(_ context.Context, _ string) (*redis.TrackJob, error) {
			return nil, errors.New("redis error")
		},
	}

	uc := appstream.NewStartUseCase(repo, engine, queue)
	if err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID}); err != nil {
		t.Fatalf("expected no error despite queue error, got %v", err)
	}
}

func TestStart_NotFound(t *testing.T) {
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return nil, nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}
	uc := appstream.NewStartUseCase(repo, &mockEngine{}, queue)

	err := uc.Execute(context.Background(), appstream.StartInput{StreamID: uuid.New(), OwnerID: uuid.New()})
	if !errors.Is(err, stream.ErrStreamNotFound) {
		t.Errorf("expected ErrStreamNotFound, got %v", err)
	}
}

func TestStart_NotOwner(t *testing.T) {
	s := newOfflineStream(uuid.New())
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}
	uc := appstream.NewStartUseCase(repo, &mockEngine{}, queue)

	err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: uuid.New()})
	if !errors.Is(err, stream.ErrNotStreamOwner) {
		t.Errorf("expected ErrNotStreamOwner, got %v", err)
	}
}

func TestStart_AlreadyLive(t *testing.T) {
	ownerID := uuid.New()
	s := newLiveStream(ownerID)
	repo := &mockStreamRepo{
		findByID: func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}
	uc := appstream.NewStartUseCase(repo, &mockEngine{}, queue)

	err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID})
	if !errors.Is(err, stream.ErrStreamAlreadyLive) {
		t.Errorf("expected ErrStreamAlreadyLive, got %v", err)
	}
}

func TestStart_EngineError(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)
	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return nil },
	}
	engine := &mockEngine{
		start: func(_ context.Context, _ uuid.UUID, _ string) error { return errors.New("engine error") },
		stop:  func(_ uuid.UUID) error { return nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}
	uc := appstream.NewStartUseCase(repo, engine, queue)

	err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID})
	if err == nil {
		t.Fatal("expected error from engine, got nil")
	}
}

func TestStart_UpdateStatusError(t *testing.T) {
	ownerID := uuid.New()
	s := newOfflineStream(ownerID)
	repo := &mockStreamRepo{
		findByID:     func(_ context.Context, _ uuid.UUID) (*stream.Stream, error) { return s, nil },
		updateStatus: func(_ context.Context, _ uuid.UUID, _ stream.Status) error { return errors.New("db error") },
	}
	engine := &mockEngine{
		start: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
		stop:  func(_ uuid.UUID) error { return nil },
	}
	queue := &mockQueue{pop: func(_ context.Context, _ string) (*redis.TrackJob, error) { return nil, nil }}
	uc := appstream.NewStartUseCase(repo, engine, queue)

	err := uc.Execute(context.Background(), appstream.StartInput{StreamID: s.ID, OwnerID: ownerID})
	if err == nil {
		t.Fatal("expected error from updateStatus, got nil")
	}
}
