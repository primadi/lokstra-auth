package passkey

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

var (
	// ErrInvalidCredential indicates invalid passkey credential
	ErrInvalidCredential = errors.New("invalid passkey credential")

	// ErrCredentialNotFound indicates credential not found in store
	ErrCredentialNotFound = errors.New("credential not found")

	// ErrRegistrationFailed indicates passkey registration failed
	ErrRegistrationFailed = errors.New("passkey registration failed")

	// ErrAuthenticationFailed indicates passkey authentication failed
	ErrAuthenticationFailed = errors.New("passkey authentication failed")
)

// Config holds configuration for Passkey authenticator
type Config struct {
	// RPDisplayName is the relying party display name (e.g., "My App")
	RPDisplayName string

	// RPID is the relying party ID (e.g., "example.com")
	RPID string

	// RPOrigins are allowed origins for WebAuthn
	RPOrigins []string

	// Timeout for authentication ceremony (default: 60 seconds)
	Timeout time.Duration

	// RequireResidentKey requires passkey to be stored on authenticator
	RequireResidentKey bool

	// UserVerification level: "required", "preferred", "discouraged"
	UserVerification string

	// CredentialStore for storing passkey credentials
	CredentialStore CredentialStore
}

// DefaultConfig returns default passkey configuration
func DefaultConfig(rpID, rpDisplayName string) *Config {
	return &Config{
		RPDisplayName:      rpDisplayName,
		RPID:               rpID,
		RPOrigins:          []string{fmt.Sprintf("https://%s", rpID)},
		Timeout:            60 * time.Second,
		RequireResidentKey: false,
		UserVerification:   "preferred",
		CredentialStore:    NewInMemoryCredentialStore(),
	}
}

// Authenticator implements passkey (WebAuthn) authentication
type Authenticator struct {
	config   *Config
	webAuthn *webauthn.WebAuthn
	sessions map[string]*webauthn.SessionData // challenge -> session
}

// NewAuthenticator creates a new passkey authenticator
func NewAuthenticator(config *Config) (*Authenticator, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.CredentialStore == nil {
		config.CredentialStore = NewInMemoryCredentialStore()
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	// Create WebAuthn instance
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     config.RPOrigins,
	}

	web, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn: %w", err)
	}

	return &Authenticator{
		config:   config,
		webAuthn: web,
		sessions: make(map[string]*webauthn.SessionData),
	}, nil
}

// BeginRegistration starts passkey registration ceremony
func (a *Authenticator) BeginRegistration(ctx context.Context, user *User) (*RegistrationOptions, error) {
	// Create registration options
	options, session, err := a.webAuthn.BeginRegistration(user)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	// Store session
	challenge := base64.StdEncoding.EncodeToString(options.Response.Challenge)
	a.sessions[challenge] = session

	return &RegistrationOptions{
		Challenge:              challenge,
		RelyingParty:           options.Response.RelyingParty,
		User:                   options.Response.User,
		PubKeyCredParams:       options.Response.Parameters,
		Timeout:                options.Response.Timeout,
		AuthenticatorSelection: options.Response.AuthenticatorSelection,
		Attestation:            string(options.Response.Attestation),
	}, nil
}

// FinishRegistration completes passkey registration
func (a *Authenticator) FinishRegistration(ctx context.Context, userID string, response string) error {
	// Get user
	user, err := a.config.CredentialStore.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	// Parse response
	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader([]byte(response)))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	// Get session (try all sessions since we don't know which challenge)
	var session *webauthn.SessionData
	for _, s := range a.sessions {
		session = s
		break
	}

	if session == nil {
		return fmt.Errorf("%w: no active session", ErrRegistrationFailed)
	}

	// Verify registration
	credential, err := a.webAuthn.CreateCredential(user, *session, parsedResponse)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	// Store credential
	err = a.config.CredentialStore.StoreCredential(ctx, userID, credential)
	if err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	// Clean up sessions
	for challenge := range a.sessions {
		delete(a.sessions, challenge)
	}

	return nil
}

// BeginLogin starts passkey authentication ceremony
func (a *Authenticator) BeginLogin(ctx context.Context, userID string) (*LoginOptions, error) {
	// Get user
	user, err := a.config.CredentialStore.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Create login options
	options, session, err := a.webAuthn.BeginLogin(user)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Store session
	challenge := base64.StdEncoding.EncodeToString(options.Response.Challenge)
	a.sessions[challenge] = session

	return &LoginOptions{
		Challenge:        challenge,
		Timeout:          options.Response.Timeout,
		RPID:             options.Response.RelyingPartyID,
		AllowCredentials: options.Response.AllowedCredentials,
		UserVerification: string(options.Response.UserVerification),
	}, nil
}

// FinishLogin completes passkey authentication
func (a *Authenticator) FinishLogin(ctx context.Context, userID string, response string) (*LoginResult, error) {
	// Get user
	user, err := a.config.CredentialStore.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Parse response
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader([]byte(response)))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Get session (try all sessions since we don't know which challenge)
	var session *webauthn.SessionData
	for _, s := range a.sessions {
		session = s
		break
	}

	if session == nil {
		return nil, fmt.Errorf("%w: no active session", ErrAuthenticationFailed)
	}

	// Verify authentication
	credential, err := a.webAuthn.ValidateLogin(user, *session, parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Update credential usage
	err = a.config.CredentialStore.UpdateCredential(ctx, userID, credential)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Clean up sessions
	for challenge := range a.sessions {
		delete(a.sessions, challenge)
	}

	return &LoginResult{
		UserID:       userID,
		CredentialID: base64.StdEncoding.EncodeToString(credential.ID),
		AAGUID:       base64.StdEncoding.EncodeToString(credential.Authenticator.AAGUID),
		SignCount:    credential.Authenticator.SignCount,
	}, nil
}

// LoginResult contains the result of passkey authentication
type LoginResult struct {
	UserID       string `json:"user_id"`
	CredentialID string `json:"credential_id"`
	AAGUID       string `json:"aaguid"`
	SignCount    uint32 `json:"sign_count"`
}

// RegistrationOptions contains options for passkey registration
type RegistrationOptions struct {
	Challenge              string                          `json:"challenge"`
	RelyingParty           protocol.RelyingPartyEntity     `json:"rp"`
	User                   protocol.UserEntity             `json:"user"`
	PubKeyCredParams       []protocol.CredentialParameter  `json:"pubKeyCredParams"`
	Timeout                int                             `json:"timeout"`
	AuthenticatorSelection protocol.AuthenticatorSelection `json:"authenticatorSelection"`
	Attestation            string                          `json:"attestation"`
}

// LoginOptions contains options for passkey authentication
type LoginOptions struct {
	Challenge        string                          `json:"challenge"`
	Timeout          int                             `json:"timeout"`
	RPID             string                          `json:"rpId"`
	AllowCredentials []protocol.CredentialDescriptor `json:"allowCredentials"`
	UserVerification string                          `json:"userVerification"`
}

// User implements webauthn.User interface
type User struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

// WebAuthnID returns user ID
func (u *User) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns username
func (u *User) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns display name
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials returns user credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// WebAuthnIcon returns user icon URL (deprecated)
func (u *User) WebAuthnIcon() string {
	return ""
}

// GenerateUserID generates a random user ID
func GenerateUserID() ([]byte, error) {
	id := make([]byte, 32)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}
	return id, nil
}
