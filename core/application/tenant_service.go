package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

// @RouterService name="tenant-service", prefix="${api-auth-prefix:/api/auth}/core/tenants", middlewares=["recovery", "request_logger", "auth"]
type TenantService struct {
	// @Inject "@store.tenant-store"
	Store repository.TenantStore
	// @Inject "@store.user-store"
	UserStore repository.UserStore
	// @Inject "@store.app-store"
	AppStore repository.AppStore
	// @Inject "@store.user-app-store"
	UserAppStore repository.UserAppStore
	// @Inject "@email-service"
	EmailService serviceapi.EmailSender
}

// @Route "POST /"
func (s *TenantService) CreateTenant(ctx *request.Context,
	req *domain.CreateTenantRequest) (*domain.Tenant, error) {

	// Validate owner email
	if req.OwnerEmail == "" {
		return nil, fmt.Errorf("owner_email is required")
	}

	// Check if tenant name already exists
	existing, err := s.Store.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with name '%s' already exists", req.Name)
	}

	// Check if tenant ID already exists
	existingByID, _ := s.Store.Get(ctx, req.ID)
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

	owner := &domain.User{
		ID:            ownerID,
		TenantID:      req.ID,
		Username:      ownerUsername,
		Email:         req.OwnerEmail,
		FullName:      req.OwnerFullName,
		Status:        domain.UserStatusActive,
		IsTenantOwner: true,
		Metadata:      &ownerMetadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Note: User will be created AFTER tenant is created
	// because user needs valid tenant_id

	// Initialize settings - use provided or defaults
	settings := &domain.TenantSettings{}
	if req.Settings != nil {
		settings = req.Settings
	}

	// Initialize metadata - use provided or empty map
	metadata := &map[string]any{}
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// Create tenant
	tenant := &domain.Tenant{
		ID:        req.ID,
		Name:      req.Name,
		OwnerID:   ownerID,
		DBDsn:     req.DBDsn,
		DBSchema:  req.DBSchema,
		Status:    domain.TenantStatusActive,
		Settings:  settings,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx.BeginTransaction("db_auth")

	// Save tenant to store
	if err := s.Store.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Now create owner user
	if err := s.UserStore.Create(ctx, owner); err != nil {
		// Rollback tenant if user creation fails - no need because transaction will rollback
		// _ = s.Store.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create owner user: %w", err)
	}

	// Create default admin app for tenant
	defaultAppID := req.ID + "-admin"
	if req.AppID != "" {
		defaultAppID = req.AppID
	}
	defaultApp := &domain.App{
		ID:        defaultAppID,
		TenantID:  req.ID,
		Name:      req.Name + " Admin",
		Type:      "admin",
		Status:    domain.AppStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.AppStore.Create(ctx, defaultApp); err != nil {
		// Rollback on error - no need because transaction will rollback
		// _ = s.UserStore.Delete(ctx, tenant.ID, owner.ID)
		// _ = s.Store.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create default app: %w", err)
	}

	// Grant owner access to default admin app
	if err := s.UserAppStore.GrantAccess(ctx, tenant.ID, defaultApp.ID, owner.ID); err != nil {
		// Rollback on error - no need because transaction will rollback
		// _ = s.AppStore.Delete(ctx, tenant.ID, defaultApp.ID)
		// _ = s.UserStore.Delete(ctx, tenant.ID, owner.ID)
		// _ = s.Store.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to grant owner access to app: %w", err)
	}

	// Send welcome email with password reset link
	if req.SendWelcomeEmail {
		// Generate password reset token
		resetToken := s.GenerateResetToken(owner.Email)

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

	return tenant, nil
}

// GenerateResetToken generates a password reset token for the given email
func (s *TenantService) GenerateResetToken(email string) string {
	// For demonstration, return a dummy token
	// In production, this should generate a secure token and store it with expiration
	return "reset-token-for-" + email
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx *request.Context, req *domain.GetTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.Store.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// @Route "PUT /{id}"
func (s *TenantService) UpdateTenant(ctx *request.Context, req *domain.UpdateTenantRequest) (*domain.Tenant, error) {
	// Get existing tenant
	tenant, err := s.Store.Get(ctx, req.ID)
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
			tenant.Config = &domain.TenantConfig{}
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
	if err := s.Store.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// @Route "DELETE /{id}"
func (s *TenantService) DeleteTenant(ctx *request.Context, req *domain.DeleteTenantRequest) error {
	// Check if tenant exists
	exists, err := s.Store.Exists(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("tenant not found: %s", req.ID)
	}

	// Delete from store
	if err := s.Store.Delete(ctx, req.ID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// @Route "GET /"
func (s *TenantService) ListTenants(ctx *request.Context, req *domain.ListTenantsRequest) ([]*domain.Tenant, error) {
	tenants, err := s.Store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// @Route "POST /{id}/activate"
func (s *TenantService) ActivateTenant(ctx *request.Context, req *domain.ActivateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.Store.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domain.TenantStatusActive
	tenant.UpdatedAt = time.Now()

	if err := s.Store.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	return tenant, nil
}

// @Route "POST /{id}/suspend"
func (s *TenantService) SuspendTenant(ctx *request.Context, req *domain.SuspendTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.Store.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domain.TenantStatusSuspended
	tenant.UpdatedAt = time.Now()

	if err := s.Store.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
	}

	return tenant, nil
}
