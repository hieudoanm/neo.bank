package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

type KeyManager struct {
	publicKey *rsa.PublicKey
	keyID     string
	mu        sync.RWMutex
}

func NewKeyManager(publicKeyPEM []byte, keyID string) (*KeyManager, error) {
	var publicKey *rsa.PublicKey
	if len(publicKeyPEM) > 0 {
		pk, err := parsePublicKeyPEM(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		publicKey = pk
	}
	if keyID == "" {
		keyID = uuid.New().String()
	}
	return &KeyManager{publicKey: publicKey, keyID: keyID}, nil
}

func parsePublicKeyPEM(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pkcs1Key, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse RSA public key: %w", err)
		}
		return pkcs1Key, nil
	}
	rsaKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsed key is not an RSA public key")
	}
	return rsaKey, nil
}

func (m *KeyManager) Validate(tokenString string) (*AccessClaims, error) {
	m.mu.RLock()
	publicKey := m.publicKey
	m.mu.RUnlock()
	if publicKey == nil {
		return nil, fmt.Errorf("no public key configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	}, jwt.WithIssuedAt())
	if err != nil {
		return nil, err
	}
	// Reject tokens that declare a different key ID; the signature check is the
	// real boundary, but this stops a stale kid from being accepted silently
	// during key rotation.
	if kid, _ := token.Header["kid"].(string); kid != "" && m.keyID != "" && kid != m.keyID {
		return nil, fmt.Errorf("unexpected key id")
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	if claims.Subject == "" || claims.AccountID == "" || claims.Email == "" {
		return nil, fmt.Errorf("token missing required claims")
	}
	return claims, nil
}
