package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"Lullify_Backend/internal/domain/user"
	"Lullify_Backend/internal/infrastructure/token"
)

type AdminHandler struct {
	users  user.Repository
	tokens *token.JWTService
}

func NewAdminHandler(users user.Repository, tk *token.JWTService) *AdminHandler {
	return &AdminHandler{users: users, tokens: tk}
}

func (h *AdminHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/admin/users", h.ListUsers)
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}", h.DeleteUser)
	mux.HandleFunc("GET /api/v1/admin/stats", h.Stats)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}

	users, err := h.users.FindAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, userToAdminJSON(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out, "total": len(out)})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if id == claims.UserID {
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	if err := h.users.DeleteByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}

	ctx := r.Context()

	total, err := h.users.CountAll(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}
	admins, err := h.users.CountByRole(ctx, user.RoleAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}
	broadcasters, err := h.users.CountByRole(ctx, user.RoleBroadcaster)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}
	listeners, err := h.users.CountByRole(ctx, user.RoleUser)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total_users":  total,
		"admins":       admins,
		"broadcasters": broadcasters,
		"listeners":    listeners,
	})
}

func (h *AdminHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (*token.Claims, bool) {
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
	if claims.Role != user.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin access required")
		return nil, false
	}
	return claims, true
}

func userToAdminJSON(u *user.User) map[string]any {
	return map[string]any{
		"id":         u.ID.String(),
		"username":   u.Username,
		"email":      u.Email,
		"role":       string(u.Role),
		"created_at": u.CreatedAt.Format(time.RFC3339),
		"updated_at": u.UpdatedAt.Format(time.RFC3339),
	}
}

// MetricsMiddleware protège un handler avec une vérification admin
func (h *AdminHandler) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := h.requireAdmin(w, r); !ok {
			return
		}
		next.ServeHTTP(w, r)
	})
}
