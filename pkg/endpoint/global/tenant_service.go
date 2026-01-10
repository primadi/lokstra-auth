package application

import (
	"fmt"
	"time"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/domain/event"
	dtocore "github.com/primadi/lokstra-auth/pkg/dto/core"
	"github.com/primadi/lokstra-auth/pkg/service/token"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

// @EndpointService name="tenant-service", prefix="${auth.path-prefix:/auth}/tenants", middlewares=["recovery", "request_logger"]
type TenantService struct {
	// @Inject "@auth.tenant-store"
	TenantStore store.TenantStore
	// @Inject "@auth.user-store"
	UserStore store.UserStore
	// @Inject "@auth.app-store"
	AppStore store.AppStore
	// @Inject "@auth.user-app-store"
	UserAppStore store.UserAppStore
	// @Inject "@email-service"
	EmailService serviceapi.EmailSender
	// @Inject "@auth.token-manager"
	TokenManager token.TokenManager
	// @Inject "@auth.event-bus"
	EventBus serviceapi.EventBus
}

// @Route "POST /"
func (s *TenantService) CreateTenant(ctx *request.Context,
	req *dtocore.CreateTenantRequest) (*domaincore.Tenant, error) {

	// Validate owner email
	if req.OwnerEmail == "" {
		return nil, fmt.Errorf("owner_email is required")
	}

	// Check if tenant name already exists
	existing, err := s.TenantStore.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with name '%s' already exists", req.Name)
	}

	// Check if tenant ID already exists
	existingByID, _ := s.TenantStore.Get(ctx, req.ID)
	if existingByID != nil {
		return nil, fmt.Errorf("tenant with ID '%s' already exists", req.ID)
	}

	// Generate owner username from email if not provided
	ownerUsername := req.OwnerUsername
	if ownerUsername == "" {
		// Extract username from email (before @)
		atIndex := 0
		for i, ch := range req.OwnerEmail {
			if ch == '@' {
				atIndex = i
				break
			}
		}
		if atIndex > 0 {
			ownerUsername = req.OwnerEmail[:atIndex]
		} else {
			ownerUsername = req.OwnerEmail
		}
	}

	// Create owner user first
	// User will be created in the new tenant's context
	ownerID := req.ID + "-owner" // Generate unique owner ID
	ownerMetadata := map[string]any{
		"is_tenant_owner": true,
		"created_via":     "tenant_creation",
	}

	owner := &domaincore.User{
		ID:            ownerID,
		TenantID:      req.ID,
		Username:      ownerUsername,
		Email:         req.OwnerEmail,
		FullName:      req.OwnerFullName,
		Status:        domaincore.UserStatusActive,
		IsTenantOwner: true,
		Metadata:      &ownerMetadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Note: User will be created AFTER tenant is created
	// because user needs valid tenant_id

	// Initialize settings - use provided or defaults
	settings := &domaincore.TenantSettings{}
	if req.Settings != nil {
		settings = req.Settings
	}

	// Initialize metadata - use provided or empty map
	metadata := &map[string]any{}
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// Create tenant
	tenant := &domaincore.Tenant{
		ID:        req.ID,
		Name:      req.Name,
		OwnerID:   nil,
		DBDsn:     req.DBDsn,
		DBSchema:  req.DBSchema,
		Status:    domaincore.TenantStatusActive,
		Settings:  settings,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx.BeginTransaction("@auth.db_auth")

	// 1. Save tenant to store
	if err := s.TenantStore.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 2. Create owner user
	if err := s.UserStore.Create(ctx, owner); err != nil {
		// Rollback tenant if user creation fails - no need because transaction will rollback
		// _ = s.Store.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create owner user: %w", err)
	}

	tenant.OwnerID = &owner.ID
	if err := s.TenantStore.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant with owner ID: %w", err)
	}

	// 3. Send welcome email with password reset link
	if req.SendWelcomeEmail {
		// Generate password reset token using TokenManager
		resetToken, err := s.TokenManager.GenerateResetToken(ctx, owner.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to generate reset token: %w", err)
		}

		message := &serviceapi.EmailMessage{
			To: []string{owner.Email},
			Subject: fmt.Sprintf(lokstra_registry.GetConfig("reset-password.subject",
				"Welcome to %s - Set Your Password"), tenant.Name),
			Body: fmt.Sprintf(lokstra_registry.GetConfig("reset-password.template",
				"Hello %s,\n\n"+
					"Welcome to %s! Please set your password using the following link:\n\n"+
					"https://example.com/reset-password?token=%s\n\n"+
					"Best regards,\nThe %s Team"), owner.FullName, tenant.Name, resetToken, tenant.Name),
		}
		// Send email
		s.EmailService.Send(ctx, message)
	}

	// 4. Publish TENANT_CREATED event
	if s.EventBus != nil {
		evt := event.NewTenantCreatedEvent(tenant, ownerID)
		if err := s.EventBus.Publish(ctx, evt); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to publish TENANT_CREATED event: %v\n", err)
		}
	}

	return tenant, nil
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx *request.Context, req *dtocore.GetTenantRequest) (*domaincore.Tenant, error) {
	tenant, err := s.TenantStore.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// @Route "PUT /{id}"
func (s *TenantService) UpdateTenant(ctx *request.Context, req *dtocore.UpdateTenantRequest) (*domaincore.Tenant, error) {
	// Get existing tenant
	tenant, err := s.TenantStore.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.DBDsn != "" {
		tenant.DBDsn = req.DBDsn
	}
	if req.DBSchema != "" {
		tenant.DBSchema = req.DBSchema
	}
	if req.Settings != nil {
		tenant.Settings = req.Settings
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	// MERGE Config - preserve existing sub-configs if not provided
	if req.Config != nil {
		if tenant.Config == nil {
			tenant.Config = &domaincore.TenantConfig{}
		}

		// Only update sub-configs that are provided
		if req.Config.DefaultCredentials != nil {
			tenant.Config.DefaultCredentials = req.Config.DefaultCredentials
		}
		if req.Config.DefaultTokenConfig != nil {
			tenant.Config.DefaultTokenConfig = req.Config.DefaultTokenConfig
		}
		if req.Config.Security != nil {
			tenant.Config.Security = req.Config.Security
		}
	}

	// Update timestamp
	tenant.UpdatedAt = time.Now()

	// Save to store
	if err := s.TenantStore.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Publish TENANT_UPDATED event
	if s.EventBus != nil {
		evt := event.NewTenantUpdatedEvent(tenant, nil) // TODO: pass previous state
		if err := s.EventBus.Publish(ctx, evt); err != nil {
			fmt.Printf("Failed to publish TENANT_UPDATED event: %v\n", err)
		}
	}

	return tenant, nil
}

// @Route "DELETE /{id}"
func (s *TenantService) DeleteTenant(ctx *request.Context, req *dtocore.DeleteTenantRequest) error {
	// Get tenant before deleting (for event)
	tenant, err := s.TenantStore.Get(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Delete from store
	if err := s.TenantStore.Delete(ctx, req.ID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Publish TENANT_DELETED event
	if s.EventBus != nil {
		evt := event.NewTenantDeletedEvent(tenant)
		if err := s.EventBus.Publish(ctx, evt); err != nil {
			fmt.Printf("Failed to publish TENANT_DELETED event: %v\n", err)
		}
	}

	return nil
}

// @Route "GET /"
func (s *TenantService) ListTenants(ctx *request.Context, req *dtocore.ListTenantsRequest) ([]*domaincore.Tenant, error) {
	tenants, err := s.TenantStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// @Route "POST /{id}/activate"
func (s *TenantService) ActivateTenant(ctx *request.Context, req *dtocore.ActivateTenantRequest) (*domaincore.Tenant, error) {
	tenant, err := s.TenantStore.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domaincore.TenantStatusActive
	tenant.UpdatedAt = time.Now()

	if err := s.TenantStore.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	// Publish TENANT_ACTIVATED event
	if s.EventBus != nil {
		evt := event.NewTenantActivatedEvent(tenant)
		if err := s.EventBus.Publish(ctx, evt); err != nil {
			fmt.Printf("Failed to publish TENANT_ACTIVATED event: %v\n", err)
		}
	}

	return tenant, nil
}

// @Route "POST /{id}/suspend"
func (s *TenantService) SuspendTenant(ctx *request.Context, req *dtocore.SuspendTenantRequest) (*domaincore.Tenant, error) {
	tenant, err := s.TenantStore.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domaincore.TenantStatusSuspended
	tenant.UpdatedAt = time.Now()

	if err := s.TenantStore.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
	}

	// Publish TENANT_SUSPENDED event
	if s.EventBus != nil {
		evt := event.NewTenantSuspendedEvent(tenant)
		if err := s.EventBus.Publish(ctx, evt); err != nil {
			fmt.Printf("Failed to publish TENANT_SUSPENDED event: %v\n", err)
		}
	}

	return tenant, nil
}
