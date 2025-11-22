package domain

// =============================================================================
// Credential Configuration
// =============================================================================

// CredentialConfig configures authentication credential providers
type CredentialConfig struct {
	// Basic authentication (username/password)
	EnableBasic bool                   `json:"enable_basic"`
	BasicConfig *BasicCredentialConfig `json:"basic_config,omitempty"`

	// API Key authentication
	EnableAPIKey bool                    `json:"enable_apikey"`
	APIKeyConfig *APIKeyCredentialConfig `json:"apikey_config,omitempty"`

	// OAuth2 authentication
	EnableOAuth2 bool                    `json:"enable_oauth2"`
	OAuth2Config *OAuth2CredentialConfig `json:"oauth2_config,omitempty"`

	// Passwordless authentication (email/SMS magic link)
	EnablePasswordless bool                          `json:"enable_passwordless"`
	PasswordlessConfig *PasswordlessCredentialConfig `json:"passwordless_config,omitempty"`

	// Passkey authentication (WebAuthn)
	EnablePasskey bool                     `json:"enable_passkey"`
	PasskeyConfig *PasskeyCredentialConfig `json:"passkey_config,omitempty"`
}

// BasicCredentialConfig configuration for basic authentication
type BasicCredentialConfig struct {
	// Validation rules
	MinUsernameLength int  `json:"min_username_length"` // Default: 3
	MaxUsernameLength int  `json:"max_username_length"` // Default: 32
	MinPasswordLength int  `json:"min_password_length"` // Default: 8
	RequireStrongPwd  bool `json:"require_strong_pwd"`  // Default: true (uppercase, lowercase, number)

	// Security settings
	MaxLoginAttempts    int `json:"max_login_attempts"`    // Default: 5
	LockoutDurationSecs int `json:"lockout_duration_secs"` // Default: 300 (5 minutes)

	// Session settings
	SessionTimeoutSecs int `json:"session_timeout_secs"` // Default: 3600 (1 hour)
}

// APIKeyCredentialConfig configuration for API key authentication
type APIKeyCredentialConfig struct {
	// Key generation
	SecretLength int    `json:"secret_length"` // Default: 32 bytes (256-bit)
	HashAlgo     string `json:"hash_algo"`     // "sha3-256" or "sha256", default: sha3-256

	// Key lifecycle
	DefaultExpiryDays int  `json:"default_expiry_days"` // Default: 365, 0 = never expires
	AllowNeverExpire  bool `json:"allow_never_expire"`  // Default: true

	// Rate limiting per key
	RateLimitPerMinute int `json:"rate_limit_per_minute"` // Default: 60
}

// OAuth2CredentialConfig configuration for OAuth2 providers
type OAuth2CredentialConfig struct {
	// Enabled providers
	Providers []OAuth2ProviderConfig `json:"providers"`

	// OAuth2 flow settings
	CallbackURL     string `json:"callback_url"`      // e.g., https://api.example.com/api/cred/oauth2/callback
	StateExpirySecs int    `json:"state_expiry_secs"` // Default: 600 (10 minutes)

	// Session settings
	SessionTimeoutSecs int `json:"session_timeout_secs"` // Default: 3600 (1 hour)
}

// OAuth2ProviderConfig configuration for a specific OAuth2 provider
type OAuth2ProviderConfig struct {
	Name         string   `json:"name"`          // "google", "azure", "github", "facebook", etc.
	Enabled      bool     `json:"enabled"`       // Enable/disable this provider
	ClientID     string   `json:"client_id"`     // OAuth2 client ID
	ClientSecret string   `json:"client_secret"` // OAuth2 client secret
	Scopes       []string `json:"scopes"`        // Requested scopes

	// Provider-specific settings
	AuthURL  string `json:"auth_url,omitempty"`  // Authorization URL (if custom)
	TokenURL string `json:"token_url,omitempty"` // Token URL (if custom)
	UserURL  string `json:"user_url,omitempty"`  // User info URL (if custom)
}

// PasswordlessCredentialConfig configuration for passwordless authentication
type PasswordlessCredentialConfig struct {
	// Delivery methods
	EnableEmail bool `json:"enable_email"` // Magic link via email
	EnableSMS   bool `json:"enable_sms"`   // OTP via SMS

	// Code/Link settings
	CodeLength     int `json:"code_length"`      // Default: 6 (for OTP)
	CodeExpirySecs int `json:"code_expiry_secs"` // Default: 300 (5 minutes)
	LinkExpirySecs int `json:"link_expiry_secs"` // Default: 600 (10 minutes)

	// Rate limiting
	MaxAttemptsPerEmail int `json:"max_attempts_per_email"` // Default: 5
	MaxAttemptsPerPhone int `json:"max_attempts_per_phone"` // Default: 5
	CooldownSecs        int `json:"cooldown_secs"`          // Default: 60
}

// PasskeyCredentialConfig configuration for passkey (WebAuthn) authentication
type PasskeyCredentialConfig struct {
	// Relying Party settings
	RPName    string   `json:"rp_name"`    // e.g., "My Application"
	RPID      string   `json:"rp_id"`      // e.g., "example.com"
	RPOrigins []string `json:"rp_origins"` // e.g., ["https://example.com", "https://www.example.com"]

	// Authenticator settings
	RequireResidentKey    bool   `json:"require_resident_key"`   // Default: false
	UserVerification      string `json:"user_verification"`      // "required", "preferred", "discouraged"
	AttestationPreference string `json:"attestation_preference"` // "none", "indirect", "direct"

	// Timeout settings
	TimeoutSecs int `json:"timeout_secs"` // Default: 60
}

// DefaultCredentialConfig returns default credential configuration
func DefaultCredentialConfig() *CredentialConfig {
	return &CredentialConfig{
		EnableBasic: true,
		BasicConfig: &BasicCredentialConfig{
			MinUsernameLength:   3,
			MaxUsernameLength:   32,
			MinPasswordLength:   8,
			RequireStrongPwd:    true,
			MaxLoginAttempts:    5,
			LockoutDurationSecs: 300,
			SessionTimeoutSecs:  3600,
		},
		EnableAPIKey: true,
		APIKeyConfig: &APIKeyCredentialConfig{
			SecretLength:       32,
			HashAlgo:           "sha3-256",
			DefaultExpiryDays:  365,
			AllowNeverExpire:   true,
			RateLimitPerMinute: 60,
		},
		EnableOAuth2:       false, // Disabled by default, requires configuration
		EnablePasswordless: false,
		EnablePasskey:      false,
	}
}
