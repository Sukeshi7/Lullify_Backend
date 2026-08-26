package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	appfavorite "Lullify_Backend/internal/application/favorite"
	"Lullify_Backend/internal/domain/favorite"
	"Lullify_Backend/internal/infrastructure/token"
)

type FavoriteHandler struct {
	add    *appfavorite.AddUseCase
	remove *appfavorite.RemoveUseCase
	list   *appfavorite.ListUseCase
	tokens *token.JWTService
}

func NewFavoriteHandler(
	add *appfavorite.AddUseCase,
	remove *appfavorite.RemoveUseCase,
	list *appfavorite.ListUseCase,
	tk *token.JWTService,
) *FavoriteHandler {
	return &FavoriteHandler{add: add, remove: remove, list: list, tokens: tk}
}

func (h *FavoriteHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/favorites", h.Add)
	mux.HandleFunc("DELETE /api/v1/favorites/{streamId}", h.Remove)
	mux.HandleFunc("GET /api/v1/favorites", h.ListMine)
}

// --- Add ---

type addFavoriteBody struct {
	StreamID string `json:"stream_id"`
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body addFavoriteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	streamID, err := uuid.Parse(body.StreamID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	f, err := h.add.Execute(r.Context(), appfavorite.AddInput{
		UserID:   claims.UserID,
		StreamID: streamID,
	})
	if err != nil {
		if errors.Is(err, favorite.ErrAlreadyFavorited) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to add favorite")
		return
	}

	writeJSON(w, http.StatusCreated, favoriteToJSON(f))
}

// --- Remove ---

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	streamID, err := uuid.Parse(r.PathValue("streamId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid stream id")
		return
	}

	if err := h.remove.Execute(r.Context(), appfavorite.RemoveInput{
		UserID:   claims.UserID,
		StreamID: streamID,
	}); err != nil {
		if errors.Is(err, favorite.ErrFavoriteNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove favorite")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- ListMine ---

func (h *FavoriteHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	favorites, err := h.list.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch favorites")
		return
	}

	out := make([]map[string]any, 0, len(favorites))
	for _, f := range favorites {
		out = append(out, favoriteToJSON(f))
	}
	writeJSON(w, http.StatusOK, map[string]any{"favorites": out, "total": len(out)})
}

// --- helpers ---

func (h *FavoriteHandler) authenticate(r *http.Request) (*token.Claims, bool) {
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

func favoriteToJSON(f *favorite.Favorite) map[string]any {
	return map[string]any{
		"id":         f.ID.String(),
		"user_id":    f.UserID.String(),
		"stream_id":  f.StreamID.String(),
		"created_at": f.CreatedAt.Format(time.RFC3339),
	}
}