package services

import (
	"context"
	"testing"
	"time"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/repositories"
)

func contactPoints() []string {
	pts := []string{"localhost:9042"}
	return pts
}

func connectCassandra(t *testing.T) *gocql.Session {
	t.Helper()
	cluster := gocql.NewCluster(contactPoints()...)
	cluster.Keyspace = "auth_keyspace"
	cluster.Timeout = 3 * time.Second
	cluster.ConnectTimeout = 3 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		t.Skipf("Cassandra unavailable, skipping integration test: %v", err)
	}
	return session
}

func ensureSchema(t *testing.T, session *gocql.Session) {
	t.Helper()
	stmts := []string{
		`CREATE KEYSPACE IF NOT EXISTS auth_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		`USE auth_keyspace`,
		`CREATE TABLE IF NOT EXISTS users (user_id UUID, email TEXT, account_id UUID, password_hash TEXT, created_at TIMESTAMP, updated_at TIMESTAMP, PRIMARY KEY (user_id))`,
		`CREATE TABLE IF NOT EXISTS users_by_email (email TEXT, user_id UUID, account_id UUID, PRIMARY KEY (email))`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (token_hash TEXT, user_id UUID, account_id UUID, email TEXT, expires_at TIMESTAMP, revoked_at TIMESTAMP, created_at TIMESTAMP, PRIMARY KEY (token_hash))`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens_by_user (user_id UUID, token_hash TEXT, expires_at TIMESTAMP, revoked_at TIMESTAMP, created_at TIMESTAMP, PRIMARY KEY (user_id, token_hash))`,
		`CREATE TABLE IF NOT EXISTS audit_events (service TEXT, day DATE, timestamp TIMESTAMP, event_id UUID, event_type TEXT, actor_id UUID, resource_id UUID, ip_address TEXT, user_agent TEXT, details TEXT, PRIMARY KEY ((service, day), timestamp, event_id))`,
	}
	for _, stmt := range stmts {
		if err := session.Query(stmt).Exec(); err != nil {
			t.Fatalf("setup schema: %v", err)
		}
	}
}

func truncate(t *testing.T, session *gocql.Session) {
	t.Helper()
	for _, table := range []string{"users", "users_by_email", "refresh_tokens", "refresh_tokens_by_user"} {
		session.Query(`TRUNCATE ` + table).Exec()
	}
}

func TestRegisterLoginRefreshLogoutIntegration(t *testing.T) {
	session := connectCassandra(t)
	defer session.Close()
	ensureSchema(t, session)
	defer truncate(t, session)

	userRepo := repositories.NewUserRepository(session)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(session)
	keyManager, err := auth.GenerateDevKeyPair("integration-key")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}
	svc := NewAuthService(userRepo, refreshTokenRepo, keyManager)

	ctx := context.Background()
	email := "integration-" + t.Name() + "@example.com"
	password := "Str0ng!Passw0rd"

	// Register
	reg, err := svc.Register(ctx, email, password)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if reg.UserID == "" || reg.AccountID == "" || reg.AccessToken == "" || reg.RefreshToken == "" {
		t.Fatal("Register returned empty fields")
	}

	// Duplicate email rejected
	if _, err := svc.Register(ctx, email, password); err == nil {
		t.Fatal("expected duplicate registration to fail")
	}

	// Login
	login, err := svc.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if login.AccountID != reg.AccountID {
		t.Fatalf("login accountId %s != register accountId %s", login.AccountID, reg.AccountID)
	}

	// Wrong password rejected
	if _, err := svc.Login(ctx, email, "Wrong!Passw0rd1"); err == nil {
		t.Fatal("expected wrong password to fail")
	}

	// Refresh
	refreshed, err := svc.Refresh(ctx, login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("Refresh returned empty tokens")
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("Refresh did not rotate the refresh token")
	}

	// Old refresh token no longer valid
	if _, err := svc.Refresh(ctx, login.RefreshToken); err == nil {
		t.Fatal("expected rotated refresh token to be rejected")
	}

	// Logout invalidates the new refresh token
	if err := svc.Logout(ctx, refreshed.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := svc.Refresh(ctx, refreshed.RefreshToken); err == nil {
		t.Fatal("expected revoked refresh token to be rejected")
	}
}
