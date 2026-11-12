package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024 // 64 MB
	iterations  = 3
	parallelism = 4
	keyLength   = 32
	saltLength  = 16
)

type encodedHash struct {
	version     int
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        string
	hash        string
}

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", fmt.Errorf("password too short: minimum 12 characters")
	}
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return encodeHash(hash, salt), nil
}

func VerifyPassword(password, encodedHash string) bool {
	h, err := parseHash(encodedHash)
	if err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(h.salt)
	if err != nil {
		return false
	}
	hash, err := base64.RawStdEncoding.DecodeString(h.hash)
	if err != nil {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, uint32(len(hash)))
	if len(computed) != len(hash) {
		return false
	}
	return subtleConstantTimeCompare(computed, hash) == 0
}

func encodeHash(hash, salt []byte) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func parseHash(encoded string) (*encodedHash, error) {
	if !strings.HasPrefix(encoded, "$argon2id$") {
		return nil, fmt.Errorf("unsupported hash format")
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil, fmt.Errorf("malformed hash")
	}
	var version, memory, iterations int
	var parallelism uint
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, err
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return nil, err
	}
	return &encodedHash{
		version:     version,
		memory:      uint32(memory),
		iterations:  uint32(iterations),
		parallelism: uint8(parallelism),
		salt:        parts[4],
		hash:        parts[5],
	}, nil
}

func subtleConstantTimeCompare(a, b []byte) int {
	// constant-time comparison
	result := 0
	if len(a) != len(b) {
		result = 1
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		result |= int(a[i] ^ b[i])
	}
	return result
}
