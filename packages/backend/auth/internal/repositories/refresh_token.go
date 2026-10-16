package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
)

type RefreshTokenRepository struct {
	session *gocql.Session
}

func NewRefreshTokenRepository(session *gocql.Session) *RefreshTokenRepository {
	return &RefreshTokenRepository{session: session}
}

const (
	insertRefreshToken = `INSERT INTO refresh_tokens (token_hash, user_id, account_id, email, expires_at, revoked_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	refreshTokenByHash = `SELECT token_hash, user_id, account_id, email, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token_hash = ?`
	revokeRefreshToken = `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ? IF revoked_at IS NULL`
)

func (r *RefreshTokenRepository) Insert(ctx context.Context, tokenHash string, userID uuid.UUID, accountID uuid.UUID, email string, expiresAt time.Time) error {
	now := time.Now().UTC()
	return r.session.Query(insertRefreshToken, tokenHash, userID, accountID, email, expiresAt, nil, now).
		WithContext(ctx).
		Consistency(gocql.One).
		Exec()
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	var revokedAt time.Time
	var hasRevoked bool
	err := r.session.Query(refreshTokenByHash, tokenHash).
		WithContext(ctx).
		Consistency(gocql.One).
		Scan(&rt.TokenHash, &rt.UserID, &rt.AccountID, &rt.Email, &rt.ExpiresAt, &revokedAt, &hasRevoked, &rt.CreatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if hasRevoked {
		rt.RevokedAt = &revokedAt
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()
	applied, err := r.session.Query(revokeRefreshToken, now, tokenHash).
		WithContext(ctx).
		Consistency(gocql.One).
		SerialConsistency(gocql.Serial).
		ScanCAS()
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if !applied {
		return errors.ErrInvalidRefreshToken
	}
	return nil
}
