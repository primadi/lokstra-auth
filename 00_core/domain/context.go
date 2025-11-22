package domain

// Context represents the authentication and authorization context
// Every operation in lokstra-auth MUST include this context
//
// Organizational Hierarchy:
//   Tenant (Organization) → App (Application) → Branch (Location/Store/Office)
//
// Examples:
//   - Multi-store retail: Each store is a branch
//   - Multi-branch bank: Each office is a branch
//   - Multi-location company: Each location is a branch
//   - Multi-franchise: Each franchise outlet is a branch
type Context struct {
	TenantID  string         `json:"tenant_id"`           // REQUIRED: Tenant context (organization level)
	AppID     string         `json:"app_id"`              // REQUIRED: App context (application level)
	BranchID  string         `json:"branch_id,omitempty"` // OPTIONAL: Branch context (location/store/office level)
	UserID    string         `json:"user_id,omitempty"`   // OPTIONAL: User context
	SessionID string         `json:"session_id,omitempty"`
	IPAddress string         `json:"ip_address,omitempty"`
	UserAgent string         `json:"user_agent,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// NewContext creates a new context with tenant and app
func NewContext(tenantID, appID string) *Context {
	return &Context{
		TenantID: tenantID,
		AppID:    appID,
		Metadata: make(map[string]any),
	}
}

// NewContextWithBranch creates a new context with tenant, app, and branch
func NewContextWithBranch(tenantID, appID, branchID string) *Context {
	return &Context{
		TenantID: tenantID,
		AppID:    appID,
		BranchID: branchID,
		Metadata: make(map[string]any),
	}
}

// WithBranch adds branch context
func (c *Context) WithBranch(branchID string) *Context {
	c.BranchID = branchID
	return c
}

// WithUser adds user context
func (c *Context) WithUser(userID string) *Context {
	c.UserID = userID
	return c
}

// WithSession adds session context
func (c *Context) WithSession(sessionID string) *Context {
	c.SessionID = sessionID
	return c
}

// WithIP adds IP address
func (c *Context) WithIP(ipAddress string) *Context {
	c.IPAddress = ipAddress
	return c
}

// WithUserAgent adds user agent
func (c *Context) WithUserAgent(userAgent string) *Context {
	c.UserAgent = userAgent
	return c
}

// WithMetadata adds custom metadata
func (c *Context) WithMetadata(key string, value any) *Context {
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	c.Metadata[key] = value
	return c
}

// Validate validates the context
func (c *Context) Validate() error {
	if c.TenantID == "" {
		return ErrInvalidTenantID
	}
	if c.AppID == "" {
		return ErrInvalidAppID
	}
	return nil
}
