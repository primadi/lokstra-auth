package simple

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	token "github.com/primadi/lokstra-auth/token"
	"github.com/primadi/lokstra-auth/token/revocation_list"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// tokenData stores token information in memory
type tokenData struct {
	Claims    token.Claims
	ExpiresAt time.Time
}

// Manager handles simple opaque token generation and verification
// @Service "simple-token-manager"
type Manager struct {
	// @InjectCfgValue "token"
	Config         *Config
	revocationList *revocation_list.InMemoryRevocationList
	mu             sync.RWMutex
	tokens         map[string]*tokenData // token -> data
}

// Init initializes the manager after dependency injection
// @PostConstruct
func (m *Manager) Init() {
	// Set defaults if not configured
	if m.Config.TokenLength == 0 {
		m.Config.TokenLength = 32
	}
	if m.Config.TokenDuration == 0 {
		m.Config.TokenDuration = 1 * time.Hour
	}

	// Initialize map
	m.tokens = make(map[string]*tokenData)

	// Start cleanup goroutine
	go m.cleanup()
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
	tokenBytes := make([]byte, m.Config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	tokenValue := base64.URLEncoding.EncodeToString(tokenBytes)
	now := time.Now()
	expiresAt := now.Add(m.Config.TokenDuration)

	// Store token data
	m.mu.Lock()
	m.tokens[tokenValue] = &tokenData{
		Claims:    claims,
		ExpiresAt: expiresAt,
	}
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
	data, ok := m.tokens[tokenValue]
	m.mu.RUnlock()

	if !ok {
		return &token.VerificationResult{
			Valid: false,
			Error: ErrInvalidToken,
		}, nil
	}

	// Check expiration
	if time.Now().After(data.ExpiresAt) {
		// Clean up expired token
		m.mu.Lock()
		delete(m.tokens, tokenValue)
		m.mu.Unlock()

		return &token.VerificationResult{
			Valid: false,
			Error: ErrExpiredToken,
		}, nil
	}

	// Check revocation
	if m.Config.EnableRevocation {
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
	tenantID, hasTenant := data.Claims.GetTenantID()
	if !hasTenant || tenantID == "" {
		return &token.VerificationResult{
			Valid: false,
			Error: errors.New("token missing tenant_id claim"),
		}, nil
	}

	appID, hasApp := data.Claims.GetAppID()
	if !hasApp || appID == "" {
		return &token.VerificationResult{
			Valid: false,
			Error: errors.New("token missing app_id claim"),
		}, nil
	}

	return &token.VerificationResult{
		Valid:  true,
		Claims: data.Claims,
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
	if !m.Config.EnableRevocation {
		return errors.New("revocation not enabled")
	}

	m.mu.RLock()
	data, ok := m.tokens[tokenValue]
	m.mu.RUnlock()

	if !ok {
		return ErrInvalidToken
	}

	return m.revocationList.Add(ctx, tokenValue, data.ExpiresAt)
}

// cleanup removes expired tokens periodically
func (m *Manager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for tokenValue, data := range m.tokens {
			if now.After(data.ExpiresAt) {
				delete(m.tokens, tokenValue)
			}
		}
		m.mu.Unlock()

		// Cleanup revocation list
		if m.Config.EnableRevocation {
			m.revocationList.Cleanup(context.Background())
		}
	}
}

// GenerateResetToken generates a one-time password reset token for email
func (m *Manager) GenerateResetToken(ctx context.Context, email string) (string, error) {
	// Generate a secure random token using the configured token length
	tokenBytes := make([]byte, m.Config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Encode as base64 URL-safe string
	resetToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// TODO: Store the reset token with email and expiration time in a store
	// For now, just return the token
	// In production, you should:
	// 1. Hash the token before storing
	// 2. Store with email and expiration (e.g., 1 hour)
	// 3. Implement verification logic

	return resetToken, nil
}
