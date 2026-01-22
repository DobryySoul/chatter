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
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, device_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pg.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Revoked, token.DeviceID)
	if err != nil {
		r.logger.Error("Failed to create refresh token", zap.Error(err))
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeUserDeviceTokens(ctx context.Context, userID uint64, deviceID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true, updated_at = now()
		WHERE user_id = $1 AND device_id = $2 AND revoked = false
	`

	_, err := r.pg.Exec(ctx, query, userID, deviceID)
	if err != nil {
		r.logger.Error("Failed to revoke user device tokens", zap.Error(err))
		return fmt.Errorf("failed to revoke user device tokens: %w", err)
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

func (r *RefreshTokenRepository) ListActiveSessions(ctx context.Context, userID uint64) ([]domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, expires_at, revoked, updated_at, device_id
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked = false AND expires_at > now()
		ORDER BY updated_at DESC
	`

	rows, err := r.pg.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to list sessions", zap.Error(err))
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.RefreshToken
	for rows.Next() {
		var token domain.RefreshToken
		if err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.ExpiresAt,
			&token.Revoked,
			&token.UpdatedAt,
			&token.DeviceID,
		); err != nil {
			r.logger.Error("Failed to scan session", zap.Error(err))
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, token)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Failed to read sessions", zap.Error(err))
		return nil, fmt.Errorf("failed to read sessions: %w", err)
	}

	return sessions, nil
}
