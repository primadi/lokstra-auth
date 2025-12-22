package domain

import (
	"errors"
	"time"
)

var (
	ErrProviderNotFound      = errors.New("credential provider not found")
	ErrProviderAlreadyExists = errors.New("credential provider already exists")
	ErrInvalidProviderType   = errors.New("invalid provider type")
	ErrProviderDisabled      = errors.New("provider is disabled")
)

// ProviderType represents the type of credential provider
type ProviderType string

const (
	// OAuth2 Providers
	ProviderTypeGoogle    ProviderType = "google"
	ProviderTypeGitHub    ProviderType = "github"
	ProviderTypeMicrosoft ProviderType = "microsoft"
	ProviderTypeFacebook  ProviderType = "facebook"
	ProviderTypeApple     ProviderType = "apple"
	ProviderTypeLinkedIn  ProviderType = "linkedin"
	ProviderTypeTwitter   ProviderType = "twitter"

	// Enterprise SSO
	ProviderTypeSAML ProviderType = "saml"
	ProviderTypeOIDC ProviderType = "oidc"
	ProviderTypeLDAP ProviderType = "ldap"

	// Email/SMS
	ProviderTypeEmail ProviderType = "email" // For magic link/OTP
	ProviderTypeSMS   ProviderType = "sms"   // For OTP

	// Others
	ProviderTypeWebAuthn ProviderType = "webauthn" // Passkey/FIDO2
)

// ProviderStatus represents the status of a provider
type ProviderStatus string

const (
	ProviderStatusActive   ProviderStatus = "active"
	ProviderStatusDisabled ProviderStatus = "disabled"
)

// CredentialProvider represents a configured authentication provider
// This is per tenant+app, allowing multiple configurations of same provider type
// Example: tenant "acme" app "web" can have 2 Google providers (prod + dev)
type CredentialProvider struct {
	ID          string          `json:"id"`          // Unique ID (UUID)
	TenantID    string          `json:"tenant_id"`   // Belongs to tenant
	AppID       string          `json:"app_id"`      // Belongs to app (NULL = tenant-level default)
	Type        ProviderType    `json:"type"`        // Provider type
	Name        string          `json:"name"`        // Display name (e.g., "Google Production")
	Description string          `json:"description"` // Optional description
	Status      ProviderStatus  `json:"status"`      // active, disabled
	Config      map[string]any  `json:"config"`      // Provider-specific configuration
	Metadata    *map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// OAuth2Config extracts OAuth2-specific configuration
func (p *CredentialProvider) OAuth2Config() *OAuth2ProviderDetail {
	if !p.IsOAuth2() {
		return nil
	}

	return &OAuth2ProviderDetail{
		ClientID:     p.getConfigString("client_id"),
		ClientSecret: p.getConfigString("client_secret"),
		RedirectURL:  p.getConfigString("redirect_url"),
		Scopes:       p.getConfigStringSlice("scopes"),
		AuthURL:      p.getConfigString("auth_url"),
		TokenURL:     p.getConfigString("token_url"),
		UserInfoURL:  p.getConfigString("user_info_url"),
	}
}

// SAMLConfig extracts SAML-specific configuration
func (p *CredentialProvider) SAMLConfig() *SAMLProviderDetail {
	if p.Type != ProviderTypeSAML {
		return nil
	}

	return &SAMLProviderDetail{
		EntityID:             p.getConfigString("entity_id"),
		SSOURL:               p.getConfigString("sso_url"),
		Certificate:          p.getConfigString("certificate"),
		SignAuthnRequests:    p.getConfigBool("sign_authn_requests"),
		WantAssertionsSigned: p.getConfigBool("want_assertions_signed"),
	}
}

// EmailConfig extracts email provider configuration
func (p *CredentialProvider) EmailConfig() *EmailProviderDetail {
	if p.Type != ProviderTypeEmail {
		return nil
	}

	return &EmailProviderDetail{
		SMTPHost:     p.getConfigString("smtp_host"),
		SMTPPort:     p.getConfigInt("smtp_port"),
		SMTPUsername: p.getConfigString("smtp_username"),
		SMTPPassword: p.getConfigString("smtp_password"),
		FromAddress:  p.getConfigString("from_address"),
		FromName:     p.getConfigString("from_name"),
	}
}

// Helper methods to extract config values
func (p *CredentialProvider) getConfigString(key string) string {
	if val, ok := p.Config[key].(string); ok {
		return val
	}
	return ""
}

func (p *CredentialProvider) getConfigInt(key string) int {
	if val, ok := p.Config[key].(float64); ok {
		return int(val)
	}
	if val, ok := p.Config[key].(int); ok {
		return val
	}
	return 0
}

func (p *CredentialProvider) getConfigBool(key string) bool {
	if val, ok := p.Config[key].(bool); ok {
		return val
	}
	return false
}

func (p *CredentialProvider) getConfigStringSlice(key string) []string {
	if val, ok := p.Config[key].([]any); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	if val, ok := p.Config[key].([]string); ok {
		return val
	}
	return []string{}
}

// IsOAuth2 checks if this is an OAuth2 provider
func (p *CredentialProvider) IsOAuth2() bool {
	return p.Type == ProviderTypeGoogle ||
		p.Type == ProviderTypeGitHub ||
		p.Type == ProviderTypeMicrosoft ||
		p.Type == ProviderTypeFacebook ||
		p.Type == ProviderTypeApple ||
		p.Type == ProviderTypeLinkedIn ||
		p.Type == ProviderTypeTwitter ||
		p.Type == ProviderTypeOIDC
}

// IsActive checks if provider is active
func (p *CredentialProvider) IsActive() bool {
	return p.Status == ProviderStatusActive
}

// Validate validates provider data
func (p *CredentialProvider) Validate() error {
	if p.ID == "" {
		return errors.New("provider ID is required")
	}
	if p.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	// AppID is optional (NULL = tenant-level default)
	if p.Type == "" {
		return ErrInvalidProviderType
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Config == nil {
		return errors.New("config is required")
	}
	return nil
}

// =============================================================================
// Provider Configuration Details
// =============================================================================

// OAuth2ProviderDetail contains OAuth2-specific configuration
type OAuth2ProviderDetail struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url,omitempty"`      // Custom auth URL
	TokenURL     string   `json:"token_url,omitempty"`     // Custom token URL
	UserInfoURL  string   `json:"user_info_url,omitempty"` // Custom user info URL
}

// SAMLProviderDetail contains SAML-specific configuration
type SAMLProviderDetail struct {
	EntityID             string `json:"entity_id"`
	SSOURL               string `json:"sso_url"`
	Certificate          string `json:"certificate"`            // X.509 certificate
	SignAuthnRequests    bool   `json:"sign_authn_requests"`    // Sign authentication requests
	WantAssertionsSigned bool   `json:"want_assertions_signed"` // Require signed assertions
}

// EmailProviderDetail contains email provider configuration (for magic link/OTP)
type EmailProviderDetail struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
}

