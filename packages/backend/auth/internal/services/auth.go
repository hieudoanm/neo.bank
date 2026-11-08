package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/auth"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/security"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
)

const (
	accessTokenExpiresIn = 900 // 15 minutes in seconds
	refreshTokenTTL      = 7 * 24 * time.Hour
)

// AccessTokenExpiresIn returns the access token lifetime in seconds.
func AccessTokenExpiresIn() int {
	return accessTokenExpiresIn
}

// UserStore is the persistence dependency for user records.
type UserStore interface {
	Create(ctx context.Context, email string, accountID uuid.UUID, passwordHash string) (*models.User, error)
	Delete(ctx context.Context, userID uuid.UUID, email string) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

// UserEventPublisher publishes auth lifecycle events so downstream services can
// react (for example, the Account Service provisions the account).
type UserEventPublisher interface {
	PublishUserEvent(ctx context.Context, ev models.UserCreatedEvent) error
}

// RefreshTokenStore is the persistence dependency for refresh tokens.
type RefreshTokenStore interface {
	Insert(ctx context.Context, tokenHash string, userID uuid.UUID, accountID uuid.UUID, email string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// TokenSigner issues access tokens.
type TokenSigner interface {
	SignAccessToken(userID, accountID, email string) (string, error)
}

type AuthService struct {
	userRepo         UserStore
	refreshTokenRepo RefreshTokenStore
	keyManager       TokenSigner
	publisher        UserEventPublisher
}

func NewAuthService(userRepo UserStore, refreshTokenRepo RefreshTokenStore, keyManager TokenSigner, publisher UserEventPublisher) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		keyManager:       keyManager,
		publisher:        publisher,
	}
}

type RegisterResult struct {
	UserID       string
	AccountID    string
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*RegisterResult, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	accountID := uuid.New()
	user, err := s.userRepo.Create(ctx, email, accountID, passwordHash)
	if err != nil {
		var appErr *applierrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, applierrors.ErrInternal
	}

	ev := models.UserCreatedEvent{
		EventID:      uuid.New().String(),
		EventType:    models.EventTypeUserCreated,
		EventVersion: models.EventVersionV1,
		UserID:       user.ID.String(),
		AccountID:    user.AccountID.String(),
		OccurredAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.publisher.PublishUserEvent(ctx, ev); err != nil {
		// The account cannot be provisioned without the event, so roll the user
		// back to keep registration and account creation consistent.
		if delErr := s.userRepo.Delete(ctx, user.ID, user.Email); delErr != nil {
			return nil, applierrors.ErrInternal
		}
		return nil, applierrors.ErrServiceUnavailable
	}

	accessToken, refreshToken, err := s.issueTokens(ctx, user.ID, user.AccountID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("issue tokens: %w", err)
	}

	return &RegisterResult{
		UserID:       user.ID.String(),
		AccountID:    user.AccountID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// dummyPasswordHash is a valid Argon2id hash verified against when the email
// is unknown, so login responses take the same time whether or not the account
// exists (mitigates user enumeration via timing).
var dummyPasswordHash = func() string {
	h, err := security.HashPassword("dummy-password-for-timing")
	if err != nil {
		panic(err)
	}
	return h
}()

func (s *AuthService) Login(ctx context.Context, email, password string) (*RegisterResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		security.VerifyPassword(password, dummyPasswordHash)
		// Generic failure to avoid account enumeration.
		return nil, applierrors.ErrInvalidCredentials
	}
	if !security.VerifyPassword(password, user.PasswordHash) {
		return nil, applierrors.ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.issueTokens(ctx, user.ID, user.AccountID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("issue tokens: %w", err)
	}

	return &RegisterResult{
		UserID:       user.ID.String(),
		AccountID:    user.AccountID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	tokenHash := auth.HashRefreshToken(refreshToken)
	rt, err := s.refreshTokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, applierrors.ErrInvalidRefreshToken
	}
	if rt.RevokedAt != nil {
		// A rotated token was presented again: possible token theft. Revoke the
		// user's whole token family before rejecting.
		if err := s.refreshTokenRepo.RevokeAllForUser(ctx, rt.UserID); err != nil {
			return nil, applierrors.ErrInternal
		}
		return nil, applierrors.ErrInvalidRefreshToken
	}
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, applierrors.ErrInvalidRefreshToken
	}

	// Rotate: revoke the used refresh token, issue a new pair.
	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash); err != nil {
		return nil, applierrors.ErrInvalidRefreshToken
	}

	accessToken, newRefreshToken, err := s.issueTokens(ctx, rt.UserID, rt.AccountID, rt.Email)
	if err != nil {
		return nil, applierrors.ErrInternal
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := auth.HashRefreshToken(refreshToken)
	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash); err != nil {
		// Already revoked or unknown tokens are treated as logged out so the
		// endpoint stays idempotent.
		if errors.Is(err, applierrors.ErrInvalidRefreshToken) {
			return nil
		}
		return err
	}
	return nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, email string) (string, string, error) {
	accessToken, err := s.keyManager.SignAccessToken(userID.String(), accountID.String(), email)
	if err != nil {
		return "", "", err
	}

	refreshRaw := uuid.New().String()
	refreshTokenHash := auth.HashRefreshToken(refreshRaw)
	expiresAt := time.Now().UTC().Add(refreshTokenTTL)
	if err := s.refreshTokenRepo.Insert(ctx, refreshTokenHash, userID, accountID, email, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshRaw, nil
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func validateEmail(email string) error {
	if email == "" {
		return applierrors.ValidationError(map[string]string{"email": "email is required"})
	}
	if len(email) > 254 || !emailPattern.MatchString(email) {
		return applierrors.ValidationError(map[string]string{"email": "enter a valid email address"})
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 12 {
		return applierrors.ValidationError(map[string]string{"password": "password must be at least 12 characters"})
	}
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if !(hasUpper && hasLower && hasDigit && hasSymbol) {
		return applierrors.ValidationError(map[string]string{"password": "password must contain uppercase, lowercase, number, and symbol"})
	}
	return nil
}
