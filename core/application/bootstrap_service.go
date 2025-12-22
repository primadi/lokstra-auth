package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// BootstrapService handles platform initialization and tenant creation with auto-admin
//
// @RouterService name="bootstrap-service", prefix="${api-auth-prefix:/api/auth}/bootstrap"
type BootstrapService struct {
	// @Inject "@store.tenant-store"
	tenantRepo repository.TenantStore
	// @Inject "@store.app-store"
	appRepo repository.AppStore
	// @Inject "@store.user-store"
	userRepo repository.UserStore
	// @Inject "@store.user-app-store"
	userAppRepo repository.UserAppStore
}

// CreateTenantWithAdminRequest defines the request to create a tenant with auto-admin
type CreateTenantWithAdminRequest struct {
	// Tenant details
	TenantID string                 `json:"id"`
	Name     string                 `json:"name"`
	Domain   string                 `json:"domain"`
	DbDSN    string                 `json:"db_dsn"`
	DbSchema string                 `json:"db_schema"`
	Settings *domain.TenantSettings `json:"settings,omitempty"`
	Metadata map[string]any         `json:"metadata,omitempty"`

	// Admin user details
	Admin struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"` // Plain password, will be hashed
		FullName string `json:"full_name,omitempty"`
	} `json:"admin"`
}

// CreateTenantWithAdminResponse returns the created entities
type CreateTenantWithAdminResponse struct {
	Tenant            *domain.Tenant `json:"tenant"`
	DefaultApp        *domain.App    `json:"default_app"`
	AdminUser         *domain.User   `json:"admin_user"`
	AdminUsername     string         `json:"admin_username"`
	AdminTempPassword string         `json:"admin_temp_password,omitempty"` // Only if auto-generated
}

// CreateTenantWithAdmin creates a new tenant with automatic admin user
// This is typically used by platform admins to bootstrap new tenants
//
// @Route "POST /"
// @Middleware ["recovery", "request_logger", "auth", "platform_admin"]
func (s *BootstrapService) CreateTenantWithAdmin(
	ctx context.Context,
	req CreateTenantWithAdminRequest,
) (*CreateTenantWithAdminResponse, error) {
	// Validate request
	if err := s.validateCreateTenantRequest(req); err != nil {
		return nil, err
	}

	// Check if tenant already exists
	existing, err := s.tenantRepo.Get(ctx, req.TenantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with ID '%s' already exists", req.TenantID)
	}

	// 1. Create tenant
	tenant := &domain.Tenant{
		ID:       req.TenantID,
		Name:     req.Name,
		Domain:   req.Domain,
		DBDsn:    req.DbDSN,
		DBSchema: req.DbSchema,
		Status:   "active",
		Settings: req.Settings,
		Metadata: &req.Metadata,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 2. Create default admin app
	appID := req.TenantID + "-admin"
	isDefaultAdminApp := true
	app := &domain.App{
		ID:       appID,
		TenantID: req.TenantID,
		Name:     req.Name + " Admin",
		Type:     "admin",
		Status:   "active",
		Config: &domain.AppConfig{
			CustomSettings: map[string]any{
				"is_default_admin_app": isDefaultAdminApp,
			},
		},
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		// Rollback tenant
		_ = s.tenantRepo.Delete(ctx, req.TenantID)
		return nil, fmt.Errorf("failed to create admin app: %w", err)
	}

	// 3. Create admin user
	userID := req.TenantID + "-admin-user"
	metadata := map[string]any{
		"is_tenant_owner": true,
		"created_via":     "bootstrap",
	}
	user := &domain.User{
		ID:            userID,
		TenantID:      req.TenantID,
		Username:      req.Admin.Username,
		Email:         req.Admin.Email,
		FullName:      req.Admin.FullName,
		Status:        "active",
		IsTenantOwner: true,
		Metadata:      &metadata,
	}

	// Hash password using UserStore.SetPassword after user creation
	// Note: We'll create user first, then set password

	if err := s.userRepo.Create(ctx, user); err != nil {
		// Rollback
		_ = s.appRepo.Delete(ctx, req.TenantID, appID)
		_ = s.tenantRepo.Delete(ctx, req.TenantID)
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Set password hash (using bcrypt via UserStore)
	// TODO: Use proper password hashing with crypto/bcrypt
	if err := s.userRepo.SetPassword(ctx, req.TenantID, userID, req.Admin.Password); err != nil {
		// Rollback
		_ = s.userRepo.Delete(ctx, req.TenantID, userID)
		_ = s.appRepo.Delete(ctx, req.TenantID, appID)
		_ = s.tenantRepo.Delete(ctx, req.TenantID)
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// 4. Grant admin user access to admin app
	if err := s.userAppRepo.GrantAccess(ctx, req.TenantID, appID, userID); err != nil {
		// Rollback
		_ = s.userRepo.Delete(ctx, req.TenantID, userID)
		_ = s.appRepo.Delete(ctx, req.TenantID, appID)
		_ = s.tenantRepo.Delete(ctx, req.TenantID)
		return nil, fmt.Errorf("failed to grant admin access: %w", err)
	}

	// TODO: Assign tenant-admin role to user (requires RBAC integration)

	return &CreateTenantWithAdminResponse{
		Tenant:        tenant,
		DefaultApp:    app,
		AdminUser:     user,
		AdminUsername: req.Admin.Username,
	}, nil
}

func (s *BootstrapService) validateCreateTenantRequest(req CreateTenantWithAdminRequest) error {
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if req.Name == "" {
		return errors.New("tenant name is required")
	}
	if req.Domain == "" {
		return errors.New("domain is required")
	}
	if req.DbDSN == "" {
		return errors.New("db_dsn is required")
	}
	if req.DbSchema == "" {
		return errors.New("db_schema is required")
	}
	if req.Admin.Username == "" {
		return errors.New("admin username is required")
	}
	if req.Admin.Email == "" {
		return errors.New("admin email is required")
	}
	if req.Admin.Password == "" {
		return errors.New("admin password is required")
	}
	if len(req.Admin.Password) < 8 {
		return errors.New("admin password must be at least 8 characters")
	}
	return nil
}

// IsPlatformInitialized checks if the platform has been bootstrapped
//
// @Route "GET /status"
func (s *BootstrapService) IsPlatformInitialized(ctx context.Context) (bool, error) {
	// Check if platform tenant exists
	platform, err := s.tenantRepo.Get(ctx, "platform")
	if err != nil {
		return false, nil
	}
	return platform != nil, nil
}
