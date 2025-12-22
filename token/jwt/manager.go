package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	token "github.com/primadi/lokstra-auth/token"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrMissingClaims    = errors.New("missing required claims")
)

// Config holds JWT configuration
type Config struct {
	// SigningMethod is the signing algorithm (HS256, RS256, ES256, etc.)
	SigningMethod jwt.SigningMethod

	// SigningKey is the key used to sign tokens
	SigningKey any

	// VerifyingKey is the key used to verify tokens (can be same as SigningKey)
	VerifyingKey any

	// Issuer is the token issuer
	Issuer string

	// Audience is the intended audience
	Audience []string

	// AccessTokenDuration is how long access tokens are valid
	AccessTokenDuration time.Duration

	// RefreshTokenDuration is how long refresh tokens are valid
	RefreshTokenDuration time.Duration

	// EnableRevocation enables token revocation support
	EnableRevocation bool

	// RevocationList is the revocation list (optional)
	RevocationList token.TokenRevocationList
}

// DefaultConfig returns a default JWT configuration
func DefaultConfig(secret string) *Config {
	return &Config{
		SigningMethod:        jwt.SigningMethodHS256,
		SigningKey:           []byte(secret),
		VerifyingKey:         []byte(secret),
		Issuer:               "lokstra-auth",
		Audience:             []string{"lokstra"},
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	}
}

// Manager handles JWT token generation and verification
type Manager struct {
	config         *Config
	revocationList token.TokenRevocationList
}

// NewManager creates a new JWT manager
func NewManager(config *Config) *Manager {
	m := &Manager{
		config: config,
	}

	if config.EnableRevocation {
		if config.RevocationList != nil {
			m.revocationList = config.RevocationList
		} else {
			m.revocationList = NewInMemoryRevocationList()
		}
	}

	return m
}

// Generate creates a new JWT token from the provided claims
func (m *Manager) Generate(ctx context.Context, claims token.Claims) (*token.Token, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.AccessTokenDuration)

	// Validate required multi-tenant claims
	tenantID, hasTenant := claims.GetTenantID()
	if !hasTenant || tenantID == "" {
		return nil, errors.New("tenant_id is required in claims")
	}

	appID, hasApp := claims.GetAppID()
	if !hasApp || appID == "" {
		return nil, errors.New("app_id is required in claims")
	}

	// Generate unique JTI (JWT ID) for this token
	jti, err := generateJTI()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JTI: %w", err)
	}

	// Build JWT claims
	jwtClaims := jwt.MapClaims{
		"jti":       jti,
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
		"iss":       m.config.Issuer,
		"aud":       m.config.Audience,
		"tenant_id": tenantID,
		"app_id":    appID,
	}

	// Add custom claims
	for k, v := range claims {
		jwtClaims[k] = v
	}

	// Create token
	jwtToken := jwt.NewWithClaims(m.config.SigningMethod, jwtClaims)

	// Sign token
	tokenString, err := jwtToken.SignedString(m.config.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &token.Token{
		Value:     tokenString,
		Type:      "Bearer",
		TenantID:  tenantID,
		AppID:     appID,
		ExpiresAt: expiresAt,
		IssuedAt:  now,
		Metadata: map[string]any{
			"algorithm": m.config.SigningMethod.Alg(),
		},
	}, nil
}

// Verify validates a JWT token and extracts its claims
func (m *Manager) Verify(ctx context.Context, tokenValue string) (*token.VerificationResult, error) {
	// Parse and verify token
	jwtToken, err := jwt.Parse(tokenValue, func(t *jwt.Token) (any, error) {
		// Verify signing method
		if t.Method.Alg() != m.config.SigningMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.config.VerifyingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &token.VerificationResult{
				Valid: false,
				Error: ErrExpiredToken,
			}, nil
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return &token.VerificationResult{
				Valid: false,
				Error: ErrInvalidSignature,
			}, nil
		}
		return &token.VerificationResult{
			Valid: false,
			Error: ErrInvalidToken,
		}, nil
	}

	// Extract claims
	jwtClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return &token.VerificationResult{
			Valid: false,
			Error: ErrInvalidToken,
		}, nil
	}

	// Check revocation using JTI only
	if m.config.EnableRevocation && m.revocationList != nil {
		// Get JTI (JWT ID) from claims - all tokens now have JTI
		if jti, ok := jwtClaims["jti"].(string); ok && jti != "" {
			revoked, err := m.revocationList.IsRevoked(ctx, jti)
			if err == nil && revoked {
				return &token.VerificationResult{
					Valid: false,
					Error: errors.New("token has been revoked"),
				}, nil
			}
		}
	}

	// Verify issuer
	if m.config.Issuer != "" {
		iss, err := jwtClaims.GetIssuer()
		if err != nil || iss != m.config.Issuer {
			return &token.VerificationResult{
				Valid: false,
				Error: fmt.Errorf("invalid issuer"),
			}, nil
		}
	}

	// Convert to token.Claims
	claims := make(token.Claims)
	for k, v := range jwtClaims {
		claims[k] = v
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
			"algorithm": jwtToken.Method.Alg(),
			"tenant_id": tenantID,
			"app_id":    appID,
		},
	}, nil
}

