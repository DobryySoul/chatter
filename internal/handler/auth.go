package handler

import (
	"chatter/internal/domain"
	"chatter/internal/usecase"
	"context"
	"encoding/json"
	"net/http"
)

type Handler struct {
	service AuthService
}

type AuthService interface {
	Register(ctx context.Context, username, password string) (*domain.User, string, error)
	Login(ctx context.Context, username, password string) (*domain.User, string, error)
}

func NewHandler(service AuthService) *Handler {
	return &Handler{service: service}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	ID       uint64 `json:"id"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, token, err := h.service.Register(r.Context(), req.Username, req.Password)
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

	writeJSON(w, authResponse{ID: user.ID, Token: token, Username: user.Username})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Username, req.Password)
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

	writeJSON(w, authResponse{ID: user.ID, Token: token, Username: user.Username})
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
