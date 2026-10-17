package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/security"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
)

type fakeUserStore struct {
	user        *models.User
	createErr   error
	findErr     error
	createCalls int
}

func (f *fakeUserStore) Create(_ context.Context, email string, accountID uuid.UUID, passwordHash string) (*models.User, error) {
	f.createCalls++
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.user = &models.User{ID: uuid.New(), Email: email, AccountID: accountID, PasswordHash: passwordHash}
	return f.user, nil
}

func (f *fakeUserStore) FindByEmail(_ context.Context, _ string) (*models.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

func (f *fakeUserStore) FindByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

type fakeRefreshTokenStore struct {
	insertErr error
	getToken  *models.RefreshToken
	getErr    error
	revokeErr error
	inserted  []string
	revoked   []string
}

func (f *fakeRefreshTokenStore) Insert(_ context.Context, tokenHash string, _ uuid.UUID, _ uuid.UUID, _ string, _ time.Time) error {
	f.inserted = append(f.inserted, tokenHash)
	return f.insertErr
}

func (f *fakeRefreshTokenStore) GetByHash(_ context.Context, _ string) (*models.RefreshToken, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getToken, nil
}

func (f *fakeRefreshTokenStore) Revoke(_ context.Context, tokenHash string) error {
	f.revoked = append(f.revoked, tokenHash)
	return f.revokeErr
}

type fakeSigner struct {
	token string
	err   error
}

func (f fakeSigner) SignAccessToken(_, _, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}

func newTestUser(t *testing.T, password string) *models.User {
	t.Helper()
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	return &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		AccountID:    uuid.New(),
		PasswordHash: hash,
	}
}

const validPassword = "Str0ng!Passw0rd"

func TestAccessTokenExpiresIn(t *testing.T) {
	if got := AccessTokenExpiresIn(); got != 900 {
		t.Fatalf("AccessTokenExpiresIn() = %d, want 900", got)
	}
}

func TestRegisterSuccess(t *testing.T) {
	users := &fakeUserStore{}
	tokens := &fakeRefreshTokenStore{}
	svc := NewAuthService(users, tokens, fakeSigner{token: "access-token"})

	res, err := svc.Register(context.Background(), "user@example.com", validPassword)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.UserID == "" || res.AccountID == "" {
		t.Fatalf("missing ids: %+v", res)
	}
	if res.AccessToken != "access-token" || res.RefreshToken == "" {
		t.Fatalf("unexpected tokens: %+v", res)
	}
	if len(tokens.inserted) != 1 {
		t.Fatalf("expected one refresh token insert, got %d", len(tokens.inserted))
	}
}

func TestRegisterValidationErrors(t *testing.T) {
	svc := NewAuthService(&fakeUserStore{}, &fakeRefreshTokenStore{}, fakeSigner{token: "t"})

	if _, err := svc.Register(context.Background(), "", validPassword); err == nil {
		t.Fatal("expected error for empty email")
	}
	if _, err := svc.Register(context.Background(), "user@example.com", "short"); err == nil {
		t.Fatal("expected error for weak password")
	}
}

func TestRegisterStoreErrors(t *testing.T) {
	t.Run("app error passthrough", func(t *testing.T) {
		users := &fakeUserStore{createErr: applierrors.ErrEmailAlreadyRegistered}
		svc := NewAuthService(users, &fakeRefreshTokenStore{}, fakeSigner{token: "t"})
		_, err := svc.Register(context.Background(), "user@example.com", validPassword)
		if !errors.Is(err, applierrors.ErrEmailAlreadyRegistered) {
			t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
		}
	})

	t.Run("unknown error becomes internal", func(t *testing.T) {
		users := &fakeUserStore{createErr: errors.New("boom")}
		svc := NewAuthService(users, &fakeRefreshTokenStore{}, fakeSigner{token: "t"})
		_, err := svc.Register(context.Background(), "user@example.com", validPassword)
		if !errors.Is(err, applierrors.ErrInternal) {
			t.Fatalf("expected ErrInternal, got %v", err)
		}
	})
}

func TestRegisterSignerFailure(t *testing.T) {
	svc := NewAuthService(&fakeUserStore{}, &fakeRefreshTokenStore{}, fakeSigner{err: errors.New("sign")})
	if _, err := svc.Register(context.Background(), "user@example.com", validPassword); err == nil {
		t.Fatal("expected issue tokens error")
	}
}

