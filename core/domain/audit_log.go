package domain

import (
	"errors"
	"time"
)

var (
	ErrAuditLogNotFound = errors.New("audit log not found")
)

// AuditLog represents a centralized audit log entry
type AuditLog struct {
	ID           int64           `json:"id"`
	TenantID     *string         `json:"tenant_id,omitempty"`
	AppID        *string         `json:"app_id,omitempty"`
	UserID       *string         `json:"user_id,omitempty"`
	SessionID    *string         `json:"session_id,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	Method       *string         `json:"method,omitempty"`
	Path         *string         `json:"path,omitempty"`
	StatusCode   *int            `json:"status_code,omitempty"`
	RequestBody  *map[string]any `json:"request_body,omitempty"`
	ResponseBody *map[string]any `json:"response_body,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	Source       *string         `json:"source,omitempty"`
	Success      bool            `json:"success"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	Metadata     *map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// CreateAuditLogRequest represents a request to create an audit log
type CreateAuditLogRequest struct {
	TenantID     *string         `json:"tenant_id,omitempty"`
	AppID        *string         `json:"app_id,omitempty"`
	UserID       *string         `json:"user_id,omitempty"`
	SessionID    *string         `json:"session_id,omitempty"`
	Action       string          `json:"action" validate:"required"`
	ResourceType string          `json:"resource_type" validate:"required"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	Method       *string         `json:"method,omitempty"`
	Path         *string         `json:"path,omitempty"`
	StatusCode   *int            `json:"status_code,omitempty"`
	RequestBody  *map[string]any `json:"request_body,omitempty"`
	ResponseBody *map[string]any `json:"response_body,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	Source       *string         `json:"source,omitempty"`
	Success      bool            `json:"success"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	Metadata     *map[string]any `json:"metadata,omitempty"`
}

// ListAuditLogsRequest represents a request to list audit logs
type ListAuditLogsRequest struct {
	TenantID     *string `json:"tenant_id,omitempty"`
	AppID        *string `json:"app_id,omitempty"`
	UserID       *string `json:"user_id,omitempty"`
	Action       *string `json:"action,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Source       *string `json:"source,omitempty"`
	Success      *bool   `json:"success,omitempty"`
	FromDate     *string `json:"from_date,omitempty"` // ISO8601 format
	ToDate       *string `json:"to_date,omitempty"`   // ISO8601 format
	Limit        int     `json:"limit" validate:"min=1,max=1000"`
	Offset       int     `json:"offset" validate:"min=0"`
}

// AuditAction constants for common actions
const (
	// Authentication actions
	ActionLogin          = "login"
	ActionLogout         = "logout"
	ActionLoginFailed    = "login_failed"
	ActionRegister       = "register"
	ActionPasswordChange = "password_change"
	ActionPasswordReset  = "password_reset"
	ActionTokenRefresh   = "token_refresh"
	ActionTokenRevoke    = "token_revoke"

	// Authorization actions
	ActionAuthzCheck = "authz_check"
	ActionAuthzDeny  = "authz_deny"

	// CRUD actions
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"

	// Account management
	ActionAccountLock     = "account_lock"
	ActionAccountUnlock   = "account_unlock"
	ActionAccountSuspend  = "account_suspend"
	ActionAccountActivate = "account_activate"

	// RBAC actions
	ActionRoleAssign       = "role_assign"
	ActionRoleRevoke       = "role_revoke"
	ActionPermissionGrant  = "permission_grant"
	ActionPermissionRevoke = "permission_revoke"

	// Ownership actions
	ActionOwnershipTransfer = "ownership_transfer"

	// Configuration actions
	ActionConfigUpdate = "config_update"
)

// ResourceType constants for common resources
const (
	ResourceTenant             = "tenant"
	ResourceApp                = "app"
	ResourceBranch             = "branch"
	ResourceUser               = "user"
	ResourceUserIdentity       = "user_identity"
	ResourceAppKey             = "app_key"
	ResourceCredentialProvider = "credential_provider"
	ResourceRole               = "role"
	ResourcePermission         = "permission"
	ResourcePolicy             = "policy"
	ResourceToken              = "token"
	ResourceSession            = "session"
)

// Source constants
const (
	SourceAPI    = "api"
	SourceWeb    = "web"
	SourceMobile = "mobile"
	SourceSystem = "system"
	SourceCron   = "cron"
	SourceCLI    = "cli"
)

// Validate validates the audit log creation request
func (r *CreateAuditLogRequest) Validate() error {
	if r.Action == "" {
		return errors.New("action is required")
	}
	if r.ResourceType == "" {
		return errors.New("resource_type is required")
	}
	return nil
}
