package middleware

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/request"
)

// AuditLogger is the interface for audit logging implementations
type AuditLogger interface {
	// Log writes an audit entry
	Log(entry *AuditEntry) error
}

// AuditEntry represents a logged operation
type AuditEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	TenantID   string         `json:"tenant_id"`
	AppID      string         `json:"app_id"`
	UserID     string         `json:"user_id"`
	Username   string         `json:"username"`
	Action     string         `json:"action"` // HTTP method + path
	Resource   string         `json:"resource"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	StatusCode int            `json:"status_code"`
	Duration   time.Duration  `json:"duration"`
	Error      string         `json:"error,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// AuditMiddleware logs all operations for compliance and security monitoring
type AuditMiddleware struct {
	logger AuditLogger
	skip   func(c *request.Context) bool
}

// AuditConfig holds configuration for audit middleware
type AuditConfig struct {
	// Logger is the audit log implementation
	Logger AuditLogger

	// Skip allows filtering which requests to audit
	// Return true to skip auditing for this request
	Skip func(c *request.Context) bool
}

// NewAuditMiddleware creates a new audit logging middleware
func NewAuditMiddleware(config AuditConfig) *AuditMiddleware {
	return &AuditMiddleware{
		logger: config.Logger,
		skip:   config.Skip,
	}
}

// Handler returns the middleware handler function
func (m *AuditMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Skip if configured
		if m.skip != nil && m.skip(c) {
			return c.Next()
		}

		start := time.Now()

		// Get identity (if available from AuthMiddleware)
		var tenantID, appID, userID, username string
		if identity, ok := GetIdentity(c); ok {
			tenantID = identity.TenantID
			appID = identity.AppID
			userID = identity.Subject.ID
			username = identity.Subject.Principal
		}

		// If no identity, try to get tenant from header
		if tenantID == "" {
			tenantID = c.R.Header.Get("X-Tenant-ID")
		}
		if appID == "" {
			appID = c.R.Header.Get("X-App-ID")
		}

		// Get IP address
		ipAddress := c.R.RemoteAddr
		if forwarded := c.R.Header.Get("X-Forwarded-For"); forwarded != "" {
			ipAddress = forwarded
		} else if realIP := c.R.Header.Get("X-Real-IP"); realIP != "" {
			ipAddress = realIP
		}

		// Continue to handler
		handlerErr := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		statusCode := c.W.StatusCode()
		if statusCode == 0 {
			statusCode = 200 // Default if not set
		}

		// Build audit entry
		entry := &AuditEntry{
			Timestamp:  start,
			TenantID:   tenantID,
			AppID:      appID,
			UserID:     userID,
			Username:   username,
			Action:     fmt.Sprintf("%s %s", c.R.Method, c.R.URL.Path),
			Resource:   c.R.URL.Path,
			IPAddress:  ipAddress,
			UserAgent:  c.R.UserAgent(),
			StatusCode: statusCode,
			Duration:   duration,
		}

		// Add error if any
		if handlerErr != nil {
			entry.Error = handlerErr.Error()
		}

		// Add query parameters as metadata
		if len(c.R.URL.Query()) > 0 {
			entry.Metadata = make(map[string]any)
			for k, v := range c.R.URL.Query() {
				if len(v) == 1 {
					entry.Metadata[k] = v[0]
				} else {
					entry.Metadata[k] = v
				}
			}
		}

		// Log audit entry (non-blocking, don't fail request if logging fails)
		go func() {
			if err := m.logger.Log(entry); err != nil {
				// In production, you might want to log this error somewhere
				// fmt.Printf("audit logging failed: %v\n", err)
			}
		}()

		return handlerErr
	}
}

// ConsoleAuditLogger logs audit entries to console (for development)
type ConsoleAuditLogger struct{}

// NewConsoleAuditLogger creates a new console audit logger
func NewConsoleAuditLogger() *ConsoleAuditLogger {
	return &ConsoleAuditLogger{}
}

func (l *ConsoleAuditLogger) Log(entry *AuditEntry) error {
	emoji := "✅"
	if entry.StatusCode >= 400 {
		emoji = "❌"
	} else if entry.StatusCode >= 300 {
		emoji = "🔄"
	}

	fmt.Printf("[AUDIT] %s %s | %s | %s@%s | %s | %d | %v\n",
		emoji,
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		entry.Action,
		entry.Username,
		entry.TenantID,
		entry.IPAddress,
		entry.StatusCode,
		entry.Duration.Round(time.Millisecond),
	)

	if entry.Error != "" {
		fmt.Printf("        Error: %s\n", entry.Error)
	}

	return nil
}

// InMemoryAuditLogger stores audit logs in memory (for testing)
type InMemoryAuditLogger struct {
	entries []*AuditEntry
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		entries: make([]*AuditEntry, 0),
	}
}

func (l *InMemoryAuditLogger) Log(entry *AuditEntry) error {
	l.entries = append(l.entries, entry)
	return nil
}

// GetEntries returns all logged entries
func (l *InMemoryAuditLogger) GetEntries() []*AuditEntry {
	return l.entries
}

// Clear removes all entries
func (l *InMemoryAuditLogger) Clear() {
	l.entries = make([]*AuditEntry, 0)
}
