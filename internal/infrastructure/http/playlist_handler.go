package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/playlist"
	"Lullify_Backend/internal/infrastructure/token"
)

type PlaylistHandler struct {
	playlists playlist.Repository
	tracks    playlist.TrackRepository
	tokens    *token.JWTService
}

func NewPlaylistHandler(p playlist.Repository, t playlist.TrackRepository, tk *token.JWTService) *PlaylistHandler {
	return &PlaylistHandler{playlists: p, tracks: t, tokens: tk}
}

func (h *PlaylistHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/playlists", h.Create)
	mux.HandleFunc("GET /api/v1/playlists", h.ListMine)
}

// --- Create ---

type createPlaylistBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

func (h *PlaylistHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body createPlaylistBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, playlist.ErrEmptyTitle.Error())
		return
	}

	now := time.Now().UTC()
	p := &playlist.Playlist{
		ID:          uuid.New(),
		OwnerID:     claims.UserID,
		Title:       title,
		Description: strings.TrimSpace(body.Description),
		IsPublic:    body.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.playlists.Create(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create playlist")
		return
	}

	writeJSON(w, http.StatusCreated, playlistToJSON(p))
}

// --- ListMine ---

func (h *PlaylistHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.playlists.FindByOwner(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch playlists")
		return
	}

	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, playlistToJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": out, "total": len(out)})
}

// --- helpers ---

func (h *PlaylistHandler) authenticate(r *http.Request) (*token.Claims, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, false
	}
	claims, err := h.tokens.ParseAccess(strings.TrimPrefix(auth, "Bearer "))
	if err != nil || errors.Is(err, token.ErrInvalidToken) {
		return nil, false
	}
	return claims, true
}

func playlistToJSON(p *playlist.Playlist) map[string]any {
	return map[string]any{
		"id":          p.ID.String(),
		"owner_id":    p.OwnerID.String(),
		"title":       p.Title,
		"description": p.Description,
		"is_public":   p.IsPublic,
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}
