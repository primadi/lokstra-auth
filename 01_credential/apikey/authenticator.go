package apikey

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	credential "github.com/primadi/lokstra-auth/01_credential"
	"golang.org/x/crypto/sha3"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid API key")
	ErrAPIKeyExpired    = errors.New("API key expired")
	ErrAPIKeyRevoked    = errors.New("API key revoked")
	ErrAPIKeyNotFound   = errors.New("API key not found")
	ErrInvalidKeyFormat = errors.New("invalid API key format")
)

// Credentials represents API key credentials
type Credentials struct {
	APIKey string
	Prefix string // Optional: for key identification (e.g., "sk_live_", "pk_test_")
}

func (c *Credentials) Type() string {
	return "apikey"
}

func (c *Credentials) Validate() error {
	if c.APIKey == "" {
		return errors.New("API key is required")
	}
	return nil
}

// APIKey represents a stored API key
type APIKey struct {
	ID        string
	KeyHash   string                 // Hashed version of the key
	Prefix    string                 // Key prefix for identification
	UserID    string                 // Owner of the key
	Name      string                 // Descriptive name
	Scopes    []string               // Allowed permissions/scopes
	Metadata  map[string]interface{} // Additional metadata
	CreatedAt time.Time
	ExpiresAt *time.Time // nil = never expires
	LastUsed  *time.Time
	Revoked   bool
	RevokedAt *time.Time
}

// KeyStore manages API keys
type KeyStore interface {
	// GetByHash retrieves an API key by its hash
	GetByHash(ctx context.Context, hash string) (*APIKey, error)

	// GetByPrefix retrieves all API keys with the given prefix
	GetByPrefix(ctx context.Context, prefix string) ([]*APIKey, error)

	// Store saves an API key
	Store(ctx context.Context, key *APIKey) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, keyID string, timestamp time.Time) error

	// Revoke marks an API key as revoked
	Revoke(ctx context.Context, keyID string) error

	// Delete removes an API key
	Delete(ctx context.Context, keyID string) error
}

// Authenticator handles API key authentication
type Authenticator struct {
	keyStore KeyStore
	hasher   *KeyHasher
}

// Config holds configuration for API key authenticator
type Config struct {
	KeyStore KeyStore
}

// NewAuthenticator creates a new API key authenticator
func NewAuthenticator(config *Config) *Authenticator {
	if config == nil {
		config = &Config{}
	}

	// Use in-memory store if not provided
	if config.KeyStore == nil {
		config.KeyStore = NewInMemoryKeyStore()
	}

	return &Authenticator{
		keyStore: config.KeyStore,
		hasher:   NewKeyHasher(),
	}
}

// Authenticate verifies API key credentials
func (a *Authenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	apiKeyCreds, ok := creds.(*Credentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   errors.New("invalid credentials type"),
		}, nil
	}

	if err := apiKeyCreds.Validate(); err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Hash the provided key
	keyHash := a.hasher.Hash(apiKeyCreds.APIKey)

	// Get API key from store
	apiKey, err := a.keyStore.GetByHash(ctx, keyHash)
	if err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrAPIKeyNotFound,
		}, nil
	}

	// Check if key is revoked
	if apiKey.Revoked {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrAPIKeyRevoked,
		}, nil
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrAPIKeyExpired,
		}, nil
	}

	// Update last used timestamp (async, don't wait)
	go func() {
		_ = a.keyStore.UpdateLastUsed(context.Background(), apiKey.ID, time.Now())
	}()

	// Build claims
	claims := map[string]interface{}{
		"sub":       apiKey.UserID,
		"key_id":    apiKey.ID,
		"key_name":  apiKey.Name,
		"scopes":    apiKey.Scopes,
		"auth_type": "apikey",
	}

	// Add metadata to claims
	for key, value := range apiKey.Metadata {
		if _, exists := claims[key]; !exists {
			claims[key] = value
		}
	}

	return &credential.AuthenticationResult{
		Success: true,
		Subject: apiKey.UserID,
		Claims:  claims,
	}, nil
}

// Type returns the authenticator type
func (a *Authenticator) Type() string {
	return "apikey"
}

// GenerateKey generates a new API key
func (a *Authenticator) GenerateKey(ctx context.Context, userID, name string, scopes []string, expiresIn *time.Duration) (keyString string, apiKey *APIKey, err error) {
	// Generate random key
	keyString, err = a.hasher.Generate()
	if err != nil {
		return "", nil, err
	}

	// Extract prefix (first 8 characters)
	prefix := ""
	if len(keyString) >= 8 {
		prefix = keyString[:8]
	}

	// Hash the key
	keyHash := a.hasher.Hash(keyString)

	// Create API key record
	apiKey = &APIKey{
		ID:        generateID(),
		KeyHash:   keyHash,
		Prefix:    prefix,
		UserID:    userID,
		Name:      name,
		Scopes:    scopes,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	// Set expiration
	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		apiKey.ExpiresAt = &expiresAt
	}

	// Store the key
	if err := a.keyStore.Store(ctx, apiKey); err != nil {
		return "", nil, err
	}

	return keyString, apiKey, nil
}

// RevokeKey revokes an API key
func (a *Authenticator) RevokeKey(ctx context.Context, keyID string) error {
	return a.keyStore.Revoke(ctx, keyID)
}

// KeyHasher handles key hashing
type KeyHasher struct{}

func NewKeyHasher() *KeyHasher {
	return &KeyHasher{}
}

// Generate generates a new random API key
func (h *KeyHasher) Generate() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode as base64
	key := base64.URLEncoding.EncodeToString(bytes)

	// Remove padding
	key = strings.TrimRight(key, "=")

	return key, nil
}

// Hash hashes an API key using SHA3-256
func (h *KeyHasher) Hash(key string) string {
	hash := sha3.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// Compare compares a plain key with a hashed key
func (h *KeyHasher) Compare(plainKey, hashedKey string) bool {
	hash := h.Hash(plainKey)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(hashedKey)) == 1
}

// InMemoryKeyStore is an in-memory implementation of KeyStore
type InMemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*APIKey // keyed by hash
}

func NewInMemoryKeyStore() *InMemoryKeyStore {
	return &InMemoryKeyStore{
		keys: make(map[string]*APIKey),
	}
}

func (s *InMemoryKeyStore) GetByHash(ctx context.Context, hash string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.keys[hash]
	if !ok {
		return nil, ErrAPIKeyNotFound
	}

	return key, nil
}

func (s *InMemoryKeyStore) GetByPrefix(ctx context.Context, prefix string) ([]*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*APIKey
	for _, key := range s.keys {
		if key.Prefix == prefix {
			results = append(results, key)
		}
	}

	return results, nil
}

func (s *InMemoryKeyStore) Store(ctx context.Context, key *APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[key.KeyHash] = key
	return nil
}

func (s *InMemoryKeyStore) UpdateLastUsed(ctx context.Context, keyID string, timestamp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range s.keys {
		if key.ID == keyID {
			key.LastUsed = &timestamp
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

func (s *InMemoryKeyStore) Revoke(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range s.keys {
		if key.ID == keyID {
			key.Revoked = true
			now := time.Now()
			key.RevokedAt = &now
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

func (s *InMemoryKeyStore) Delete(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hash, key := range s.keys {
		if key.ID == keyID {
			delete(s.keys, hash)
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

// Helper function to generate unique IDs
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
