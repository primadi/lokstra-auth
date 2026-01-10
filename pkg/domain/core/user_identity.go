package core

import (
	"errors"
	"time"
)

var (
	ErrUserIdentityNotFound      = errors.New("user identity not found")
	ErrUserIdentityAlreadyExists = errors.New("user identity already exists")
	ErrInvalidProviderID         = errors.New("invalid provider ID")
	ErrDuplicateProviderIdentity = errors.New("provider identity already linked to another user")
)

// IdentityProvider represents the authentication provider
type IdentityProvider string

const (
	IdentityProviderLocal     IdentityProvider = "local"      // Username/password
	IdentityProviderGoogle    IdentityProvider = "google"     // Google OAuth2
	IdentityProviderGitHub    IdentityProvider = "github"     // GitHub OAuth2
	IdentityProviderFacebook  IdentityProvider = "facebook"   // Facebook OAuth2
	IdentityProviderMicrosoft IdentityProvider = "microsoft"  // Microsoft OAuth2
	IdentityProviderApple     IdentityProvider = "apple"      // Apple OAuth2
	IdentityProviderLinkedIn  IdentityProvider = "linkedin"   // LinkedIn OAuth2
	IdentityProviderTwitter   IdentityProvider = "twitter"    // Twitter OAuth2
	IdentityProviderSAML      IdentityProvider = "saml"       // SAML SSO
	IdentityProviderOIDC      IdentityProvider = "oidc"       // Generic OIDC
	IdentityProviderPasskey   IdentityProvider = "passkey"    // WebAuthn/FIDO2
	IdentityProviderMagicLink IdentityProvider = "magic_link" // Passwordless email
	IdentityProviderOTP       IdentityProvider = "otp"        // Passwordless OTP
	IdentityProviderLDAP      IdentityProvider = "ldap"       // LDAP/Active Directory
	IdentityProviderAPIKey    IdentityProvider = "apikey"     // API Key (for services)
)

// UserIdentity represents a linked identity from an authentication provider
// Each user can have multiple identities (e.g., email + Google + GitHub)
type UserIdentity struct {
	ID         string           `json:"id"`                 // Unique identifier (UUID)
	UserID     string           `json:"user_id"`            // Foreign key to users table
	TenantID   string           `json:"tenant_id"`          // Belongs to tenant
	Provider   IdentityProvider `json:"provider"`           // Authentication provider
	ProviderID string           `json:"provider_id"`        // Unique ID from provider (e.g., Google sub, GitHub ID)
	Email      string           `json:"email"`              // Email from provider (may differ from user.email)
	Username   string           `json:"username"`           // Username from provider (optional)
	Verified   bool             `json:"verified"`           // Email verification status from provider
	Metadata   *map[string]any  `json:"metadata,omitempty"` // Additional provider-specific data
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// Validate validates user identity data
func (ui *UserIdentity) Validate() error {
	if ui.ID == "" {
		return errors.New("identity ID is required")
	}
	if ui.UserID == "" {
		return ErrInvalidUserID
	}
	if ui.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if ui.Provider == "" {
		return errors.New("provider is required")
	}
	if ui.ProviderID == "" {
		return ErrInvalidProviderID
	}
	return nil
}

// IsOAuth2Provider checks if this is an OAuth2 provider
func (ui *UserIdentity) IsOAuth2Provider() bool {
	return ui.Provider == IdentityProviderGoogle ||
		ui.Provider == IdentityProviderGitHub ||
		ui.Provider == IdentityProviderFacebook ||
		ui.Provider == IdentityProviderMicrosoft ||
		ui.Provider == IdentityProviderApple ||
		ui.Provider == IdentityProviderLinkedIn ||
		ui.Provider == IdentityProviderTwitter
}

// IsPasswordlessProvider checks if this is a passwordless provider
func (ui *UserIdentity) IsPasswordlessProvider() bool {
	return ui.Provider == IdentityProviderMagicLink ||
		ui.Provider == IdentityProviderOTP
}

// IsLocalProvider checks if this is local authentication
func (ui *UserIdentity) IsLocalProvider() bool {
	return ui.Provider == IdentityProviderLocal
}
