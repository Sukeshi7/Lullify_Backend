package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	apphistory "Lullify_Backend/internal/application/history"
	"Lullify_Backend/internal/domain/history"
	"Lullify_Backend/internal/infrastructure/token"
)

type HistoryHandler struct {
	record *apphistory.RecordUseCase
	list   *apphistory.ListUseCase
	tokens *token.JWTService
}

func NewHistoryHandler(record *apphistory.RecordUseCase, list *apphistory.ListUseCase, tk *token.JWTService) *HistoryHandler {
	return &HistoryHandler{record: record, list: list, tokens: tk}
}

func (h *HistoryHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/history", h.Record)
	mux.HandleFunc("GET /api/v1/history", h.ListMine)
}

// --- Record ---

type recordHistoryBody struct {
	TrackTitle string `json:"track_title"`
	Artist     string `json:"artist"`
	StreamID   string `json:"stream_id"`
}

func (h *HistoryHandler) Record(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body recordHistoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	input := apphistory.RecordInput{
		UserID:     claims.UserID,
		TrackTitle: body.TrackTitle,
		Artist:     body.Artist,
	}
	if body.StreamID != "" {
		if sid, err := uuid.Parse(body.StreamID); err == nil {
			input.StreamID = &sid
		}
	}

	entry, err := h.record.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, history.ErrEmptyTitle) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to record history")
		return
	}

	writeJSON(w, http.StatusCreated, historyToJSON(entry))
}

// --- ListMine ---

func (h *HistoryHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entries, err := h.list.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, historyToJSON(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": out, fieldTotal: len(out)})
}

// --- helpers ---

func (h *HistoryHandler) authenticate(r *http.Request) (*token.Claims, bool) {
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

func historyToJSON(e *history.Entry) map[string]any {
	out := map[string]any{
		"id":          e.ID.String(),
		"user_id":     e.UserID.String(),
		"track_title": e.TrackTitle,
		"artist":      e.Artist,
		"played_at":   e.PlayedAt.Format(time.RFC3339),
	}
	if e.StreamID != nil {
		out["stream_id"] = e.StreamID.String()
	}
	return out
}
