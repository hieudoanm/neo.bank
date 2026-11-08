package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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

type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	mu         sync.RWMutex
}

func NewKeyManager(privateKeyPEM, publicKeyPEM []byte, keyID string) (*KeyManager, error) {
	privateKey, err := parsePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	publicKey, err := parsePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	if keyID == "" {
		keyID = uuid.New().String()
	}
	if privateKey.Size() < 256 {
		return nil, fmt.Errorf("RSA key size must be at least 2048 bits")
	}
	return &KeyManager{privateKey: privateKey, publicKey: publicKey, keyID: keyID}, nil
}

func GenerateDevKeyPair(keyID string) (*KeyManager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}
	if keyID == "" {
		keyID = uuid.New().String()
	}
	return &KeyManager{privateKey: privateKey, publicKey: &privateKey.PublicKey, keyID: keyID}, nil
}

func parsePrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse RSA private key: pkcs1=%v pkcs8=%v", err, err2)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not an RSA private key")
		}
		key = rsaKey
	}
	return key, nil
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

func (m *KeyManager) PublicKey() *rsa.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicKey
}

func (m *KeyManager) SignAccessToken(userID, accountID, email string) (string, error) {
	m.mu.RLock()
	privateKey := m.privateKey
	keyID := m.keyID
	m.mu.RUnlock()

	now := time.Now()
	claims := AccessClaims{
		AccountID: accountID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID
	return token.SignedString(privateKey)
}

func (m *KeyManager) Validate(tokenString string) (*AccessClaims, error) {
	m.mu.RLock()
	publicKey := m.publicKey
	m.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodRS256 {
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

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
