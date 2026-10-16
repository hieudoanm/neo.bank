package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
)

type UserRepository struct {
	session *gocql.Session
}

func NewUserRepository(session *gocql.Session) *UserRepository {
	return &UserRepository{session: session}
}

const (
	insertEmailLWT = `INSERT INTO users_by_email (email, user_id, account_id) VALUES (?, ?, ?) IF NOT EXISTS`
	insertUser     = `INSERT INTO users (user_id, email, account_id, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	selectByEmail  = `SELECT user_id, email, account_id, password_hash, created_at, updated_at FROM users_by_email WHERE email = ?`
	selectByID     = `SELECT user_id, email, account_id, password_hash, created_at, updated_at FROM users WHERE user_id = ?`
	deleteEmail    = `DELETE FROM users_by_email WHERE email = ?`
)

func (r *UserRepository) Create(ctx context.Context, email string, accountID uuid.UUID, passwordHash string) (*models.User, error) {
	now := time.Now().UTC()
	userID := uuid.New()

	applied, err := r.session.Query(insertEmailLWT, email, userID, accountID).
		WithContext(ctx).
		Consistency(gocql.One).
		SerialConsistency(gocql.Serial).
		ScanCAS()
	if err != nil {
		return nil, fmt.Errorf("check email uniqueness: %w", err)
	}
	if !applied {
		return nil, errors.ErrEmailAlreadyRegistered
	}

	if err := r.session.Query(insertUser, userID, email, accountID, passwordHash, now, now).
		WithContext(ctx).
		Consistency(gocql.One).
		Exec(); err != nil {
		r.session.Query(deleteEmail, email).WithContext(ctx).Exec()
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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	var createdAt, updatedAt time.Time
	err := r.session.Query(selectByEmail, email).
		WithContext(ctx).
		Consistency(gocql.One).
		Scan(&u.ID, &u.Email, &u.AccountID, &u.PasswordHash, &createdAt, &updatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var u models.User
	var createdAt, updatedAt time.Time
	err := r.session.Query(selectByID, userID).
		WithContext(ctx).
		Consistency(gocql.One).
		Scan(&u.ID, &u.Email, &u.AccountID, &u.PasswordHash, &createdAt, &updatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return &u, nil
}
