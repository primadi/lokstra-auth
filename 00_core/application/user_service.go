package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra-auth/00_core/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// UserService manages user lifecycle and operations within tenants
// @RouterService name="user-service", prefix="/api/registration/tenants/{tenant_id}/users", middlewares=["recovery", "request-logger"]
type UserService struct {
	// @Inject "user-store"
	Store *service.Cached[repository.UserStore]
	// @Inject "user-app-store"
	UserAppStore *service.Cached[repository.UserAppStore]
	// @Inject "tenant-service"
	TenantService *service.Cached[*TenantService]
	// @Inject "app-service"
	AppService *service.Cached[*AppService]
}

// CreateUser creates a new user within a tenant
// @Route "POST /"
func (s *UserService) CreateUser(ctx *request.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	// Validate tenant exists and is active
	tenant, err := s.TenantService.MustGet().GetTenant(ctx, &domain.GetTenantRequest{
		ID: req.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	if tenant.Status != domain.TenantStatusActive {
		return nil, fmt.Errorf("tenant is not active: %s", tenant.Status)
	}

	// Check if username already exists in tenant
	existing, err := s.Store.MustGet().GetByUsername(ctx, tenant.ID, req.Username)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with username '%s' already exists in tenant '%s'",
			req.Username, tenant.ID)
	}

	// Check if email already exists in tenant
	existing, err = s.Store.MustGet().GetByEmail(ctx, tenant.ID, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with email '%s' already exists in tenant '%s'",
			req.Email, tenant.ID)
	}

	metadata := &map[string]any{}
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// Create user
	user := &domain.User{
		TenantID:  req.TenantID,
		ID:        req.ID,
		Username:  req.Username,
		Email:     req.Email,
		Status:    domain.UserStatusActive,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to store
	if err := s.Store.MustGet().Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID within a tenant
// @Route "GET /{id}"
func (s *UserService) GetUser(ctx *request.Context, req *domain.GetUserRequest) (*domain.User, error) {
	user, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username within a tenant
// @Route "GET /by-username/{username}"
func (s *UserService) GetUserByUsername(ctx *request.Context, req *domain.GetUserByUsernameRequest) (*domain.User, error) {
	user, err := s.Store.MustGet().GetByUsername(ctx, req.TenantID, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email within a tenant
// @Route "GET /by-email/{email}"
func (s *UserService) GetUserByEmail(ctx *request.Context, req *domain.GetUserByEmailRequest) (*domain.User, error) {
	user, err := s.Store.MustGet().GetByEmail(ctx, req.TenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
// @Route "PUT /{id}"
func (s *UserService) UpdateUser(ctx *request.Context, req *domain.UpdateUserRequest) error {
	// Check if user exists
	dUser, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if req.Username != "" && req.Username != dUser.Username {
		// Check if new username already exists
		existing, err := s.Store.MustGet().GetByUsername(ctx, req.TenantID, req.Username)
		if err == nil && existing != nil {
			return fmt.Errorf("user with username '%s' already exists in tenant '%s'",
				req.Username, req.TenantID)
		}
		dUser.Username = req.Username
	}
	if req.Email != "" && req.Email != dUser.Email {
		// Check if new email already exists
		existing, err := s.Store.MustGet().GetByEmail(ctx, req.TenantID, req.Email)
		if err == nil && existing != nil {
			return fmt.Errorf("user with email '%s' already exists in tenant '%s'",
				req.Email, req.TenantID)
		}
		dUser.Email = req.Email
	}

	if req.Status != "" {
		dUser.Status = req.Status
	}

	if req.Metadata != nil {
		dUser.Metadata = req.Metadata
	}

	// Update timestamp
	dUser.UpdatedAt = time.Now()

	// Save to store
	if err := s.Store.MustGet().Update(ctx, dUser); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
// @Route "DELETE /{id}"
func (s *UserService) DeleteUser(ctx *request.Context, req *domain.DeleteUserRequest) error {
	// Check if user exists
	exists, err := s.Store.MustGet().Exists(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found: %s in tenant %s", req.ID, req.TenantID)
	}

	// Delete from store
	if err := s.Store.MustGet().Delete(ctx, req.TenantID, req.ID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers lists all users for a tenant
// @Route "GET /"
func (s *UserService) ListUsers(ctx *request.Context, req *domain.ListUsersRequest) ([]*domain.User, error) {
	users, err := s.Store.MustGet().List(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// AssignUserToApp grants a user access to an app
// Authorization (roles, permissions) will be managed separately in 04_authz layer
// @Route "POST /{user_id}/assign-app"
func (s *UserService) AssignUserToApp(ctx *request.Context, req *domain.AssignUserToAppRequest) error {
	// Verify user exists
	_, err := s.Store.MustGet().Get(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify app exists
	_, err = s.AppService.MustGet().GetApp(ctx, &domain.GetAppRequest{
		TenantID: req.TenantID,
		ID:       req.AppID,
	})
	if err != nil {
		return fmt.Errorf("app not found: %w", err)
	}

	// Check if user already has access
	hasAccess, err := s.UserAppStore.MustGet().HasAccess(ctx, req.TenantID, req.AppID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}
	if hasAccess {
		return fmt.Errorf("user '%s' already has access to app '%s'", req.UserID, req.AppID)
	}

	// Grant access
	if err := s.UserAppStore.MustGet().GrantAccess(ctx, req.TenantID, req.AppID, req.UserID); err != nil {
		return fmt.Errorf("failed to grant app access: %w", err)
	}

	return nil
}

// RemoveUserFromApp revokes a user's access from an app
// @Route "DELETE /{user_id}/remove-app"
func (s *UserService) RemoveUserFromApp(ctx *request.Context, req *domain.RemoveUserFromAppRequest) error {
	// Check if access exists
	hasAccess, err := s.UserAppStore.MustGet().HasAccess(ctx, req.TenantID, req.AppID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return fmt.Errorf("user '%s' does not have access to app '%s'", req.UserID, req.AppID)
	}

	// Revoke access
	if err := s.UserAppStore.MustGet().RevokeAccess(ctx, req.TenantID, req.AppID, req.UserID); err != nil {
		return fmt.Errorf("failed to revoke app access: %w", err)
	}

	return nil
}

// ListUserApps lists all apps a user has access to
// @Route "GET /{user_id}/apps"
func (s *UserService) ListUserApps(ctx *request.Context, req *domain.GetUserRequest) ([]string, error) {
	appIDs, err := s.UserAppStore.MustGet().ListUserApps(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user apps: %w", err)
	}

	return appIDs, nil
}

// ActivateUser activates a user
// @Route "POST /{id}/activate"
func (s *UserService) ActivateUser(ctx *request.Context, req *domain.ActivateUserRequest) error {
	user, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Status = domain.UserStatusActive
	user.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, user); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// SuspendUser suspends a user
// @Route "POST /{id}/suspend"
func (s *UserService) SuspendUser(ctx *request.Context, req *domain.SuspendUserRequest) error {
	user, err := s.Store.MustGet().Get(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Status = domain.UserStatusSuspended
	user.UpdatedAt = time.Now()

	if err := s.Store.MustGet().Update(ctx, user); err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	return nil
}
