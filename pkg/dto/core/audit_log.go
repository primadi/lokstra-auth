package core

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
