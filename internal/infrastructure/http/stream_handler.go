package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	appstream "Lullify_Backend/internal/application/stream"
	"Lullify_Backend/internal/domain/stream"
	infrastream "Lullify_Backend/internal/infrastructure/stream"
	"Lullify_Backend/internal/infrastructure/token"

	"github.com/google/uuid"
)

type StreamHandler struct {
	create *appstream.CreateUseCase
	start  *appstream.StartUseCase
	stop   *appstream.StopUseCase
	repo   stream.Repository
	tokens *token.JWTService
	engine *infrastream.Engine
}

func NewStreamHandler(
	create *appstream.CreateUseCase,
	start *appstream.StartUseCase,
	stop *appstream.StopUseCase,
	repo stream.Repository,
	tokens *token.JWTService,
	engine *infrastream.Engine,
) *StreamHandler {
	return &StreamHandler{
		create: create,
		start:  start,
		stop:   stop,
		repo:   repo,
		tokens: tokens,
		engine: engine,
	}
}

func (h *StreamHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/streams", h.ListActive)
	mux.HandleFunc("POST /api/v1/streams", h.Create)
	mux.HandleFunc("POST /api/v1/streams/{id}/start", h.Start)
	mux.HandleFunc("POST /api/v1/streams/{id}/stop", h.Stop)
	mux.HandleFunc("GET /streams/{id}/playlist.m3u8", h.HLSPlaylist)
	mux.HandleFunc("GET /streams/{id}/{segment}", h.HLSSegment)
}

func (h *StreamHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	streams, err := h.repo.FindActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch streams")
		return
	}

	type streamResponse struct {
		ID            string  `json:"id"`
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		MountPoint    string  `json:"mount_point"`
		Status        string  `json:"status"`
		ListenerCount int     `json:"listener_count"`
		StartedAt     *string `json:"started_at,omitempty"`
	}

	result := make([]streamResponse, 0, len(streams))
	for _, s := range streams {
		var startedAt *string
		if s.StartedAt != nil {
			t := s.StartedAt.Format("2006-01-02T15:04:05Z")
			startedAt = &t
		}
		result = append(result, streamResponse{
			ID:            s.ID.String(),
			Title:         s.Title,
			Description:   s.Description,
			MountPoint:    s.MountPoint,
			Status:        string(s.Status),
			ListenerCount: s.ListenerCount,
			StartedAt:     startedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"streams": result,
		"total":   len(result),
	})
}

func (h *StreamHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID, err := h.ownerIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		MountPoint  string `json:"mount_point"`
	}
	// decodeErr pour éviter le shadow sur err ligne précédente
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	s, err := h.create.Execute(r.Context(), appstream.CreateInput{
		OwnerID:     ownerID,
		Title:       body.Title,
		Description: body.Description,
		MountPoint:  body.MountPoint,
	})
	if err != nil {
		writeError(w, statusForStreamError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"stream": map[string]string{
			"id":          s.ID.String(),
			"title":       s.Title,
			"description": s.Description,
			"mount_point": s.MountPoint,
			"status":      string(s.Status),
		},
	})
}

func (h *StreamHandler) Start(w http.ResponseWriter, r *http.Request) {
	ownerID, err := h.ownerIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	if err := h.start.Execute(r.Context(), appstream.StartInput{
		StreamID: streamID,
		OwnerID:  ownerID,
	}); err != nil {
		writeError(w, statusForStreamError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

func (h *StreamHandler) Stop(w http.ResponseWriter, r *http.Request) {
	ownerID, err := h.ownerIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	if err := h.stop.Execute(r.Context(), appstream.StopInput{
		StreamID: streamID,
		OwnerID:  ownerID,
	}); err != nil {
		writeError(w, statusForStreamError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "offline"})
}

func (h *StreamHandler) HLSPlaylist(w http.ResponseWriter, r *http.Request) {
	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	segmenter, err := h.engine.GetSegmenter(streamID)
	if err != nil {
		writeError(w, http.StatusNotFound, "stream not found or not live")
		return
	}

	playlist, err := segmenter.Playlist()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "playlist not available")
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	if _, writeErr := w.Write(playlist); writeErr != nil {
		_ = writeErr
	}
}

func (h *StreamHandler) HLSSegment(w http.ResponseWriter, r *http.Request) {
	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	segName := r.PathValue("segment")
	segPath := fmt.Sprintf("/tmp/lullify/%s/%s", streamID, segName)

	data, err := os.ReadFile(segPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "segment not found")
		return
	}

	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Cache-Control", "no-cache")
	if _, writeErr := w.Write(data); writeErr != nil {
		_ = writeErr
	}
}

func (h *StreamHandler) ownerIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return uuid.Nil, errors.New("missing token")
	}
	claims, err := h.tokens.ParseAccess(authHeader[7:])
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func statusForStreamError(err error) int {
	switch {
	case errors.Is(err, stream.ErrStreamNotFound):
		return http.StatusNotFound
	case errors.Is(err, stream.ErrStreamAlreadyLive):
		return http.StatusConflict
	case errors.Is(err, stream.ErrStreamNotLive):
		return http.StatusConflict
	case errors.Is(err, stream.ErrNotStreamOwner):
		return http.StatusForbidden
	case errors.Is(err, stream.ErrEmptyTitle):
		return http.StatusBadRequest
	case errors.Is(err, stream.ErrMountPointTaken):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
