package application

import (
	"context"

	coredomain "github.com/primadi/lokstra-auth/00_core/domain"
)

// ConfigResolver resolves credential configuration with proper inheritance
// Order: App Config → Tenant Config → Global Default
type ConfigResolver struct {
	tenantStore interface {
		Get(ctx context.Context, tenantID string) (*coredomain.Tenant, error)
	}
	appStore interface {
		Get(ctx context.Context, tenantID, appID string) (*coredomain.App, error)
	}
}

// GetEffectiveConfig returns the effective credential config for an app
// Priority: App Config > Tenant Config > Global Default
func (r *ConfigResolver) GetEffectiveConfig(ctx context.Context, tenantID, appID string) (*coredomain.CredentialConfig, error) {
	// Try to get app config
	if appID != "" {
		app, err := r.appStore.Get(ctx, tenantID, appID)
		if err == nil && app.Config != nil && app.Config.Credentials != nil {
			return app.Config.Credentials, nil
		}
	}

	// Try to get tenant default config
	tenant, err := r.tenantStore.Get(ctx, tenantID)
	if err == nil && tenant.Config != nil && tenant.Config.DefaultCredentials != nil {
		return tenant.Config.DefaultCredentials, nil
	}

	// Return global default
	return coredomain.DefaultCredentialConfig(), nil
}

// IsBasicEnabled checks if basic auth is enabled
func (r *ConfigResolver) IsBasicEnabled(ctx context.Context, tenantID, appID string) bool {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil {
		return false
	}
	return config.EnableBasic
}

// IsAPIKeyEnabled checks if API key auth is enabled
func (r *ConfigResolver) IsAPIKeyEnabled(ctx context.Context, tenantID, appID string) bool {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil {
		return false
	}
	return config.EnableAPIKey
}

// IsOAuth2Enabled checks if OAuth2 auth is enabled
func (r *ConfigResolver) IsOAuth2Enabled(ctx context.Context, tenantID, appID string) bool {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil {
		return false
	}
	return config.EnableOAuth2
}

// GetBasicConfig returns basic auth config with defaults
func (r *ConfigResolver) GetBasicConfig(ctx context.Context, tenantID, appID string) *coredomain.BasicCredentialConfig {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil || config.BasicConfig == nil {
		return &coredomain.BasicCredentialConfig{
			MinUsernameLength:   3,
			MaxUsernameLength:   32,
			MinPasswordLength:   8,
			RequireStrongPwd:    true,
			MaxLoginAttempts:    5,
			LockoutDurationSecs: 300,
			SessionTimeoutSecs:  3600,
		}
	}
	return config.BasicConfig
}

// GetAPIKeyConfig returns API key config with defaults
func (r *ConfigResolver) GetAPIKeyConfig(ctx context.Context, tenantID, appID string) *coredomain.APIKeyCredentialConfig {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil || config.APIKeyConfig == nil {
		return &coredomain.APIKeyCredentialConfig{
			SecretLength:       32,
			HashAlgo:           "sha3-256",
			DefaultExpiryDays:  365,
			AllowNeverExpire:   true,
			RateLimitPerMinute: 60,
		}
	}
	return config.APIKeyConfig
}

// GetOAuth2Config returns OAuth2 config
func (r *ConfigResolver) GetOAuth2Config(ctx context.Context, tenantID, appID string) *coredomain.OAuth2CredentialConfig {
	config, err := r.GetEffectiveConfig(ctx, tenantID, appID)
	if err != nil || config.OAuth2Config == nil {
		return nil
	}
	return config.OAuth2Config
}
