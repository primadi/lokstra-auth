package core

import (
	"time"

	"github.com/google/uuid"
)

// RefreshTokenStatus represents the status of a refresh token
type RefreshTokenStatus string

const (
	RefreshTokenStatusActive  RefreshTokenStatus = "active"
	RefreshTokenStatusUsed    RefreshTokenStatus = "used"
	RefreshTokenStatusExpired RefreshTokenStatus = "expired"
	RefreshTokenStatusRevoked RefreshTokenStatus = "revoked"
)

// RefreshToken represents a refresh token for OAuth2/JWT authentication
type RefreshToken struct {
	ID            uuid.UUID          `json:"id"`
	TenantID      string             `json:"tenant_id"`
	UserID        string             `json:"user_id"`
	AppID         *string            `json:"app_id,omitempty"`
	SessionID     *string            `json:"session_id,omitempty"`
	TokenHash     string             `json:"-"` // Never expose in JSON
	TokenFamily   *string            `json:"token_family,omitempty"`
	RotatedFromID *uuid.UUID         `json:"rotated_from_id,omitempty"` // Previous token ID in rotation chain
	Status        RefreshTokenStatus `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
	ExpiresAt     time.Time          `json:"expires_at"`
	UsedAt        *time.Time         `json:"used_at,omitempty"`
	RevokedAt     *time.Time         `json:"revoked_at,omitempty"`
	IPAddress     *string            `json:"ip_address,omitempty"`
	UserAgent     *string            `json:"user_agent,omitempty"`
	Metadata      *map[string]any    `json:"metadata,omitempty"`
}

// IsActive checks if token is active
func (t *RefreshToken) IsActive() bool {
	return t.Status == RefreshTokenStatusActive
}

// IsExpired checks if token is expired
func (t *RefreshToken) IsExpired() bool {
	return t.Status == RefreshTokenStatusExpired || time.Now().After(t.ExpiresAt)
}

// IsUsed checks if token has been used
func (t *RefreshToken) IsUsed() bool {
	return t.Status == RefreshTokenStatusUsed
}

// IsRevoked checks if token is revoked
func (t *RefreshToken) IsRevoked() bool {
	return t.Status == RefreshTokenStatusRevoked
}

// IsWithinGracePeriod checks if token is within grace period after being used
// Grace period allows recently-used tokens to be valid for multi-tab scenarios
func (t *RefreshToken) IsWithinGracePeriod(gracePeriodSeconds int) bool {
	if t.Status != RefreshTokenStatusUsed || t.UsedAt == nil {
		return false
	}

	gracePeriod := time.Duration(gracePeriodSeconds) * time.Second
	return time.Since(*t.UsedAt) <= gracePeriod
}

// CanBeUsedWithGracePeriod checks if token can be used considering grace period
func (t *RefreshToken) CanBeUsedWithGracePeriod(gracePeriodSeconds int) bool {
	// Active tokens can always be used
	if t.Status == RefreshTokenStatusActive {
		return true
	}

	// Used tokens can be reused within grace period
	if t.Status == RefreshTokenStatusUsed && gracePeriodSeconds > 0 {
		return t.IsWithinGracePeriod(gracePeriodSeconds)
	}

	return false
}
