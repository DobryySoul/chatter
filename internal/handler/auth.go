package handler

import (
	"chatter/internal/domain"
	"chatter/internal/usecase"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	service AuthService
}

type AuthService interface {
	Register(ctx context.Context, username, password string) (*domain.User, string, string, error)
	Login(ctx context.Context, username, password string) (string, string, *domain.User, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, *domain.User, error)
}

func NewHandler(service AuthService) *Handler {
	return &Handler{service: service}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	ID       uint64 `json:"id,omitempty"`
	Token    string `json:"token"`
	Username string `json:"username,omitempty"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := h.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case usecase.ErrEmptyCredentials:
			http.Error(w, "missing credentials", http.StatusBadRequest)
		case usecase.ErrUserExists:
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			http.Error(w, "failed to register", http.StatusInternalServerError)
		}

		return
	}

	h.setRefreshCookie(w, refreshToken)
	writeJSON(w, authResponse{
		ID:       user.ID,
		Token:    accessToken,
		Username: user.Username,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, user, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case usecase.ErrEmptyCredentials:
			http.Error(w, "missing credentials", http.StatusBadRequest)
		case usecase.ErrInvalidCreds:
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "failed to login", http.StatusInternalServerError)
		}
		return
	}

	h.setRefreshCookie(w, refreshToken)
	writeJSON(w, authResponse{
		ID:       user.ID,
		Token:    accessToken,
		Username: user.Username,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, user, err := h.service.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		// при ошибке сбрасываем куку
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	h.setRefreshCookie(w, refreshToken)
	writeJSON(w, authResponse{
		ID:       user.ID,
		Token:    accessToken,
		Username: user.Username,
	})
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(60 * 24 * time.Hour),
	})
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
