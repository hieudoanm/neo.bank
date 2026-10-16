package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const accessTokenTTL = 15 * time.Minute

type AccessClaims struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

// KeyManager verifies RS256 access tokens using a public key only. The private
// key never lives in the gateway.
type KeyManager struct {
	publicKey *rsa.PublicKey
	keyID     string
	mu        sync.RWMutex
}

func NewKeyManager(publicKeyPEM []byte, keyID string) (*KeyManager, error) {
	publicKey, err := parsePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
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
			return nil, fmt.Errorf("parse RSA public key: pkix=%v pkcs1=%v", err, err2)
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

	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
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
