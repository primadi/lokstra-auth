package validator

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
)

// CredentialValidator validates credentials format and complexity requirements.
// This is used during user registration/creation to ensure credentials meet
// security requirements before they are stored.
type CredentialValidator interface {
	// ValidateUsername checks if username meets requirements
	ValidateUsername(username string) error

	// ValidatePassword checks if password meets complexity requirements
	ValidatePassword(password string) error
}

// ConfigurableCredentialValidator extends CredentialValidator with config-aware validation
type ConfigurableCredentialValidator interface {
	CredentialValidator

	// ValidateUsernameWithConfig validates username against specific config
	ValidateUsernameWithConfig(username string, config *domaincore.BasicCredentialConfig) error

	// ValidatePasswordWithConfig validates password against specific config
	ValidatePasswordWithConfig(password string, config *domaincore.BasicCredentialConfig) error
}
