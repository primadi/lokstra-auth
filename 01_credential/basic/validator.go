package basic

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	credential "github.com/primadi/lokstra-auth/01_credential"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrEmptyUsername         = errors.New("username cannot be empty")
	ErrEmptyPassword         = errors.New("password cannot be empty")
	ErrUsernameTooShort      = errors.New("username is too short")
	ErrPasswordTooShort      = errors.New("password is too short")
	ErrPasswordTooWeak       = errors.New("password does not meet complexity requirements")
	ErrInvalidUsernameFormat = errors.New("username contains invalid characters")
)

// BasicCredentials represents username/password credentials
type BasicCredentials struct {
	Username string
	Password string
}

// Type returns the credential type
func (c *BasicCredentials) Type() string {
	return "basic"
}

// Validate checks if the credentials are well-formed
func (c *BasicCredentials) Validate() error {
	if strings.TrimSpace(c.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(c.Password) == "" {
		return ErrEmptyPassword
	}
	return nil
}

// ValidatorConfig holds configuration for basic credential validation
type ValidatorConfig struct {
	MinUsernameLength int
	MinPasswordLength int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireDigit      bool
	RequireSpecial    bool
	UsernamePattern   *regexp.Regexp
}

// DefaultValidatorConfig returns a default configuration
func DefaultValidatorConfig() *ValidatorConfig {
	return &ValidatorConfig{
		MinUsernameLength: 3,
		MinPasswordLength: 8,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    false,
		UsernamePattern:   regexp.MustCompile(`^[a-zA-Z0-9_\-\.@]+$`),
	}
}

// Validator validates basic credentials
type Validator struct {
	config *ValidatorConfig
}

// NewValidator creates a new basic credentials validator
func NewValidator(config *ValidatorConfig) *Validator {
	if config == nil {
		config = DefaultValidatorConfig()
	}
	return &Validator{
		config: config,
	}
}

// Validate checks if the credentials meet the required format and rules
func (v *Validator) Validate(ctx context.Context, creds credential.Credentials) error {
	basicCreds, ok := creds.(*BasicCredentials)
	if !ok {
		return fmt.Errorf("expected BasicCredentials, got %T", creds)
	}

	// Basic validation
	if err := basicCreds.Validate(); err != nil {
		return err
	}

	// Username validation
	if len(basicCreds.Username) < v.config.MinUsernameLength {
		return ErrUsernameTooShort
	}

	if v.config.UsernamePattern != nil && !v.config.UsernamePattern.MatchString(basicCreds.Username) {
		return ErrInvalidUsernameFormat
	}

	// Password validation
	if len(basicCreds.Password) < v.config.MinPasswordLength {
		return ErrPasswordTooShort
	}

	if err := v.validatePasswordComplexity(basicCreds.Password); err != nil {
		return err
	}

	return nil
}

// Type returns the type of credentials this validator handles
func (v *Validator) Type() string {
	return "basic"
}

func (v *Validator) validatePasswordComplexity(password string) error {
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if v.config.RequireUppercase && !hasUpper {
		return ErrPasswordTooWeak
	}
	if v.config.RequireLowercase && !hasLower {
		return ErrPasswordTooWeak
	}
	if v.config.RequireDigit && !hasDigit {
		return ErrPasswordTooWeak
	}
	if v.config.RequireSpecial && !hasSpecial {
		return ErrPasswordTooWeak
	}

	return nil
}
