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
	insertRefreshToken       = `INSERT INTO refresh_tokens (token_hash, user_id, account_id, email, expires_at, revoked_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	insertRefreshTokenByUser = `INSERT INTO refresh_tokens_by_user (user_id, token_hash, expires_at, revoked_at, created_at) VALUES (?, ?, ?, ?, ?)`
	selectTokenHashesByUser  = `SELECT token_hash FROM refresh_tokens_by_user WHERE user_id = ?`
	revokeAllRefreshTokens   = `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?`
	refreshTokenByHash       = `SELECT token_hash, user_id, account_id, email, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token_hash = ?`
	revokeRefreshToken       = `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ? IF revoked_at IS NULL`
)

func (r *RefreshTokenRepository) Insert(ctx context.Context, tokenHash string, userID uuid.UUID, accountID uuid.UUID, email string, expiresAt time.Time) error {
	now := time.Now().UTC()
	if err := r.session.Query(insertRefreshTokenByUser, userID, tokenHash, expiresAt, nil, now).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		Exec(); err != nil {
		return fmt.Errorf("index refresh token by user: %w", err)
	}
	return r.session.Query(insertRefreshToken, tokenHash, userID, accountID, email, expiresAt, nil, now).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		Exec()
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	var revokedAt *time.Time
	err := r.session.Query(refreshTokenByHash, tokenHash).
		WithContext(ctx).
		Consistency(gocql.One).
		Scan(&rt.TokenHash, &rt.UserID, &rt.AccountID, &rt.Email, &rt.ExpiresAt, &revokedAt, &rt.CreatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	rt.RevokedAt = revokedAt
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

// RevokeAllForUser revokes every refresh token issued to the user. It is used
// when a rotated (already-revoked) token is presented again, which indicates
// possible token theft; the whole family is invalidated as a precaution.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	iter := r.session.Query(selectTokenHashesByUser, userID).
		WithContext(ctx).
		Consistency(gocql.Quorum).Iter()
	var hashes []string
	for {
		var hash string
		if !iter.Scan(&hash) {
			break
		}
		hashes = append(hashes, hash)
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("list user refresh tokens: %w", err)
	}
	now := time.Now().UTC()
	for _, hash := range hashes {
		if err := r.session.Query(revokeAllRefreshTokens, now, hash).
			WithContext(ctx).
			Consistency(gocql.Quorum).Exec(); err != nil {
			return fmt.Errorf("revoke refresh token %s: %w", hash, err)
		}
	}
	return nil
}
