package simple

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	token "github.com/primadi/lokstra-auth/token"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// Config holds simple token manager configuration
type Config struct {
	// TokenLength is the length of generated tokens in bytes (default: 32)
	TokenLength int

	// TokenDuration is how long tokens are valid
	TokenDuration time.Duration

	// Store is the token store (optional, uses in-memory if not provided)
	Store token.TokenStore

	// EnableRevocation enables token revocation support
	EnableRevocation bool
}

// DefaultConfig returns a default simple token configuration
func DefaultConfig() *Config {
	return &Config{
		TokenLength:      32,
		TokenDuration:    1 * time.Hour,
		EnableRevocation: false,
	}
}

// Manager handles simple opaque token generation and verification
type Manager struct {
	config         *Config
	store          token.TokenStore
	revocationList *InMemoryRevocationList
	mu             sync.RWMutex
	tokenToClaims  map[string]token.Claims // token -> claims
	tokenToExpiry  map[string]time.Time    // token -> expiry
}

// NewManager creates a new simple token manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	if config.TokenLength == 0 {
		config.TokenLength = 32
	}

	if config.TokenDuration == 0 {
		config.TokenDuration = 1 * time.Hour
	}

	m := &Manager{
		config:        config,
		store:         config.Store,
		tokenToClaims: make(map[string]token.Claims),
		tokenToExpiry: make(map[string]time.Time),
	}

	if config.EnableRevocation {
		m.revocationList = NewInMemoryRevocationList()
	}

	// Start cleanup goroutine
	go m.cleanup()

	return m
}

// Generate creates a new opaque token from the provided claims
func (m *Manager) Generate(ctx context.Context, claims token.Claims) (*token.Token, error) {
	// Validate required multi-tenant claims
	tenantID, hasTenant := claims.GetTenantID()
	if !hasTenant || tenantID == "" {
		return nil, errors.New("tenant_id is required in claims")
	}

	appID, hasApp := claims.GetAppID()
	if !hasApp || appID == "" {
		return nil, errors.New("app_id is required in claims")
	}

	// Generate random token
	tokenBytes := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	tokenValue := base64.URLEncoding.EncodeToString(tokenBytes)
	now := time.Now()
	expiresAt := now.Add(m.config.TokenDuration)

	// Store token and claims
	m.mu.Lock()
	m.tokenToClaims[tokenValue] = claims
	m.tokenToExpiry[tokenValue] = expiresAt
	m.mu.Unlock()

	return &token.Token{
		Value:     tokenValue,
		Type:      "Bearer",
		TenantID:  tenantID,
		AppID:     appID,
		ExpiresAt: expiresAt,
		IssuedAt:  now,
		Metadata: map[string]any{
			"token_type": "opaque",
		},
	}, nil
}

// Verify validates a token and extracts its claims
func (m *Manager) Verify(ctx context.Context, tokenValue string) (*token.VerificationResult, error) {
	m.mu.RLock()
	claims, ok := m.tokenToClaims[tokenValue]
	expiresAt, expiryOk := m.tokenToExpiry[tokenValue]
	m.mu.RUnlock()

	if !ok || !expiryOk {
		return &token.VerificationResult{
			Valid: false,
			Error: ErrInvalidToken,
		}, nil
	}

	// Check expiration
	if time.Now().After(expiresAt) {
		// Clean up expired token
		m.mu.Lock()
		delete(m.tokenToClaims, tokenValue)
		delete(m.tokenToExpiry, tokenValue)
		m.mu.Unlock()

		return &token.VerificationResult{
			Valid: false,
			Error: ErrExpiredToken,
		}, nil
	}

	// Check revocation
	if m.config.EnableRevocation {
		revoked, err := m.revocationList.IsRevoked(ctx, tokenValue)
		if err != nil {
			return &token.VerificationResult{
				Valid: false,
				Error: err,
			}, nil
		}
		if revoked {
			return &token.VerificationResult{
				Valid: false,
				Error: ErrTokenRevoked,
			}, nil
		}
	}

	// Validate multi-tenant required claims
	tenantID, hasTenant := claims.GetTenantID()
	if !hasTenant || tenantID == "" {
		return &token.VerificationResult{
			Valid: false,
			Error: errors.New("token missing tenant_id claim"),
		}, nil
	}

	appID, hasApp := claims.GetAppID()
	if !hasApp || appID == "" {
		return &token.VerificationResult{
			Valid: false,
			Error: errors.New("token missing app_id claim"),
		}, nil
	}

	return &token.VerificationResult{
		Valid:  true,
		Claims: claims,
		Metadata: map[string]any{
			"token_type": "opaque",
			"tenant_id":  tenantID,
			"app_id":     appID,
		},
	}, nil
}

// Type returns the type of tokens this manager handles
func (m *Manager) Type() string {
	return "simple"
}

// Revoke revokes a token
func (m *Manager) Revoke(ctx context.Context, tokenValue string) error {
	if !m.config.EnableRevocation {
		return errors.New("revocation not enabled")
	}

	m.mu.RLock()
	expiresAt, ok := m.tokenToExpiry[tokenValue]
	m.mu.RUnlock()

	if !ok {
		return ErrInvalidToken
	}

	return m.revocationList.Add(ctx, tokenValue, expiresAt)
}

// cleanup removes expired tokens periodically
func (m *Manager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for tokenValue, expiresAt := range m.tokenToExpiry {
			if now.After(expiresAt) {
				delete(m.tokenToClaims, tokenValue)
				delete(m.tokenToExpiry, tokenValue)
			}
		}
		m.mu.Unlock()

		// Cleanup revocation list
		if m.config.EnableRevocation {
			m.revocationList.Cleanup(context.Background())
		}
	}
}

// InMemoryRevocationList is an in-memory implementation of TokenRevocationList
type InMemoryRevocationList struct {
	mu      sync.RWMutex
	revoked map[string]time.Time // tokenID -> expiresAt
}

// NewInMemoryRevocationList creates a new in-memory revocation list
func NewInMemoryRevocationList() *InMemoryRevocationList {
	return &InMemoryRevocationList{
		revoked: make(map[string]time.Time),
	}
}

func (r *InMemoryRevocationList) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked[tokenID] = expiresAt
	return nil
}

func (r *InMemoryRevocationList) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, revoked := r.revoked[tokenID]
	return revoked, nil
}

func (r *InMemoryRevocationList) Remove(ctx context.Context, tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.revoked, tokenID)
	return nil
}

func (r *InMemoryRevocationList) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for tokenID, expiresAt := range r.revoked {
		if now.After(expiresAt) {
			delete(r.revoked, tokenID)
		}
	}
	return nil
}
