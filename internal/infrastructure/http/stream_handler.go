package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	appstream "Lullify_Backend/internal/application/stream"
	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/domain/user"
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
	mux.HandleFunc("GET /api/v1/streams/mine", h.ListMine)
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
	writeJSON(w, http.StatusOK, map[string]any{
		"streams":  h.streamsToJSON(streams),
		fieldTotal: len(streams),
	})
}

// ListMine retourne les streams de l'utilisateur connecté (pour restaurer l'état broadcaster)
func (h *StreamHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	all, err := h.repo.FindActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch streams")
		return
	}

	mine := make([]*stream.Stream, 0)
	for _, s := range all {
		if s.OwnerID == claims.UserID {
			mine = append(mine, s)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"streams":  h.streamsToJSON(mine),
		fieldTotal: len(mine),
	})
}

func (h *StreamHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	if claims.Role != user.RoleBroadcaster && claims.Role != user.RoleAdmin {
		writeError(w, http.StatusForbidden, "broadcaster role required")
		return
	}

	// Vérifie qu'il n'a pas déjà un stream actif
	all, err := h.repo.FindActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check existing streams")
		return
	}
	for _, s := range all {
		if s.OwnerID == claims.UserID {
			writeError(w, http.StatusConflict, "you already have an active stream")
			return
		}
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		MountPoint  string `json:"mount_point"`
	}
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	s, err := h.create.Execute(r.Context(), appstream.CreateInput{
		OwnerID:     claims.UserID,
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
			fieldStatus:   string(s.Status),
		},
	})
}

func (h *StreamHandler) Start(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	if claims.Role != user.RoleBroadcaster && claims.Role != user.RoleAdmin {
		writeError(w, http.StatusForbidden, "broadcaster role required")
		return
	}

	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	if err := h.start.Execute(r.Context(), appstream.StartInput{
		StreamID: streamID,
		OwnerID:  claims.UserID,
	}); err != nil {
		writeError(w, statusForStreamError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{fieldStatus: "live"})
}

func (h *StreamHandler) Stop(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	streamID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	if err := h.stop.Execute(r.Context(), appstream.StopInput{
		StreamID: streamID,
		OwnerID:  claims.UserID,
	}); err != nil {
		writeError(w, statusForStreamError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{fieldStatus: "offline"})
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

// requireAuth extrait et valide le token Bearer, retourne les claims et true si OK.
func (h *StreamHandler) requireAuth(w http.ResponseWriter, r *http.Request) (*token.Claims, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}
	claims, err := h.tokens.ParseAccess(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}
	return claims, true
}

func (h *StreamHandler) streamsToJSON(streams []*stream.Stream) []map[string]any {
	result := make([]map[string]any, 0, len(streams))
	for _, s := range streams {
		entry := map[string]any{
			"id":             s.ID.String(),
			"title":          s.Title,
			"description":    s.Description,
			"mount_point":    s.MountPoint,
			fieldStatus:      string(s.Status),
			"listener_count": s.ListenerCount,
		}
		if s.StartedAt != nil {
			entry["started_at"] = s.StartedAt.Format("2006-01-02T15:04:05Z")
		}
		result = append(result, entry)
	}
	return result
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
	default:
		return http.StatusInternalServerError
	}
}
