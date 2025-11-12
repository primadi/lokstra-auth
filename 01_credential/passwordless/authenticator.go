package passwordless

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	credential "github.com/primadi/lokstra-auth/01_credential"
)

var (
	ErrInvalidToken  = errors.New("invalid passwordless token")
	ErrTokenExpired  = errors.New("passwordless token expired")
	ErrTokenNotFound = errors.New("token not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidEmail  = errors.New("invalid email address")
)

// TokenType represents the type of passwordless token
type TokenType string

const (
	TokenTypeMagicLink TokenType = "magic_link"
	TokenTypeOTP       TokenType = "otp"
)

// Credentials represents passwordless credentials
type Credentials struct {
	Email     string
	Token     string
	TokenType TokenType
}

func (c *Credentials) Type() string {
	return "passwordless"
}

func (c *Credentials) Validate() error {
	if c.Email == "" {
		return ErrInvalidEmail
	}
	if c.Token == "" {
		return errors.New("token is required")
	}
	return nil
}

// TokenData represents stored token information
type TokenData struct {
	Token     string
	Email     string
	UserID    string
	Type      TokenType
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// TokenStore manages passwordless tokens
type TokenStore interface {
	// Store saves a token
	Store(ctx context.Context, token *TokenData) error

	// Get retrieves a token
	Get(ctx context.Context, token string) (*TokenData, error)

	// MarkUsed marks a token as used
	MarkUsed(ctx context.Context, token string) error

	// Delete removes a token
	Delete(ctx context.Context, token string) error

	// Cleanup removes expired tokens
	Cleanup(ctx context.Context) error
}

// UserResolver resolves user information by email
type UserResolver interface {
	// ResolveByEmail finds a user by email and returns user ID and claims
	ResolveByEmail(ctx context.Context, email string) (userID string, claims map[string]interface{}, err error)
}

// TokenGenerator generates passwordless tokens
type TokenGenerator interface {
	// GenerateMagicLink generates a magic link token
	GenerateMagicLink() (string, error)

	// GenerateOTP generates a one-time password
	GenerateOTP() (string, error)
}

// TokenSender sends tokens to users
type TokenSender interface {
	// SendMagicLink sends a magic link email
	SendMagicLink(ctx context.Context, email, token, link string) error

	// SendOTP sends an OTP code
	SendOTP(ctx context.Context, email, code string) error
}

// Authenticator handles passwordless authentication
type Authenticator struct {
	tokenStore    TokenStore
	userResolver  UserResolver
	tokenGen      TokenGenerator
	tokenSender   TokenSender
	otpExpiry     time.Duration
	magicExpiry   time.Duration
	allowedEmails map[string]bool // Optional: whitelist of allowed emails
}

// Config holds configuration for passwordless authenticator
type Config struct {
	TokenStore     TokenStore
	UserResolver   UserResolver
	TokenGenerator TokenGenerator
	TokenSender    TokenSender

	// OTPExpiry is the duration for OTP validity (default: 5 minutes)
	OTPExpiry time.Duration

	// MagicLinkExpiry is the duration for magic link validity (default: 15 minutes)
	MagicLinkExpiry time.Duration

	// AllowedEmails is an optional whitelist of allowed email addresses
	AllowedEmails []string
}

// DefaultConfig returns default passwordless configuration
func DefaultConfig() *Config {
	return &Config{
		OTPExpiry:       5 * time.Minute,
		MagicLinkExpiry: 15 * time.Minute,
	}
}

// NewAuthenticator creates a new passwordless authenticator
func NewAuthenticator(config *Config) *Authenticator {
	if config == nil {
		config = DefaultConfig()
	}

	if config.OTPExpiry == 0 {
		config.OTPExpiry = 5 * time.Minute
	}

	if config.MagicLinkExpiry == 0 {
		config.MagicLinkExpiry = 15 * time.Minute
	}

	// Use in-memory store if not provided
	if config.TokenStore == nil {
		config.TokenStore = NewInMemoryTokenStore()
	}

	// Use default token generator if not provided
	if config.TokenGenerator == nil {
		config.TokenGenerator = NewDefaultTokenGenerator()
	}

	auth := &Authenticator{
		tokenStore:   config.TokenStore,
		userResolver: config.UserResolver,
		tokenGen:     config.TokenGenerator,
		tokenSender:  config.TokenSender,
		otpExpiry:    config.OTPExpiry,
		magicExpiry:  config.MagicLinkExpiry,
	}

	// Build allowed emails map
	if len(config.AllowedEmails) > 0 {
		auth.allowedEmails = make(map[string]bool)
		for _, email := range config.AllowedEmails {
			auth.allowedEmails[email] = true
		}
	}

	return auth
}

// Authenticate verifies passwordless credentials
func (a *Authenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	pwdlessCreds, ok := creds.(*Credentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   errors.New("invalid credentials type"),
		}, nil
	}

	if err := pwdlessCreds.Validate(); err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Check if email is allowed
	if a.allowedEmails != nil && !a.allowedEmails[pwdlessCreds.Email] {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   errors.New("email not allowed"),
		}, nil
	}

	// Get token from store
	tokenData, err := a.tokenStore.Get(ctx, pwdlessCreds.Token)
	if err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrTokenNotFound,
		}, nil
	}

	// Verify token matches email
	if tokenData.Email != pwdlessCreds.Email {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrInvalidToken,
		}, nil
	}

	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   ErrTokenExpired,
		}, nil
	}

	// Check if token was already used
	if tokenData.Used {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   errors.New("token already used"),
		}, nil
	}

	// Mark token as used
	if err := a.tokenStore.MarkUsed(ctx, pwdlessCreds.Token); err != nil {
		return nil, err
	}

	// Resolve user information
	var userID string
	var claims map[string]interface{}

	if a.userResolver != nil {
		userID, claims, err = a.userResolver.ResolveByEmail(ctx, pwdlessCreds.Email)
		if err != nil {
			return &credential.AuthenticationResult{
				Success: false,
				Error:   err,
			}, nil
		}
	} else {
		// Use token data if no resolver provided
		userID = tokenData.UserID
		claims = map[string]interface{}{
			"sub":    userID,
			"email":  pwdlessCreds.Email,
			"method": string(tokenData.Type),
		}
	}

	// Ensure claims have required fields
	if claims == nil {
		claims = make(map[string]interface{})
	}
	claims["sub"] = userID
	claims["email"] = pwdlessCreds.Email
	claims["auth_method"] = string(tokenData.Type)

	return &credential.AuthenticationResult{
		Success: true,
		Subject: userID,
		Claims:  claims,
	}, nil
}

