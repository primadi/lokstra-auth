package core

// SessionRevocationPolicy determines what to do when max concurrent sessions is reached
type SessionRevocationPolicy string

const (
	// SessionRevocationPolicyOldest revokes the oldest session when limit reached
	SessionRevocationPolicyOldest SessionRevocationPolicy = "oldest"
	// SessionRevocationPolicyNewest revokes the newest session when limit reached (reject new login)
	SessionRevocationPolicyNewest SessionRevocationPolicy = "newest"
	// SessionRevocationPolicyReject rejects new login attempt when limit reached
	SessionRevocationPolicyReject SessionRevocationPolicy = "reject"
)

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
	MaxSessionDuration         int                     `json:"max_session_duration"` // seconds
	AllowConcurrentSessions    bool                    `json:"allow_concurrent_sessions"`
	MaxConcurrentSessions      int                     `json:"max_concurrent_sessions"`        // 0 = unlimited, >0 = limit
	SessionRevocationPolicy    SessionRevocationPolicy `json:"session_revocation_policy"`      // oldest, newest, reject
	MaxRefreshTokensPerSession int                     `json:"max_refresh_tokens_per_session"` // 0 = unlimited

	// Refresh Token Rotation settings
	EnableRefreshTokenRotation      bool `json:"enable_refresh_token_rotation"`       // Rotate refresh token on each use
	RefreshTokenRotationGracePeriod int  `json:"refresh_token_rotation_grace_period"` // seconds, allow old token during this period (for multi-tab)
	DetectRefreshTokenReuse         bool `json:"detect_refresh_token_reuse"`          // Detect and revoke family if reuse detected
}

// TokenConfig holds token generation configuration
type TokenConfig struct {
	AccessTokenExpiry  int    `json:"access_token_expiry"`  // seconds
	RefreshTokenExpiry int    `json:"refresh_token_expiry"` // seconds
	TokenAlgorithm     string `json:"token_algorithm"`      // HS256, RS256, etc.
	TokenSecret        string `json:"token_secret,omitempty"`

	// Refresh token rotation (can override tenant-level settings)
	EnableRotation      *bool `json:"enable_rotation,omitempty"`       // nil = use tenant default
	RotationGracePeriod *int  `json:"rotation_grace_period,omitempty"` // nil = use tenant default
}