// =============================================================================
// DTOs
// =============================================================================

// CreateProviderRequest request to create a credential provider
type CreateProviderRequest struct {
	TenantID    string          `path:"tenant_id" validate:"required"`
	AppID       string          `json:"app_id"` // Optional, NULL = tenant-level
	Type        ProviderType    `json:"type" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Config      map[string]any  `json:"config" validate:"required"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

// GetProviderRequest request to get a provider
type GetProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// UpdateProviderRequest request to update a provider
type UpdateProviderRequest struct {
	TenantID    string          `path:"tenant_id" validate:"required"`
	ProviderID  string          `path:"provider_id" validate:"required"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Config      map[string]any  `json:"config"`
	Status      ProviderStatus  `json:"status"`
	Metadata    *map[string]any `json:"metadata"`
}

// DeleteProviderRequest request to delete a provider
type DeleteProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// ListProvidersRequest request to list providers
type ListProvidersRequest struct {
	TenantID string         `path:"tenant_id" validate:"required"`
	AppID    string         `query:"app_id"` // Filter by app (empty = all)
	Type     ProviderType   `query:"type"`   // Filter by type
	Status   ProviderStatus `query:"status"` // Filter by status
}

// EnableProviderRequest request to enable a provider
type EnableProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// DisableProviderRequest request to disable a provider
type DisableProviderRequest struct {
	TenantID   string `path:"tenant_id" validate:"required"`
	ProviderID string `path:"provider_id" validate:"required"`
}

// GetProvidersByTypeRequest request to get providers by type
type GetProvidersByTypeRequest struct {
	TenantID string       `path:"tenant_id" validate:"required"`
	AppID    string       `query:"app_id"`
	Type     ProviderType `path:"provider_type" validate:"required"`
}

// GetActiveProviderForAppRequest request to get active provider for an app
type GetActiveProviderForAppRequest struct {
	TenantID string       `path:"tenant_id" validate:"required"`
	AppID    string       `query:"app_id"`
	Type     ProviderType `query:"type" validate:"required"`
}