// Type returns the authenticator type
func (a *Authenticator) Type() string {
	return "passwordless"
}

// InitiateMagicLink creates and sends a magic link token
func (a *Authenticator) InitiateMagicLink(ctx context.Context, email, userID, baseURL string) error {
	// Generate token
	token, err := a.tokenGen.GenerateMagicLink()
	if err != nil {
		return err
	}

	// Store token
	tokenData := &TokenData{
		Token:     token,
		Email:     email,
		UserID:    userID,
		Type:      TokenTypeMagicLink,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(a.magicExpiry),
		Used:      false,
	}

	if err := a.tokenStore.Store(ctx, tokenData); err != nil {
		return err
	}

	// Send email
	if a.tokenSender != nil {
		link := fmt.Sprintf("%s/auth/verify?token=%s&email=%s", baseURL, token, email)
		return a.tokenSender.SendMagicLink(ctx, email, token, link)
	}

	return nil
}

// InitiateOTP creates and sends an OTP code
func (a *Authenticator) InitiateOTP(ctx context.Context, email, userID string) error {
	// Generate OTP
	code, err := a.tokenGen.GenerateOTP()
	if err != nil {
		return err
	}

	// Store token
	tokenData := &TokenData{
		Token:     code,
		Email:     email,
		UserID:    userID,
		Type:      TokenTypeOTP,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(a.otpExpiry),
		Used:      false,
	}

	if err := a.tokenStore.Store(ctx, tokenData); err != nil {
		return err
	}

	// Send OTP
	if a.tokenSender != nil {
		return a.tokenSender.SendOTP(ctx, email, code)
	}

	return nil
}

// InMemoryTokenStore is an in-memory implementation of TokenStore
type InMemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*TokenData
}

// NewInMemoryTokenStore creates a new in-memory token store
func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens: make(map[string]*TokenData),
	}
}

func (s *InMemoryTokenStore) Store(ctx context.Context, token *TokenData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Token] = token
	return nil
}

func (s *InMemoryTokenStore) Get(ctx context.Context, token string) (*TokenData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.tokens[token]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return data, nil
}

func (s *InMemoryTokenStore) MarkUsed(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.tokens[token]
	if !ok {
		return ErrTokenNotFound
	}

	data.Used = true
	return nil
}

func (s *InMemoryTokenStore) Delete(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
	return nil
}

func (s *InMemoryTokenStore) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, data := range s.tokens {
		if now.After(data.ExpiresAt) || data.Used {
			delete(s.tokens, token)
		}
	}

	return nil
}

// DefaultTokenGenerator implements TokenGenerator
type DefaultTokenGenerator struct{}

func NewDefaultTokenGenerator() *DefaultTokenGenerator {
	return &DefaultTokenGenerator{}
}

func (g *DefaultTokenGenerator) GenerateMagicLink() (string, error) {
	// Generate 32 random bytes and encode as base64
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (g *DefaultTokenGenerator) GenerateOTP() (string, error) {
	// Generate 6-digit OTP
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to 6 digit number
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	otp := fmt.Sprintf("%06d", num%1000000)
	return otp, nil
}
