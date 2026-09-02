package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/favorite"
	"Lullify_Backend/internal/domain/history"
	"Lullify_Backend/internal/domain/stream"
	"Lullify_Backend/internal/domain/user"
	"Lullify_Backend/internal/infrastructure/observability"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type application/json")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "bad request")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatusForError(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{user.ErrEmailAlreadyExists, http.StatusConflict},
		{user.ErrEmptyEmail, http.StatusBadRequest},
	}

	for _, c := range cases {
		got := statusForError(c.err)
		if got != c.expected {
			t.Errorf("statusForError(%v) = %d, want %d", c.err, got, c.expected)
		}
	}
}

func TestStatusForStreamError(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{stream.ErrStreamNotFound, http.StatusNotFound},
		{stream.ErrStreamAlreadyLive, http.StatusConflict},
		{stream.ErrStreamNotLive, http.StatusConflict},
		{stream.ErrNotStreamOwner, http.StatusForbidden},
		{stream.ErrEmptyTitle, http.StatusBadRequest},
	}

	for _, c := range cases {
		got := statusForStreamError(c.err)
		if got != c.expected {
			t.Errorf("statusForStreamError(%v) = %d, want %d", c.err, got, c.expected)
		}
	}
}

func TestUserToAdminJSON(t *testing.T) {
	u := &user.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      user.RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := userToAdminJSON(u)

	if result["id"] != u.ID.String() {
		t.Errorf("expected id %s, got %s", u.ID.String(), result["id"])
	}
	if result["username"] != u.Username {
		t.Errorf("expected username %s, got %s", u.Username, result["username"])
	}
	if result["email"] != u.Email {
		t.Errorf("expected email %s, got %s", u.Email, result["email"])
	}
}

func TestFavoriteToJSON(t *testing.T) {
	f := &favorite.Favorite{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		StreamID:  uuid.New(),
		CreatedAt: time.Now(),
	}

	result := favoriteToJSON(f)

	if result["id"] != f.ID.String() {
		t.Errorf("expected id %s, got %s", f.ID.String(), result["id"])
	}
	if result["stream_id"] != f.StreamID.String() {
		t.Errorf("expected stream_id %s, got %s", f.StreamID.String(), result["stream_id"])
	}
}

func TestHistoryToJSON(t *testing.T) {
	e := &history.Entry{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		TrackTitle: "Chill Track",
		Artist:     "DJ Lo",
		PlayedAt:   time.Now(),
	}

	result := historyToJSON(e)

	if result["track_title"] != e.TrackTitle {
		t.Errorf("expected track_title %s, got %s", e.TrackTitle, result["track_title"])
	}
	if result["artist"] != e.Artist {
		t.Errorf("expected artist %s, got %s", e.Artist, result["artist"])
	}
}

func TestCorsMiddleware(t *testing.T) {
	allowed := []string{"http://localhost:3000"}
	handler := corsMiddleware(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Requête avec origine autorisée
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("expected CORS header to be set for allowed origin")
	}
}

func TestCorsMiddleware_Preflight(t *testing.T) {
	allowed := []string{"http://localhost:3000"}
	handler := corsMiddleware(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestCorsMiddleware_UnknownOrigin(t *testing.T) {
	allowed := []string{"http://localhost:3000"}
	handler := corsMiddleware(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header for unknown origin")
	}
}
func TestOtelMiddleware(t *testing.T) {
	observability.InitLogger("test")

	handler := observability.OtelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
