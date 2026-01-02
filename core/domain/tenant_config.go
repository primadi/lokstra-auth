package domain

// TenantConfig holds tenant-level configuration
type TenantConfig struct {
	// Default credential configuration for all apps in this tenant
	// Apps can override this in their AppConfig.Credentials
	DefaultCredentials *CredentialConfig `json:"default_credentials"`

	// Default token configuration
	DefaultTokenConfig *TokenConfig `json:"default_token_config"`

	// Tenant-wide security settings
	Security *TenantSecurityConfig `json:"security"`
}

// TenantSecurityConfig tenant-wide security settings
type TenantSecurityConfig struct {
	// Password policy (applies to all apps unless overridden)
	EnforceStrongPassword bool `json:"enforce_strong_password"`
	MinPasswordLength     int  `json:"min_password_length"`

	// Rate limiting
	GlobalRateLimitPerMinute int `json:"global_rate_limit_per_minute"`

	// IP whitelist/blacklist
	AllowedIPs []string `json:"allowed_ips,omitempty"`
	BlockedIPs []string `json:"blocked_ips,omitempty"`

	// Session settings
	MaxSessionDuration      int  `json:"max_session_duration"` // seconds
	AllowConcurrentSessions bool `json:"allow_concurrent_sessions"`
}

// TokenConfig holds token generation configuration
type TokenConfig struct {
	AccessTokenExpiry  int    `json:"access_token_expiry"`  // seconds
	RefreshTokenExpiry int    `json:"refresh_token_expiry"` // seconds
	TokenAlgorithm     string `json:"token_algorithm"`      // HS256, RS256, etc.
	TokenSecret        string `json:"token_secret,omitempty"`
}
