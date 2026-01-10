package core

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
