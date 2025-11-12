package subject

import (
	"context"
)

// Subject represents an authenticated entity
type Subject struct {
	// ID is the unique identifier for the subject
	ID string

	// Type indicates the subject type (e.g., "user", "service", "device")
	Type string

	// Principal is the primary identifier (e.g., username, email, service name)
	Principal string

	// Attributes contains subject attributes
	Attributes map[string]any
}

// IdentityContext represents the complete identity context
type IdentityContext struct {
	// Subject is the authenticated subject
	Subject *Subject

	// Roles contains the subject's roles
	Roles []string

	// Permissions contains the subject's permissions
	Permissions []string

	// Groups contains the subject's group memberships
	Groups []string

	// Profile contains additional profile information
	Profile map[string]any

	// Session contains session-specific information
	Session *SessionInfo

	// Metadata contains additional context metadata
	Metadata map[string]any
}

// SessionInfo contains session-specific information
type SessionInfo struct {
	// ID is the session identifier
	ID string

	// CreatedAt is when the session was created
	CreatedAt int64

	// ExpiresAt is when the session expires
	ExpiresAt int64

	// IPAddress is the client IP address
	IPAddress string

	// UserAgent is the client user agent
	UserAgent string

	// Metadata contains additional session metadata
	Metadata map[string]any
}

// HasRole checks if the identity has a specific role
func (ic *IdentityContext) HasRole(role string) bool {
	for _, r := range ic.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the identity has a specific permission
func (ic *IdentityContext) HasPermission(permission string) bool {
	for _, p := range ic.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the identity has any of the specified roles
func (ic *IdentityContext) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if ic.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the identity has all of the specified roles
func (ic *IdentityContext) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !ic.HasRole(role) {
			return false
		}
	}
	return true
}

// SubjectResolver resolves a subject from claims
type SubjectResolver interface {
	// Resolve creates a Subject from claims
	Resolve(ctx context.Context, claims map[string]any) (*Subject, error)
}

// IdentityContextBuilder builds a complete identity context
type IdentityContextBuilder interface {
	// Build creates an IdentityContext from a subject
	Build(ctx context.Context, subject *Subject) (*IdentityContext, error)
}

// RoleProvider provides roles for a subject
type RoleProvider interface {
	// GetRoles retrieves roles for a subject
	GetRoles(ctx context.Context, subject *Subject) ([]string, error)
}

// PermissionProvider provides permissions for a subject
type PermissionProvider interface {
	// GetPermissions retrieves permissions for a subject
	GetPermissions(ctx context.Context, subject *Subject) ([]string, error)
}

// GroupProvider provides group memberships for a subject
type GroupProvider interface {
	// GetGroups retrieves groups for a subject
	GetGroups(ctx context.Context, subject *Subject) ([]string, error)
}

// ProfileProvider provides profile information for a subject
type ProfileProvider interface {
	// GetProfile retrieves profile information for a subject
	GetProfile(ctx context.Context, subject *Subject) (map[string]any, error)
}

// DataEnricher enriches identity context with additional data
type DataEnricher interface {
	// Enrich adds additional data to the identity context
	Enrich(ctx context.Context, identity *IdentityContext) error
}

// IdentityStore stores and retrieves identity contexts
type IdentityStore interface {
	// Store saves an identity context
	Store(ctx context.Context, sessionID string, identity *IdentityContext) error

	// Get retrieves an identity context
	Get(ctx context.Context, sessionID string) (*IdentityContext, error)

	// Delete removes an identity context
	Delete(ctx context.Context, sessionID string) error

	// Update updates an identity context
	Update(ctx context.Context, sessionID string, identity *IdentityContext) error
}

// IdentityCache caches identity contexts for performance
type IdentityCache interface {
	// Set caches an identity context
	Set(ctx context.Context, key string, identity *IdentityContext, ttl int64) error

	// Get retrieves a cached identity context
	Get(ctx context.Context, key string) (*IdentityContext, error)

	// Delete removes a cached identity context
	Delete(ctx context.Context, key string) error

	// Clear clears all cached identity contexts
	Clear(ctx context.Context) error
}
