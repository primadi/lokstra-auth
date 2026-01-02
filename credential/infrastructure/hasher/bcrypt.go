package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// Password Hashing (bcrypt)
// =============================================================================

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a hashed password with a plaintext password
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// =============================================================================
// API Key Secret Generation & Hashing (SHA3-256)
// =============================================================================

// GenerateSecureSecret generates a cryptographically secure random secret
func GenerateSecureSecret(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HashSecretSHA3 hashes a secret using SHA3-256
func HashSecretSHA3(secret string) (string, error) {
	hash := sha3.New256()
	hash.Write([]byte(secret))
	hashBytes := hash.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hashBytes), nil
}

// VerifySecretHash verifies a secret against its hash
func VerifySecretHash(secret, hash string) (bool, error) {
	computedHash, err := HashSecretSHA3(secret)
	if err != nil {
		return false, err
	}

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hash)) == 1, nil
}

// =============================================================================
// Utility Functions
// =============================================================================

// SecureCompare performs a constant-time comparison of two strings
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
