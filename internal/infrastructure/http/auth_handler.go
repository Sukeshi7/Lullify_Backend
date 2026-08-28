package http

import (
	"encoding/json"
	"errors"
	"net/http"

	appuser "Lullify_Backend/internal/application/user"
	"Lullify_Backend/internal/domain/user"
	"Lullify_Backend/internal/infrastructure/token"
)

const (
	fieldStatus = "status"
	fieldTotal  = "total"
)

type AuthHandler struct {
	register *appuser.RegisterUseCase
	login    *appuser.LoginUseCase
	users    user.Repository
	tokens   *token.JWTService
}

func NewAuthHandler(r *appuser.RegisterUseCase, l *appuser.LoginUseCase, users user.Repository, t *token.JWTService) *AuthHandler {
	return &AuthHandler{register: r, login: l, users: users, tokens: t}
}

func (h *AuthHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
}

type credentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	u, err := h.register.Execute(r.Context(), appuser.RegisterInput{
		Username: c.Username, Email: c.Email, Password: c.Password,
	})
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	h.sendTokens(w, http.StatusCreated, u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	u, err := h.login.Execute(r.Context(), appuser.LoginInput{
		Email: c.Email, Password: c.Password,
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.sendTokens(w, http.StatusOK, u)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	claims, err := h.tokens.ParseRefresh(body.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	u, err := h.users.FindByID(r.Context(), claims.UserID)
	if err != nil || u == nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}
	h.sendTokens(w, http.StatusOK, u)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) sendTokens(w http.ResponseWriter, status int, u *user.User) {
	access, refresh, err := h.tokens.GenerateTokens(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	writeJSON(w, status, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"user": map[string]string{
			"id":       u.ID.String(),
			"username": u.Username,
			"email":    u.Email,
			"role":     string(u.Role),
		},
	})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
