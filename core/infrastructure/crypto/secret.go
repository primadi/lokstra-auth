package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/sha3"
)

// GenerateSecureSecret generates a cryptographically secure random secret
// Returns base64url encoded string of the specified byte length
func GenerateSecureSecret(byteLength int) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, byteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64url (URL-safe, no padding)
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	return encoded, nil
}

// HashSecret hashes the secret using SHA3-256
// Returns base64url encoded hash
func HashSecret(secret string) (string, error) {
	// Use SHA3-256 (more secure than SHA-256)
	hasher := sha3.New256()
	if _, err := hasher.Write([]byte(secret)); err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}

	hash := hasher.Sum(nil)

	// Encode to base64url
	encoded := base64.RawURLEncoding.EncodeToString(hash)
	return encoded, nil
}

// VerifySecret verifies a secret against its hash
func VerifySecret(secret, hash string) (bool, error) {
	// Hash the provided secret
	computedHash, err := HashSecret(secret)
	if err != nil {
		return false, err
	}

	// Constant-time comparison to prevent timing attacks
	return computedHash == hash, nil
}
