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
