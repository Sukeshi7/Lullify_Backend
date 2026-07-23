package stream

import (
	"context"
	"fmt"
	"sync"

	domain "Lullify_Backend/internal/domain/stream"
	"github.com/google/uuid"
)

const (
	listenerBufferSize = 8
)

type streamSession struct {
	cancel      context.CancelFunc
	segmenter   *HLSSegmenter
	subscribers map[<-chan domain.Chunk]chan domain.Chunk
	mu          sync.RWMutex
}

func (ss *streamSession) broadcast(chunk domain.Chunk) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for _, ch := range ss.subscribers {
		select {
		case ch <- chunk:
		default:
			// auditeur trop lent — on drop plutôt que bloquer
		}
	}
}

type StreamEngine struct {
	sessions map[uuid.UUID]*streamSession
	mu       sync.RWMutex
}

func NewStreamEngine() *StreamEngine {
	return &StreamEngine{
		sessions: make(map[uuid.UUID]*streamSession),
	}
}

func (e *StreamEngine) Start(ctx context.Context, streamID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.sessions[streamID]; exists {
		return domain.ErrStreamAlreadyLive
	}

	segmenter, err := NewHLSSegmenter(streamID)
	if err != nil {
		return fmt.Errorf("creating HLS segmenter: %w", err)
	}

	sessionCtx, cancel := context.WithCancel(ctx)

	session := &streamSession{
		cancel:      cancel,
		segmenter:   segmenter,
		subscribers: make(map[<-chan domain.Chunk]chan domain.Chunk),
	}

	e.sessions[streamID] = session

	go e.produce(sessionCtx, streamID, session)

	return nil
}

func (e *StreamEngine) produce(ctx context.Context, streamID uuid.UUID, session *streamSession) {
	defer func() {
		// Cleanup HLS
		session.segmenter.Cleanup()

		// Ferme tous les channels auditeurs
		session.mu.Lock()
		for _, ch := range session.subscribers {
			close(ch)
		}
		session.subscribers = make(map[<-chan domain.Chunk]chan domain.Chunk)
		session.mu.Unlock()

		// Supprime la session
		e.mu.Lock()
		delete(e.sessions, streamID)
		e.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// TODO Sprint 4 : lire chunks depuis Redis (file de lecture playlist)
			// Pour l'instant silence — la plomberie goroutine/channel est validée
		}
	}
}

func (e *StreamEngine) Stop(streamID uuid.UUID) error {
	e.mu.Lock()
	session, exists := e.sessions[streamID]
	e.mu.Unlock()

	if !exists {
		return domain.ErrStreamNotLive
	}

	session.cancel()
	return nil
}

func (e *StreamEngine) Subscribe(streamID uuid.UUID) (<-chan domain.Chunk, error) {
	e.mu.RLock()
	session, exists := e.sessions[streamID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("stream %s is not live", streamID)
	}

	ch := make(chan domain.Chunk, listenerBufferSize)

	session.mu.Lock()
	session.subscribers[ch] = ch
	session.mu.Unlock()

	return ch, nil
}

func (e *StreamEngine) Unsubscribe(streamID uuid.UUID, ch <-chan domain.Chunk) {
	e.mu.RLock()
	session, exists := e.sessions[streamID]
	e.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	if internalCh, ok := session.subscribers[ch]; ok {
		close(internalCh)
		delete(session.subscribers, ch)
	}
	session.mu.Unlock()
}

func (e *StreamEngine) IsRunning(streamID uuid.UUID) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, exists := e.sessions[streamID]
	return exists
}

func (e *StreamEngine) GetSegmenter(streamID uuid.UUID) (*HLSSegmenter, error) {
	e.mu.RLock()
	session, exists := e.sessions[streamID]
	e.mu.RUnlock()

	if !exists {
		return nil, domain.ErrStreamNotLive
	}
	return session.segmenter, nil
}
