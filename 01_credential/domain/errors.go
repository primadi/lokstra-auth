package domain

import "errors"

var (
	// Auth context errors
	ErrMissingTenantID = errors.New("tenant_id is required in auth context")
	ErrMissingAppID    = errors.New("app_id is required in auth context")

	// Common authentication errors
	ErrInvalidCredentials   = errors.New("invalid credentials type")
	ErrAuthenticationFailed = errors.New("authentication failed")
)
