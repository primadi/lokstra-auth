package token

import (
	"context"
	"time"
)

// Token represents a security token
type Token struct {
	// Value is the actual token string
	Value string

	// Type is the token type (e.g., "Bearer", "JWT")
	Type string

	// TenantID is the tenant this token belongs to (REQUIRED for multi-tenant)
	TenantID string

	// AppID is the app this token was issued for (REQUIRED for multi-tenant)
	AppID string

	// BranchID is the branch this token is scoped to (optional)
	BranchID string

	// ExpiresAt indicates when the token expires
	ExpiresAt time.Time

	// IssuedAt indicates when the token was issued
	IssuedAt time.Time

	// Metadata contains additional token metadata
	Metadata map[string]any
}

// VerificationResult represents the result of token verification
type VerificationResult struct {
	// Valid indicates whether the token is valid
	Valid bool

	// Claims contains the extracted claims
	Claims Claims

	// Error contains the error if verification failed
	Error error

	// Metadata contains additional verification metadata
	Metadata map[string]any
}

// TokenGenerator generates tokens from claims
type TokenGenerator interface {
	// Generate creates a new token from the provided claims
	Generate(ctx context.Context, claims Claims) (*Token, error)

	// Type returns the type of tokens this generator creates
	Type() string
}

// TokenVerifier verifies and validates tokens
type TokenVerifier interface {
	// Verify validates a token and extracts its claims
	Verify(ctx context.Context, tokenValue string) (*VerificationResult, error)

	// Type returns the type of tokens this verifier handles
	Type() string
}

// ClaimExtractor extracts specific claims from a token
type ClaimExtractor interface {
	// Extract extracts claims from a token
	Extract(ctx context.Context, token *Token) (Claims, error)

	// ExtractClaim extracts a specific claim by key
	ExtractClaim(ctx context.Context, token *Token, key string) (any, error)
}

// TokenManager combines generation, verification, and refresh operations
type TokenManager interface {
	TokenGenerator
	TokenVerifier

	// GenerateRefreshToken creates a refresh token from claims
	GenerateRefreshToken(ctx context.Context, claims Claims) (*Token, error)

	// Refresh generates a new access token from a refresh token
	Refresh(ctx context.Context, refreshToken string) (*Token, error)

	// Revoke invalidates a refresh token
	Revoke(ctx context.Context, refreshToken string) error

	// GenerateResetToken generates a one-time password reset token for email
	GenerateResetToken(ctx context.Context, email string) (string, error)
}