// Type returns the type of tokens this manager handles
func (m *Manager) Type() string {
	return "jwt"
}

// GenerateRefreshToken generates a refresh token
func (m *Manager) GenerateRefreshToken(ctx context.Context, claims token.Claims) (*token.Token, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.RefreshTokenDuration)

	// Validate required multi-tenant claims
	tenantID, hasTenant := claims.GetTenantID()
	if !hasTenant || tenantID == "" {
		return nil, errors.New("tenant_id is required in claims")
	}

	appID, hasApp := claims.GetAppID()
	if !hasApp || appID == "" {
		return nil, errors.New("app_id is required in claims")
	}

	// Generate unique JTI (JWT ID) for this refresh token
	jti, err := generateJTI()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JTI: %w", err)
	}

	// Build JWT claims for refresh token
	jwtClaims := jwt.MapClaims{
		"jti":       jti,
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
		"iss":       m.config.Issuer,
		"aud":       m.config.Audience,
		"type":      "refresh",
		"tenant_id": tenantID,
		"app_id":    appID,
	}

	// Add limited custom claims (typically just subject)
	if sub, ok := claims["sub"]; ok {
		jwtClaims["sub"] = sub
	}

	// Create token
	jwtToken := jwt.NewWithClaims(m.config.SigningMethod, jwtClaims)

	// Sign token
	tokenString, err := jwtToken.SignedString(m.config.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &token.Token{
		Value:     tokenString,
		Type:      "Bearer",
		TenantID:  tenantID,
		AppID:     appID,
		ExpiresAt: expiresAt,
		IssuedAt:  now,
		Metadata: map[string]any{
			"algorithm": m.config.SigningMethod.Alg(),
			"type":      "refresh",
		},
	}, nil
}

// Revoke revokes a token by adding it to the revocation list
func (m *Manager) Revoke(ctx context.Context, tokenValue string) error {
	if !m.config.EnableRevocation || m.revocationList == nil {
		return errors.New("revocation not enabled")
	}

	// Parse token to get JTI and expiry
	jwtToken, err := jwt.Parse(tokenValue, func(t *jwt.Token) (any, error) {
		return m.config.VerifyingKey, nil
	})

	if err != nil {
		return err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}

	// Get token ID (JTI) - all tokens now have JTI
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return errors.New("token has no JTI identifier")
	}

	// Get expiry
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return err
	}

	return m.revocationList.Add(ctx, jti, exp.Time)
}

// Refresh generates a new access token from a refresh token
func (m *Manager) Refresh(ctx context.Context, refreshToken string) (*token.Token, error) {
	// Verify refresh token
	result, err := m.Verify(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if !result.Valid {
		return nil, result.Error
	}

	// Check if it's a refresh token
	tokenType, ok := result.Claims["type"]
	if !ok || tokenType != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	// Generate new access token with same subject
	return m.Generate(ctx, result.Claims)
}

// InMemoryRevocationList is an in-memory implementation of TokenRevocationList
type InMemoryRevocationList struct {
	revoked map[string]time.Time
}

// NewInMemoryRevocationList creates a new in-memory revocation list
func NewInMemoryRevocationList() *InMemoryRevocationList {
	return &InMemoryRevocationList{
		revoked: make(map[string]time.Time),
	}
}

func (r *InMemoryRevocationList) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	r.revoked[tokenID] = expiresAt
	return nil
}

func (r *InMemoryRevocationList) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	_, revoked := r.revoked[tokenID]
	return revoked, nil
}

func (r *InMemoryRevocationList) Remove(ctx context.Context, tokenID string) error {
	delete(r.revoked, tokenID)
	return nil
}

func (r *InMemoryRevocationList) Cleanup(ctx context.Context) error {
	now := time.Now()
	for tokenID, expiresAt := range r.revoked {
		if now.After(expiresAt) {
			delete(r.revoked, tokenID)
		}
	}
	return nil
}

// generateJTI generates a unique JWT ID using cryptographic random bytes
func generateJTI() (string, error) {
	b := make([]byte, 16) // 128-bit random ID
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
