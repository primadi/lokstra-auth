package domain

// AuthContext provides the authentication context for multi-tenant operations.
// Every authentication MUST include tenant_id and app_id for proper isolation.
// BranchID is optional and used for branch-scoped authentication.
type AuthContext struct {
	TenantID  string // REQUIRED: Tenant identifier
	AppID     string // REQUIRED: Application identifier
	BranchID  string // Optional: Branch identifier for branch-scoped auth
	IPAddress string // Optional: Client IP address for audit
	UserAgent string // Optional: Client user agent for audit
	SessionID string // Optional: Session identifier
}

// Validate checks if the auth context has required fields
func (ac *AuthContext) Validate() error {
	if ac.TenantID == "" {
		return ErrMissingTenantID
	}
	if ac.AppID == "" {
		return ErrMissingAppID
	}
	return nil
}

// AuthenticationResult represents the result of an authentication attempt
type AuthenticationResult struct {
	Success  bool   // Whether authentication was successful
	Subject  string // Authenticated user identifier (user ID)
	TenantID string // Tenant context (copied from AuthContext)
	AppID    string // App context (copied from AuthContext)
	BranchID string // Branch context (copied from AuthContext, optional)

	// Claims contains additional information to be included in tokens
	// Common claims: roles, permissions, scopes, email, name, etc.
	Claims map[string]any

	// Error contains the error if authentication failed
	Error error
}
