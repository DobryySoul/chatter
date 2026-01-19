package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"chatter/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrEmptyCredentials = errors.New("missing credentials")
	ErrInvalidToken     = errors.New("invalid token")
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, bool)
	GetUserByID(ctx context.Context, id uint64) (*domain.User, bool)
}

type RefreshTokenRepository interface {
	CreateRefreshTokenByHash(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
}

type TokenManager interface {
	GenerateAccessToken(userID uint64, username string) (string, error)
	GenerateRefreshToken() (string, error)
	ParseAccessToken(token string) (string, uint64, error)
}

type AuthService struct {
	authStore  AuthRepository
	tokenStore RefreshTokenRepository
	tokens     TokenManager
	refreshTTL time.Duration
}

func NewAuthService(authStore AuthRepository, tokenStore RefreshTokenRepository, tokens TokenManager, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		authStore:  authStore,
		tokenStore: tokenStore,
		tokens:     tokens,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*domain.User, string, string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, "", "", ErrEmptyCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	user, err := s.authStore.CreateUser(ctx, &domain.User{Username: username, PasswordHash: hash})
	if err != nil {
		return nil, "", "", ErrUserExists
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID, username)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, *domain.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", "", nil, ErrEmptyCredentials
	}

	user, ok := s.authStore.GetUserByUsername(ctx, username)
	if !ok {
		return "", "", nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCreds
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID, username)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return accessToken, refreshToken, user, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, *domain.User, error) {
	tokenHash := hashToken(refreshToken)

	token, err := s.tokenStore.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return "", "", nil, ErrInvalidToken
	}

	if token.Revoked {
		return "", "", nil, ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return "", "", nil, ErrInvalidToken
	}

	if err := s.tokenStore.RevokeRefreshToken(ctx, token.ID); err != nil {
		return "", "", nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	user, ok := s.authStore.GetUserByID(ctx, token.UserID)
	if !ok {
		return "", "", nil, ErrInvalidToken
	}

	accessToken, newRefreshToken, err := s.generateTokens(ctx, user.ID, user.Username)
	if err != nil {
		return "", "", nil, fmt.Errorf("")
	}

	return accessToken, newRefreshToken, user, nil
}

func (s *AuthService) generateTokens(ctx context.Context, userID uint64, username string) (string, string, error) {
	accessToken, err := s.tokens.GenerateAccessToken(userID, username)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenHash := hashToken(refreshToken)
	expiresAt := time.Now().Add(s.refreshTTL)

	err = s.tokenStore.CreateRefreshTokenByHash(ctx, &domain.RefreshToken{
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
