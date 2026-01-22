package repository

import (
	"chatter/internal/domain"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type AuthRepository struct {
	pg     *pgxpool.Pool
	logger *zap.Logger
}

func NewAuthRepository(pg *pgxpool.Pool, logger *zap.Logger) *AuthRepository {
	return &AuthRepository{
		pg:     pg,
		logger: logger,
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email
	`

	row := r.pg.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
	); err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, bool) {
	query := `
		SELECT id, username, email, password_hash
		FROM users
		WHERE username = $1
	`

	var user domain.User

	row := r.pg.QueryRow(ctx, query, username)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	); err != nil {
		r.logger.Error("Failed to get user by username", zap.Error(err))
		return nil, false
	}

	return &user, true
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uint64) (*domain.User, bool) {
	query := `
		SELECT id, username, email, password_hash
		FROM users
		WHERE id = $1
	`

	var user domain.User

	row := r.pg.QueryRow(ctx, query, id)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	); err != nil {
		r.logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, false
	}

	return &user, true
}
