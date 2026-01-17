package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"chatter/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrEmptyCredentials = errors.New("missing credentials")
)

type UserStore interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Get(ctx context.Context, username string) (*domain.User, bool)
}

type TokenManager interface {
	Generate(userID uint64, username string) (string, error)
	Parse(token string) (string, error)
}

type Service struct {
	store  UserStore
	tokens TokenManager
}

func NewService(store UserStore, tokens TokenManager) *Service {
	return &Service{store: store, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, username, password string) (*domain.User, string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, "", ErrEmptyCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	user, err := s.store.Create(ctx, &domain.User{Username: username, PasswordHash: hash})
	if err != nil {
		return nil, "", ErrUserExists
	}

	token, err := s.tokens.Generate(user.ID, username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, "", ErrEmptyCredentials
	}

	user, ok := s.store.Get(ctx, username)
	if !ok {
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, "", ErrInvalidCreds
	}

	token, err := s.tokens.Generate(user.ID, username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}
