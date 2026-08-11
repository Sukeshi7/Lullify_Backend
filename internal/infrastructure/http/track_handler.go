package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	apptrack "Lullify_Backend/internal/application/track"
	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/infrastructure/token"
)

type TrackHandler struct {
	upload  *apptrack.UploadUseCase
	tokens  *token.JWTService
	maxSize int64
}

func NewTrackHandler(upload *apptrack.UploadUseCase, tokens *token.JWTService, maxSize int64) *TrackHandler {
	return &TrackHandler{upload: upload, tokens: tokens, maxSize: maxSize}
}

func (h *TrackHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/playlists/{id}/tracks", h.Upload)
}

func (h *TrackHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	playlistID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid playlist id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize+1024)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large or invalid multipart form")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	artist := strings.TrimSpace(r.FormValue("artist"))
	formatStr := strings.TrimSpace(strings.ToLower(r.FormValue("format")))

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()

	track, err := h.upload.Execute(r.Context(), apptrack.UploadInput{
		PlaylistID: playlistID,
		UploaderID: claims.UserID,
		Title:      title,
		Artist:     artist,
		Format:     playlist.Format(formatStr),
		SizeBytes:  header.Size,
		Reader:     file,
	})
	if err != nil {
		writeError(w, statusForUploadError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, trackToJSON(track))
}

func (h *TrackHandler) authenticate(r *http.Request) (*token.Claims, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, false
	}
	claims, err := h.tokens.ParseAccess(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return nil, false
	}
	return claims, true
}

func statusForUploadError(err error) int {
	switch {
	case errors.Is(err, playlist.ErrEmptyTitle),
		errors.Is(err, playlist.ErrInvalidFormat),
		errors.Is(err, playlist.ErrEmptyFile),
		errors.Is(err, playlist.ErrInvalidAudioFile):
		return http.StatusBadRequest
	case errors.Is(err, playlist.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, playlist.ErrPlaylistNotFound):
		return http.StatusNotFound
	case errors.Is(err, playlist.ErrNotOwner):
		return http.StatusForbidden
	case errors.Is(err, playlist.ErrStorageFailure):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func trackToJSON(t *playlist.Track) map[string]any {
	return map[string]any{
		"id":           t.ID.String(),
		"playlist_id":  t.PlaylistID.String(),
		"title":        t.Title,
		"artist":       t.Artist,
		"file_path":    t.FilePath,
		"format":       string(t.Format),
		"size_bytes":   t.SizeBytes,
		"duration":     t.Duration,
		"position":     t.Position,
		"uploaded_by":  t.UploadedBy.String(),
		"created_at":   t.CreatedAt.Format(time.RFC3339),
		"updated_at":   t.UpdatedAt.Format(time.RFC3339),
	}
}