package event

import (
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra/serviceapi"
)

// serviceapi.Event Types
const (
	// Tenant Events
	EventTenantCreated   serviceapi.EventType = "TENANT_CREATED"
	EventTenantUpdated   serviceapi.EventType = "TENANT_UPDATED"
	EventTenantDeleted   serviceapi.EventType = "TENANT_DELETED"
	EventTenantActivated serviceapi.EventType = "TENANT_ACTIVATED"
	EventTenantSuspended serviceapi.EventType = "TENANT_SUSPENDED"

	// User Events
	EventUserCreated   serviceapi.EventType = "USER_CREATED"
	EventUserUpdated   serviceapi.EventType = "USER_UPDATED"
	EventUserDeleted   serviceapi.EventType = "USER_DELETED"
	EventUserActivated serviceapi.EventType = "USER_ACTIVATED"
	EventUserSuspended serviceapi.EventType = "USER_SUSPENDED"

	// Authentication Events
	EventUserLoggedIn           serviceapi.EventType = "USER_LOGGED_IN"
	EventUserLoggedOut          serviceapi.EventType = "USER_LOGGED_OUT"
	EventUserLoginFailed        serviceapi.EventType = "USER_LOGIN_FAILED"
	EventPasswordChanged        serviceapi.EventType = "PASSWORD_CHANGED"
	EventPasswordResetRequested serviceapi.EventType = "PASSWORD_RESET_REQUESTED"
	EventPasswordReset          serviceapi.EventType = "PASSWORD_RESET"

	// Session Events
	EventSessionCreated serviceapi.EventType = "SESSION_CREATED"
	EventSessionExpired serviceapi.EventType = "SESSION_EXPIRED"
	EventSessionRevoked serviceapi.EventType = "SESSION_REVOKED"

	// Token Events
	EventRefreshTokenCreated serviceapi.EventType = "REFRESH_TOKEN_CREATED"
	EventRefreshTokenUsed    serviceapi.EventType = "REFRESH_TOKEN_USED"
	EventRefreshTokenRevoked serviceapi.EventType = "REFRESH_TOKEN_REVOKED"
	EventRefreshTokenReused  serviceapi.EventType = "REFRESH_TOKEN_REUSED" // Security serviceapi.Event!

	// App Events
	EventAppCreated serviceapi.EventType = "APP_CREATED"
	EventAppUpdated serviceapi.EventType = "APP_UPDATED"
	EventAppDeleted serviceapi.EventType = "APP_DELETED"
)

// serviceapi.Event Payloads

// TenantEventPayload is the payload for tenant-related events
type TenantEventPayload struct {
	Tenant   *domaincore.Tenant
	OwnerID  string             // For TENANT_CREATED
	Previous *domaincore.Tenant // For TENANT_UPDATED (before changes)
}

// UserEventPayload is the payload for user-related events
type UserEventPayload struct {
	User     *domaincore.User
	TenantID string
	Previous *domaincore.User // For USER_UPDATED (before changes)
}

// AuthEventPayload is the payload for authentication events
type AuthEventPayload struct {
	UserID    string
	TenantID  string
	AppID     string
	SessionID string
	IPAddress string
	UserAgent string
	Success   bool
	Reason    string // For failures
}

// SessionEventPayload is the payload for session events
type SessionEventPayload struct {
	Session  *domaincore.Session
	UserID   string
	TenantID string
	AppID    string
	Reason   string // For revocation/expiration
}

// RefreshTokenEventPayload is the payload for refresh token events
type RefreshTokenEventPayload struct {
	RefreshToken *domaincore.RefreshToken
	UserID       string
	TenantID     string
	AppID        string
	SessionID    string
	IsReuse      bool   // For REFRESH_TOKEN_REUSED
	DetectedAt   string // When reuse was detected
}

// AppEventPayload is the payload for app events
type AppEventPayload struct {
	App      *domaincore.App
	TenantID string
	Previous *domaincore.App // For APP_UPDATED (before changes)
}

// Helper functions to create events

// NewTenantCreatedEvent creates a TENANT_CREATED serviceapi.Event
func NewTenantCreatedEvent(tenant *domaincore.Tenant, ownerID string) serviceapi.Event {
	return serviceapi.Event{
		Type: EventTenantCreated,
		Payload: &TenantEventPayload{
			Tenant:  tenant,
			OwnerID: ownerID,
		},
	}
}

// NewTenantUpdatedEvent creates a TENANT_UPDATED serviceapi.Event
func NewTenantUpdatedEvent(tenant *domaincore.Tenant, previous *domaincore.Tenant) serviceapi.Event {
	return serviceapi.Event{
		Type: EventTenantUpdated,
		Payload: &TenantEventPayload{
			Tenant:   tenant,
			Previous: previous,
		},
	}
}

// NewTenantDeletedEvent creates a TENANT_DELETED serviceapi.Event
func NewTenantDeletedEvent(tenant *domaincore.Tenant) serviceapi.Event {
	return serviceapi.Event{
		Type: EventTenantDeleted,
		Payload: &TenantEventPayload{
			Tenant: tenant,
		},
	}
}

// NewTenantActivatedEvent creates a TENANT_ACTIVATED serviceapi.Event
func NewTenantActivatedEvent(tenant *domaincore.Tenant) serviceapi.Event {
	return serviceapi.Event{
		Type: EventTenantActivated,
		Payload: &TenantEventPayload{
			Tenant: tenant,
		},
	}
}

// NewTenantSuspendedEvent creates a TENANT_SUSPENDED serviceapi.Event
func NewTenantSuspendedEvent(tenant *domaincore.Tenant) serviceapi.Event {
	return serviceapi.Event{
		Type: EventTenantSuspended,
		Payload: &TenantEventPayload{
			Tenant: tenant,
		},
	}
}

// NewUserCreatedEvent creates a USER_CREATED serviceapi.Event
func NewUserCreatedEvent(user *domaincore.User) serviceapi.Event {
	return serviceapi.Event{
		Type: EventUserCreated,
		Payload: &UserEventPayload{
			User:     user,
			TenantID: user.TenantID,
		},
	}
}

// NewUserLoggedInEvent creates a USER_LOGGED_IN serviceapi.Event
func NewUserLoggedInEvent(userID, tenantID, appID, sessionID, ipAddress, userAgent string) serviceapi.Event {
	return serviceapi.Event{
		Type: EventUserLoggedIn,
		Payload: &AuthEventPayload{
			UserID:    userID,
			TenantID:  tenantID,
			AppID:     appID,
			SessionID: sessionID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Success:   true,
		},
	}
}

// NewUserLoginFailedEvent creates a USER_LOGIN_FAILED serviceapi.Event
func NewUserLoginFailedEvent(userID, tenantID, appID, ipAddress, userAgent, reason string) serviceapi.Event {
	return serviceapi.Event{
		Type: EventUserLoginFailed,
		Payload: &AuthEventPayload{
			UserID:    userID,
			TenantID:  tenantID,
			AppID:     appID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Success:   false,
			Reason:    reason,
		},
	}
}

// NewRefreshTokenReusedEvent creates a REFRESH_TOKEN_REUSED serviceapi.Event (security!)
func NewRefreshTokenReusedEvent(token *domaincore.RefreshToken, detectedAt string) serviceapi.Event {
	return serviceapi.Event{
		Type: EventRefreshTokenReused,
		Payload: &RefreshTokenEventPayload{
			RefreshToken: token,
			UserID:       token.UserID,
			TenantID:     token.TenantID,
			AppID:        *token.AppID,
			IsReuse:      true,
			DetectedAt:   detectedAt,
		},
	}
}
