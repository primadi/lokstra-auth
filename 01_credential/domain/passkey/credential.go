package passkey

import "time"

// Credentials represents WebAuthn passkey credentials
type Credentials struct {
	// Ceremony type: "registration" or "authentication"
	Ceremony string `json:"ceremony"`

	// Credential ID (base64url encoded)
	CredentialID string `json:"credential_id,omitempty"`

	// Client data JSON (from navigator.credentials.create/get)
	ClientDataJSON string `json:"client_data_json"`

	// Authenticator data
	AuthenticatorData string `json:"authenticator_data"`

	// Signature (for authentication)
	Signature string `json:"signature,omitempty"`

	// User handle (for authentication)
	UserHandle string `json:"user_handle,omitempty"`

	// Attestation object (for registration)
	AttestationObject string `json:"attestation_object,omitempty"`
}

// User represents a user with passkey credentials
type User struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PasskeyCredential represents a stored passkey credential
type PasskeyCredential struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	UserID         string     `json:"user_id"`
	CredentialID   string     `json:"credential_id"` // Base64url encoded
	PublicKey      string     `json:"public_key"`    // COSE public key
	SignCount      uint32     `json:"sign_count"`
	AAGUID         string     `json:"aaguid"` // Authenticator AAGUID
	Transports     []string   `json:"transports,omitempty"`
	BackupEligible bool       `json:"backup_eligible"`
	BackupState    bool       `json:"backup_state"`
	DeviceName     string     `json:"device_name,omitempty"`
	LastUsed       *time.Time `json:"last_used,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// =============================================================================
// DTOs
// =============================================================================

// BeginRegistrationRequest request to begin passkey registration
type BeginRegistrationRequest struct {
	TenantID  string `json:"tenant_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// BeginRegistrationResponse response with WebAuthn creation options
type BeginRegistrationResponse struct {
	Success bool `json:"success"`

	// WebAuthn PublicKeyCredentialCreationOptions
	Challenge              string                         `json:"challenge"`
	RelyingParty           RelyingPartyInfo               `json:"rp"`
	User                   UserInfo                       `json:"user"`
	PubKeyCredParams       []PubKeyCredParam              `json:"pubKeyCredParams"`
	AuthenticatorSelection AuthenticatorSelectionCriteria `json:"authenticatorSelection"`
	Timeout                int                            `json:"timeout"`
	Attestation            string                         `json:"attestation"` // "none", "indirect", "direct"

	Error string `json:"error,omitempty"`
}

// FinishRegistrationRequest request to complete passkey registration
type FinishRegistrationRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`

	// WebAuthn response from navigator.credentials.create()
	CredentialID      string   `json:"credential_id" validate:"required"`
	ClientDataJSON    string   `json:"client_data_json" validate:"required"`
	AttestationObject string   `json:"attestation_object" validate:"required"`
	Transports        []string `json:"transports,omitempty"`

	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// FinishRegistrationResponse response after completing registration
type FinishRegistrationResponse struct {
	Success      bool   `json:"success"`
	CredentialID string `json:"credential_id,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	AAGUID       string `json:"aaguid,omitempty"`
	Error        string `json:"error,omitempty"`
}

// BeginAuthenticationRequest request to begin passkey authentication
type BeginAuthenticationRequest struct {
	TenantID  string `json:"tenant_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	Username  string `json:"username,omitempty"` // Optional for usernameless flow
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// BeginAuthenticationResponse response with WebAuthn request options
type BeginAuthenticationResponse struct {
	Success bool `json:"success"`

	// WebAuthn PublicKeyCredentialRequestOptions
	Challenge        string              `json:"challenge"`
	Timeout          int                 `json:"timeout"`
	RelyingPartyID   string              `json:"rpId"`
	AllowCredentials []AllowedCredential `json:"allowCredentials,omitempty"`
	UserVerification string              `json:"userVerification"` // "required", "preferred", "discouraged"

	Error string `json:"error,omitempty"`
}

// FinishAuthenticationRequest request to complete passkey authentication
type FinishAuthenticationRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	AppID    string `json:"app_id" validate:"required"`

	// WebAuthn response from navigator.credentials.get()
	CredentialID      string `json:"credential_id" validate:"required"`
	ClientDataJSON    string `json:"client_data_json" validate:"required"`
	AuthenticatorData string `json:"authenticator_data" validate:"required"`
	Signature         string `json:"signature" validate:"required"`
	UserHandle        string `json:"user_handle,omitempty"`

	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// FinishAuthenticationResponse response after completing authentication
type FinishAuthenticationResponse struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

// =============================================================================
// WebAuthn Types
// =============================================================================

// RelyingPartyInfo relying party information
type RelyingPartyInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// UserInfo user information for WebAuthn
type UserInfo struct {
	ID          string `json:"id"`          // Base64url encoded user handle
	Name        string `json:"name"`        // Username
	DisplayName string `json:"displayName"` // Display name
}

// PubKeyCredParam supported public key credential parameters
type PubKeyCredParam struct {
	Type string `json:"type"` // "public-key"
	Alg  int    `json:"alg"`  // COSE algorithm identifier (-7 for ES256, -257 for RS256)
}

// AuthenticatorSelectionCriteria authenticator selection criteria
type AuthenticatorSelectionCriteria struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"` // "platform", "cross-platform"
	RequireResidentKey      bool   `json:"requireResidentKey"`
	ResidentKey             string `json:"residentKey"`      // "discouraged", "preferred", "required"
	UserVerification        string `json:"userVerification"` // "required", "preferred", "discouraged"
}

// AllowedCredential allowed credential for authentication
type AllowedCredential struct {
	Type       string   `json:"type"` // "public-key"
	ID         string   `json:"id"`   // Base64url encoded credential ID
	Transports []string `json:"transports,omitempty"`
}
