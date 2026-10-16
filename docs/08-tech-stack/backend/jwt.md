# NEO Bank — JWT Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Purpose

This document defines how NEO Bank issues, validates, rotates, and revokes JWT access tokens and refresh tokens. It is the **how**; the **why** (RS256, short-lived access tokens, hashed refresh tokens) lives in [`§8`](../../08-tech-stack.md#8-authentication).

JWTs keep the API stateless and horizontally scalable. RS256 lets verification happen with a public key only — the private key never leaves the signing service.

## 2. Token Lifetimes

| Token type | Algorithm | Lifetime | Storage | Revocation |
|------------|-----------|----------|---------|------------|
| Access token | RS256 | 15 minutes | Client (memory only) | Short expiry; no server-side revocation |
| Refresh token | SHA-256 hash stored | 7 days | Hashed in Cassandra | Explicit revoke on logout / password change |

**Rules:**
- Access tokens expire after 15 minutes; clients must use the refresh endpoint to obtain new ones.
- Refresh tokens are hashed (SHA-256) before storage — a database leak does not yield usable tokens.
- Refresh tokens are rotated on every use (new refresh token issued, old one revoked).
- Refresh tokens are invalidated on logout or password change.

## 3. Token Claims

### Access Token Claims

```go
// internal/core/auth/claims.go
type AccessClaims struct {
    AccountID string `json:"accountId"`
    Email     string `json:"email"`
    jwt.RegisteredClaims
}
```

**Required claims:**
- `sub` — subject (user ID)
- `accountId` — account ID for ownership checks
- `email` — user email
- `iat` — issued at
- `exp` — expiration (15 minutes from issue)
- `jti` — JWT ID (unique token identifier, for potential blocklisting)

### Refresh Token

Stored as `SHA-256(token_value)` in the `refresh_tokens` table with `user_id`, `expires_at`, and `revoked_at`.

## 4. Signing & Validation

### Key Management

```go
// internal/core/auth/keys.go
type KeyManager struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    keyID      string
}

func NewKeyManager(privateKeyPEM, publicKeyPEM []byte, keyID string) (*KeyManager, error) {
    // Parse PEM blocks into *rsa.PrivateKey and *rsa.PublicKey
    // Validate key sizes (min 2048 bits)
}
```

**Rules:**
- Private key signs access tokens; public key verifies them.
- Keys rotated every 90 days.
- Old public keys retained during rotation for in-flight token verification.
- Keys loaded from environment variables or secret manager — never committed to the repository.

### Signing an Access Token

```go
func (m *KeyManager) SignAccessToken(userID, accountID, email string) (string, error) {
    now := time.Now()
    claims := AccessClaims{
        AccountID: accountID,
        Email:     email,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   userID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
            ID:        uuid.New().String(),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = m.keyID
    return token.SignedString(m.privateKey)
}
```

**Rules:**
- Always sign with RS256.
- Include `kid` (key ID) header for key rotation support.
- Set `exp` — never issue tokens without expiration.
- Use UUID v7 for `jti`.

### Validating an Access Token

```go
func (m *KeyManager) Validate(tokenString string) (*AccessClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return m.publicKey, nil
    })
    if err != nil {
        return nil, err
    }
    claims, ok := token.Claims.(*AccessClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}
```

**Rules:**
- Validate the signing method is RS256 — reject `none` and HS256.
- Check `exp`, `iat`, `nbf` via `jwt.ParseWithClaims`.
- Reject tokens with missing required claims (`sub`, `accountId`, `email`).

## 5. Refresh Token Flow

### Issue Refresh Token (on login)

```go
func (s *AuthService) IssueRefreshToken(ctx context.Context, userID string) (string, error) {
    raw := uuid.New().String()
    hash := sha256.Sum256([]byte(raw))
    
    // Store hash in Cassandra
    if err := s.refreshTokenRepo.Insert(ctx, hash, userID, time.Now().Add(7*24*time.Hour)); err != nil {
        return "", err
    }
    return raw, nil // Return raw token to client once
}
```

### Rotate Refresh Token (on refresh)

```go
func (s *AuthService) RotateRefreshToken(ctx context.Context, oldHash string) (string, string, error) {
    // 1. Look up old hash
    rt, err := s.refreshTokenRepo.GetByHash(ctx, oldHash)
    if err != nil || rt.RevokedAt != nil {
        return "", "", errors.New("invalid or revoked refresh token")
    }
    if time.Now().After(rt.ExpiresAt) {
        return "", "", errors.New("refresh token expired")
    }
    
    // 2. Revoke old hash
    if err := s.refreshTokenRepo.Revoke(ctx, oldHash); err != nil {
        return "", "", err
    }
    
    // 3. Issue new refresh token
    newRaw, err := s.IssueRefreshToken(ctx, rt.UserID)
    if err != nil {
        return "", "", err
    }
    
    // 4. Issue new access token
    accessToken, err := s.keyManager.SignAccessToken(rt.UserID, rt.AccountID, rt.Email)
    if err != nil {
        return "", "", err
    }
    
    return accessToken, newRaw, nil
}
```

**Rules:**
- Refresh tokens are rotated on every use — old hash revoked, new hash issued.
- Revoked refresh tokens cannot be reused.
- Expired refresh tokens require re-login.

### Logout (Revoke Refresh Token)

```go
func (s *AuthService) RevokeRefreshToken(ctx context.Context, rawToken string) error {
    hash := sha256.Sum256([]byte(rawToken))
    return s.refreshTokenRepo.Revoke(ctx, hash[:])
}
```

**Rules:**
- Logout revokes the refresh token hash.
- Password change revokes all refresh tokens for the user (force re-login on all devices).

## 6. Gin Middleware

```go
// internal/api/middleware/auth.go
func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
    tokenStr := extractBearerToken(c.Request)
    if tokenStr == "" {
        c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHENTICATED", "Authentication required", "", getRequestID(c)))
        c.Abort()
        return
    }

    claims, err := m.keyManager.Validate(tokenStr)
    if err != nil {
        c.JSON(http.StatusUnauthorized, errorResponse("TOKEN_EXPIRED", "Access token expired or invalid", "", getRequestID(c)))
        c.Abort()
        return
    }

    c.Set("userId", claims.Subject)
    c.Set("accountId", claims.AccountID)
    c.Set("email", claims.Email)
    c.Next()
}
```

**Rules:**
- Extract bearer token from `Authorization: Bearer <token>` header.
- Validate with `jwt.ParseWithClaims` (RS256 only).
- Store claims in Gin context via `c.Set` for handlers to read.
- `c.Abort()` after writing error — stops the handler chain.

## 7. Password Hashing (Argon2id)

```go
// internal/core/security/password.go
func HashPassword(password string) (string, error) {
    hash, err := argon2.IDKey(
        []byte(password),
        []byte(salt),           // random 16-byte salt
        3,                      // iterations
        64*1024,                // memory 64 MB
        4,                      // parallelism
        32,                     // key length
    )
    return encodeHash(hash, salt), nil
}

func VerifyPassword(password, encodedHash string) bool {
    // Decode salt and hash from encoded string
    // Recompute argon2.IDKey with same parameters
    // Compare with subtle.ConstantTimeCompare
}
```

**Rules:**
- Argon2id: memory=64MB, iterations=3, parallelism=4, key length=32 bytes.
- Random 16-byte salt per password.
- Use `subtle.ConstantTimeCompare` for comparison — never `==`.
- Password minimum: 12 characters, uppercase + lowercase + number + symbol.
- Passwords are hashed before storage; never logged or returned in responses.

## 8. Key Rotation

Every 90 days:

1. Generate a new RSA key pair.
2. Sign new tokens with the new private key; set `kid` to the new key ID.
3. Keep the old public key in the validation key map (indexed by `kid`) so in-flight tokens still verify.
4. After all old tokens have expired (15 minutes), retire the old public key.
5. Store the new private key in secret manager; update application configuration.

**Rules:**
- Rotation is transparent to clients — no re-login required.
- Old public keys retained until all tokens signed with the old private key have expired.
- Rotation events logged and audited.

## 9. Security Rules

| Rule | Reason |
|------|--------|
| RS256 only — reject `none` and HS256 | Prevents algorithm confusion attacks |
| Access tokens expire in 15 minutes | Limits window of opportunity if stolen |
| Refresh tokens hashed in DB | Database leak does not yield usable tokens |
| Refresh token rotation on every use | Detects token reuse / theft |
| Revoke on logout and password change | Force re-login on all devices |
| `kid` header for key rotation | Zero-downtime key rotation |
| No sensitive data in token payload | JWTs are base64-encoded, not encrypted |
| Access tokens carry no PII beyond email | Minimise exposure if token is logged |

## 10. Code Style

- **`jwt.ParseWithClaims`** — never `jwt.Parse` without claims validation.
- **Signing method check** — reject anything that is not `*jwt.SigningMethodRSA`.
- **`subtle.ConstantTimeCompare`** — never `==` for token/password comparison.
- **Keys from environment or secret manager** — never hardcoded or committed.
- **`c.Set` / `c.Get`** for Gin context values (user_id, account_id).
- **`c.Abort()`** after writing an authentication error response.

---
↑ Part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md).