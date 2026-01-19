package repository

import (
	"chatter/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type RefreshTokenRepository struct {
	pg     *pgxpool.Pool
	logger *zap.Logger
}

func NewRefreshTokenRepository(pg *pgxpool.Pool, logger *zap.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		pg:     pg,
		logger: logger,
	}
}

func (r *RefreshTokenRepository) CreateRefreshTokenByHash(ctx context.Context, token *domain.RefreshToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pg.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Revoked)
	if err != nil {
		r.logger.Error("Failed to create refresh token", zap.Error(err))
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token domain.RefreshToken
	err := r.pg.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to get refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

func (r *RefreshTokenRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true, updated_at = now()
		WHERE id = $1
	`

	_, err := r.pg.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to revoke refresh token", zap.Error(err))
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