func TestLoginSuccess(t *testing.T) {
	users := &fakeUserStore{user: newTestUser(t, validPassword)}
	tokens := &fakeRefreshTokenStore{}
	svc := NewAuthService(users, tokens, fakeSigner{token: "access-token"})

	res, err := svc.Login(context.Background(), "user@example.com", validPassword)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.AccessToken != "access-token" || res.RefreshToken == "" {
		t.Fatalf("unexpected tokens: %+v", res)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	users := &fakeUserStore{findErr: errors.New("not found")}
	svc := NewAuthService(users, &fakeRefreshTokenStore{}, fakeSigner{token: "t"})
	if _, err := svc.Login(context.Background(), "user@example.com", validPassword); !errors.Is(err, applierrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown user, got %v", err)
	}

	users = &fakeUserStore{user: newTestUser(t, validPassword)}
	svc = NewAuthService(users, &fakeRefreshTokenStore{}, fakeSigner{token: "t"})
	if _, err := svc.Login(context.Background(), "user@example.com", "Wrong!Passw0rd"); !errors.Is(err, applierrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}
}

func TestLoginSignerFailure(t *testing.T) {
	users := &fakeUserStore{user: newTestUser(t, validPassword)}
	svc := NewAuthService(users, &fakeRefreshTokenStore{}, fakeSigner{err: errors.New("sign")})
	if _, err := svc.Login(context.Background(), "user@example.com", validPassword); err == nil {
		t.Fatal("expected issue tokens error")
	}
}

func TestRefreshSuccess(t *testing.T) {
	rt := &models.RefreshToken{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		Email:     "user@example.com",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	tokens := &fakeRefreshTokenStore{getToken: rt}
	svc := NewAuthService(&fakeUserStore{}, tokens, fakeSigner{token: "new-access"})

	res, err := svc.Refresh(context.Background(), "raw-refresh")
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if res.AccessToken != "new-access" || res.RefreshToken == "" {
		t.Fatalf("unexpected tokens: %+v", res)
	}
	if len(tokens.revoked) != 1 || len(tokens.inserted) != 1 {
		t.Fatalf("expected rotate (revoke+insert), revoked=%d inserted=%d", len(tokens.revoked), len(tokens.inserted))
	}
}

func TestRefreshRejectsInvalidTokens(t *testing.T) {
	revoked := time.Now().UTC()
	cases := map[string]*fakeRefreshTokenStore{
		"not found": {getErr: errors.New("missing")},
		"revoked":   {getToken: &models.RefreshToken{RevokedAt: &revoked, ExpiresAt: time.Now().UTC().Add(time.Hour)}},
		"expired":   {getToken: &models.RefreshToken{ExpiresAt: time.Now().UTC().Add(-time.Hour)}},
	}
	for name, tokens := range cases {
		t.Run(name, func(t *testing.T) {
			svc := NewAuthService(&fakeUserStore{}, tokens, fakeSigner{token: "t"})
			if _, err := svc.Refresh(context.Background(), "raw"); !errors.Is(err, applierrors.ErrInvalidRefreshToken) {
				t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
			}
		})
	}
}

func TestRefreshRevokeAndIssueErrors(t *testing.T) {
	rt := &models.RefreshToken{ExpiresAt: time.Now().UTC().Add(time.Hour)}

	tokens := &fakeRefreshTokenStore{getToken: rt, revokeErr: errors.New("revoke")}
	svc := NewAuthService(&fakeUserStore{}, tokens, fakeSigner{token: "t"})
	if _, err := svc.Refresh(context.Background(), "raw"); !errors.Is(err, applierrors.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken on revoke failure, got %v", err)
	}

	tokens = &fakeRefreshTokenStore{getToken: rt}
	svc = NewAuthService(&fakeUserStore{}, tokens, fakeSigner{err: errors.New("sign")})
	if _, err := svc.Refresh(context.Background(), "raw"); !errors.Is(err, applierrors.ErrInternal) {
		t.Fatalf("expected ErrInternal on issue failure, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	tokens := &fakeRefreshTokenStore{}
	svc := NewAuthService(&fakeUserStore{}, tokens, fakeSigner{token: "t"})
	if err := svc.Logout(context.Background(), "raw"); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if len(tokens.revoked) != 1 {
		t.Fatalf("expected one revoke, got %d", len(tokens.revoked))
	}

	tokens = &fakeRefreshTokenStore{revokeErr: errors.New("revoke")}
	svc = NewAuthService(&fakeUserStore{}, tokens, fakeSigner{token: "t"})
	if err := svc.Logout(context.Background(), "raw"); !errors.Is(err, applierrors.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}
