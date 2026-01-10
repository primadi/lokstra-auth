package validator

import (
	"errors"
	"regexp"
	"unicode"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

var (
	ErrConfigNotProvided = errors.New("credential config not provided")
)

// ConfigurableValidator validates credentials based on CredentialConfig
// This validator reads validation rules from domain.CredentialConfig
// @Service "credential-validator"
type ConfigurableValidator struct{}

// NewConfigurableValidator creates a new configurable validator

var _ CredentialValidator = (*ConfigurableValidator)(nil)

// ValidateUsername checks if username meets requirements from config
func (v *ConfigurableValidator) ValidateUsername(username string) error {
	// Use default config since we don't have context here
	// This is a fallback - actual validation should use ValidateUsernameWithConfig
	config := domaincore.DefaultCredentialConfig().BasicConfig
	return v.ValidateUsernameWithConfig(username, config)
}

// ValidatePassword checks if password meets requirements from config
func (v *ConfigurableValidator) ValidatePassword(password string) error {
	// Use default config since we don't have context here
	// This is a fallback - actual validation should use ValidatePasswordWithConfig
	config := domaincore.DefaultCredentialConfig().BasicConfig
	return v.ValidatePasswordWithConfig(password, config)
}

// ValidateUsernameWithConfig validates username against specific config
func (v *ConfigurableValidator) ValidateUsernameWithConfig(username string, config *domaincore.BasicCredentialConfig) error {
	if config == nil {
		return ErrConfigNotProvided
	}

	if len(username) < config.MinUsernameLength {
		return ErrUsernameTooShort
	}
	if len(username) > config.MaxUsernameLength {
		return ErrUsernameTooLong
	}

	// Only allow alphanumeric and underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return ErrUsernameInvalidPattern
	}

	return nil
}

// ValidatePasswordWithConfig validates password against specific config
func (v *ConfigurableValidator) ValidatePasswordWithConfig(password string, config *domaincore.BasicCredentialConfig) error {
	if config == nil {
		return ErrConfigNotProvided
	}

	if len(password) < config.MinPasswordLength {
		return ErrPasswordTooShort
	}

	if !config.RequireStrongPwd {
		return nil
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordTooWeak
	}

	return nil
}
