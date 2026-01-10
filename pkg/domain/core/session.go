package core

import "time"

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

// Session represents an active user session
type Session struct {
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	UserID           string          `json:"user_id"`
	AppID            *string         `json:"app_id,omitempty"`
	SessionTokenHash string          `json:"-"` // Never expose in JSON
	IPAddress        *string         `json:"ip_address,omitempty"`
	UserAgent        *string         `json:"user_agent,omitempty"`
	Status           SessionStatus   `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	LastActivityAt   time.Time       `json:"last_activity_at"`
	ExpiresAt        time.Time       `json:"expires_at"`
	RevokedAt        *time.Time      `json:"revoked_at,omitempty"`
	Metadata         *map[string]any `json:"metadata,omitempty"`
}

// IsActive checks if session is active
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive
}

// IsExpired checks if session is expired
func (s *Session) IsExpired() bool {
	return s.Status == SessionStatusExpired || time.Now().After(s.ExpiresAt)
}

// IsRevoked checks if session is revoked
func (s *Session) IsRevoked() bool {
	return s.Status == SessionStatusRevoked
}
