package validator

import (
	"errors"
	"regexp"
	"unicode"
)

var (
	ErrUsernameTooShort       = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong        = errors.New("username must be at most 32 characters")
	ErrUsernameInvalidPattern = errors.New("username must contain only letters, numbers, and underscores")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordTooWeak        = errors.New("password must contain at least one uppercase, one lowercase, and one number")
)

// SimpleValidator provides basic validation for credentials
type SimpleValidator struct {
	minUsernameLength int
	maxUsernameLength int
	minPasswordLength int
	requireStrongPwd  bool
}

var _ CredentialValidator = (*SimpleValidator)(nil)

// NewSimpleValidator creates a new simple validator with default settings
func NewSimpleValidator() *SimpleValidator {
	return &SimpleValidator{
		minUsernameLength: 3,
		maxUsernameLength: 32,
		minPasswordLength: 8,
		requireStrongPwd:  true,
	}
}

// ValidateUsername checks if username meets requirements
func (v *SimpleValidator) ValidateUsername(username string) error {
	if len(username) < v.minUsernameLength {
		return ErrUsernameTooShort
	}
	if len(username) > v.maxUsernameLength {
		return ErrUsernameTooLong
	}

	// Only allow alphanumeric and underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return ErrUsernameInvalidPattern
	}

	return nil
}

// ValidatePassword checks if password meets complexity requirements
func (v *SimpleValidator) ValidatePassword(password string) error {
	if len(password) < v.minPasswordLength {
		return ErrPasswordTooShort
	}

	if !v.requireStrongPwd {
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
