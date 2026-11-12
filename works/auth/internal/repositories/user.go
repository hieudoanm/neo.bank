package repositories

import (
	"context"
	"fmt"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/auth/internal/models"
)

type UserRepository struct {
	session *gocql.Session
}

func NewUserRepository(session *gocql.Session) *UserRepository {
	return &UserRepository{session: session}
}

const (
	insertEmailLWT      = `INSERT INTO users_by_email (email, user_id, created_at) VALUES (?, ?, ?) IF NOT EXISTS`
	insertUser          = `INSERT INTO users (user_id, email, account_id, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	selectUserIDByEmail = `SELECT user_id FROM users_by_email WHERE email = ?`
	selectByID          = `SELECT user_id, email, account_id, password_hash, created_at, updated_at FROM users WHERE user_id = ?`
	deleteEmail         = `DELETE FROM users_by_email WHERE email = ?`
	deleteUser          = `DELETE FROM users WHERE user_id = ?`
	insertAudit         = `INSERT INTO audit_events (service, day, timestamp, event_id, actor_user_id, action, resource_type, resource_id, correlation_id, metadata) VALUES ('auth', ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func (r *UserRepository) Create(ctx context.Context, email string, accountID uuid.UUID, passwordHash string) (*models.User, error) {
	now := time.Now().UTC()
	userID := newUUID()

	applied, err := r.session.Query(insertEmailLWT, email, gocql.UUID(userID), now).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		SerialConsistency(gocql.Serial).
		ScanCAS()
	if err != nil {
		return nil, fmt.Errorf("check email uniqueness: %w", err)
	}
	if !applied {
		return nil, errors.ErrEmailAlreadyRegistered
	}

	if err := r.session.Query(insertUser, gocql.UUID(userID), email, gocql.UUID(accountID), passwordHash, now, now).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		Exec(); err != nil {
		r.session.Query(deleteEmail, email).WithContext(ctx).Consistency(gocql.Quorum).Exec()
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &models.User{
		ID:           userID,
		Email:        email,
		AccountID:    accountID,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Delete removes a user and its email lookup row. It is used to roll back a
// registration when the provisioning event cannot be published, so the email is
// freed for a clean retry.
func (r *UserRepository) Delete(ctx context.Context, userID uuid.UUID, email string) error {
	if err := r.session.Query(deleteUser, userID).WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if err := r.session.Query(deleteEmail, email).WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("delete email lookup: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var userID uuid.UUID
	if err := r.session.Query(selectUserIDByEmail, email).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		Scan(&userID); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user id by email: %w", err)
	}
	return r.FindByID(ctx, userID)
}

func (r *UserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var u models.User
	var createdAt, updatedAt time.Time
	err := r.session.Query(selectByID, userID).
		WithContext(ctx).
		Consistency(gocql.Quorum).
		Scan(&u.ID, &u.Email, &u.AccountID, &u.PasswordHash, &createdAt, &updatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return &u, nil
}

func (r *UserRepository) WriteAudit(ctx context.Context, ev models.AuditEvent) error {
	if err := r.session.Query(insertAudit, ev.Day, ev.Timestamp, ev.EventID, ev.ActorUserID, ev.Action, ev.ResourceType, ev.ResourceID, ev.CorrelationID, ev.Metadata).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
