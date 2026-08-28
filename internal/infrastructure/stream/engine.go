package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	domain "Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/infrastructure/observability"
)

const (
	listenerBufferSize = 8
)

type streamSession struct {
	cancel        context.CancelFunc
	segmenter     *HLSSegmenter
	audioFilePath string
	subscribers   map[<-chan domain.Chunk]chan domain.Chunk
	mu            sync.RWMutex
}

// broadcast envoie un chunk à tous les auditeurs connectés.
//
//nolint:unused
func (ss *streamSession) broadcast(chunk domain.Chunk) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for _, ch := range ss.subscribers {
		select {
		case ch <- chunk:
		default:
		}
	}
}

// Engine gère le cycle de vie des streams live
type Engine struct {
	sessions map[uuid.UUID]*streamSession
	mu       sync.RWMutex
}

func NewStreamEngine() *Engine {
	return &Engine{
		sessions: make(map[uuid.UUID]*streamSession),
	}
}

func (e *Engine) Start(ctx context.Context, streamID uuid.UUID, audioFilePath string) error {
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
		cancel:        cancel,
		segmenter:     segmenter,
		audioFilePath: audioFilePath,
		subscribers:   make(map[<-chan domain.Chunk]chan domain.Chunk),
	}

	e.sessions[streamID] = session

	go e.produce(sessionCtx, streamID, session)

	return nil
}

func (e *Engine) produce(ctx context.Context, streamID uuid.UUID, session *streamSession) {
	defer func() {
		if err := session.segmenter.Cleanup(); err != nil {
			_ = err
		}

		session.mu.Lock()
		abrupt := len(session.subscribers)
		for _, ch := range session.subscribers {
			close(ch)
			observability.StreamDisconnections.WithLabelValues("abrupt").Inc()
			observability.ActiveListeners.Dec()
		}
		if abrupt == 0 {
			observability.StreamDisconnections.WithLabelValues("normal").Inc()
		}
		session.subscribers = make(map[<-chan domain.Chunk]chan domain.Chunk)
		session.mu.Unlock()

		e.mu.Lock()
		delete(e.sessions, streamID)
		e.mu.Unlock()
	}()

	if session.audioFilePath != "" {
		transcoder := NewTranscoder(session.segmenter)
		if err := transcoder.TranscodeFile(ctx, session.audioFilePath); err != nil {
			log := observability.FromContext(ctx)
			log.Error().
				Err(err).
				Str("stream_id", streamID.String()).
				Msg("transcoder error")
		}
	} else {
		<-ctx.Done()
	}
}

func (e *Engine) Stop(streamID uuid.UUID) error {
	e.mu.Lock()
	session, exists := e.sessions[streamID]
	e.mu.Unlock()

	if !exists {
		return domain.ErrStreamNotLive
	}

	session.cancel()
	return nil
}

func (e *Engine) Subscribe(streamID uuid.UUID) (<-chan domain.Chunk, error) {
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

	observability.ActiveListeners.Inc()

	return ch, nil
}

func (e *Engine) Unsubscribe(streamID uuid.UUID, ch <-chan domain.Chunk) {
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
		observability.ActiveListeners.Dec()
		observability.StreamDisconnections.WithLabelValues("normal").Inc()
	}
	session.mu.Unlock()
}

func (e *Engine) IsRunning(streamID uuid.UUID) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, exists := e.sessions[streamID]
	return exists
}

func (e *Engine) GetSegmenter(streamID uuid.UUID) (*HLSSegmenter, error) {
	e.mu.RLock()
	session, exists := e.sessions[streamID]
	e.mu.RUnlock()

	if !exists {
		return nil, domain.ErrStreamNotLive
	}
	return session.segmenter, nil
}
