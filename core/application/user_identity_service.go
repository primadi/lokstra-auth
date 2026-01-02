package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// UserIdentityService handles user identity management operations
// @RouterService name="user-identity-service", prefix="${api-auth-prefix:/api/auth}/core/tenants/{tenant_id}/users/{user_id}/identities", middlewares=["recovery", "request_logger", "auth"]
type UserIdentityService struct {
	// @Inject "@store.user-identity-store"
	userIdentityStore repository.UserIdentityStore
	// @Inject "@store.user-store"
	userStore repository.UserStore
}

// LinkIdentity links an external identity to a user
// @Route "POST /"
func (s *UserIdentityService) LinkIdentity(ctx *request.Context, req *domain.LinkIdentityRequest) (*domain.UserIdentity, error) {
	// Validate user exists
	user, err := s.userStore.Get(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, domain.ErrUserSuspended
	}

	// Check if provider identity already linked to another user
	existingUser, err := s.userIdentityStore.FindUserByProvider(ctx, req.TenantID, req.Provider, req.ProviderID)
	if err == nil && existingUser != nil && existingUser.ID != req.UserID {
		return nil, domain.ErrDuplicateProviderIdentity
	}

	// Check if this provider already linked to this user
	existingIdentity, err := s.userIdentityStore.GetByProvider(ctx, req.TenantID, req.UserID, req.Provider)
	if err == nil && existingIdentity != nil {
		// Update existing identity
		existingIdentity.ProviderID = req.ProviderID
		existingIdentity.Email = req.Email
		existingIdentity.Username = req.Username
		existingIdentity.Verified = req.Verified
		existingIdentity.Metadata = req.Metadata
		existingIdentity.UpdatedAt = time.Now()

		if err := s.userIdentityStore.Update(ctx, existingIdentity); err != nil {
			return nil, fmt.Errorf("failed to update identity: %w", err)
		}

		return existingIdentity, nil
	}

	// Create new identity
	identity := &domain.UserIdentity{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		TenantID:   req.TenantID,
		Provider:   req.Provider,
		ProviderID: req.ProviderID,
		Email:      req.Email,
		Username:   req.Username,
		Verified:   req.Verified,
		Metadata:   req.Metadata,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := identity.Validate(); err != nil {
		return nil, err
	}

	if err := s.userIdentityStore.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	return identity, nil
}

// UnlinkIdentity removes a linked identity
// @Route "DELETE /{identity_id}"
func (s *UserIdentityService) UnlinkIdentity(ctx *request.Context, req *domain.UnlinkIdentityRequest) error {
	// Validate user exists
	_, err := s.userStore.Get(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check identity belongs to user
	identity, err := s.userIdentityStore.Get(ctx, req.TenantID, req.UserID, req.IdentityID)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	if identity.UserID != req.UserID {
		return fmt.Errorf("identity does not belong to user")
	}

	// Ensure user has at least one identity left (or password)
	identities, err := s.userIdentityStore.List(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check identities: %w", err)
	}

	user, _ := s.userStore.Get(ctx, req.TenantID, req.UserID)
	hasPassword := user.PasswordHash != nil && *user.PasswordHash != ""

	if len(identities) <= 1 && !hasPassword {
		return fmt.Errorf("cannot unlink last identity: user must have at least one authentication method")
	}

	if err := s.userIdentityStore.Delete(ctx, req.TenantID, req.UserID, req.IdentityID); err != nil {
		return fmt.Errorf("failed to unlink identity: %w", err)
	}

	return nil
}

// GetIdentity retrieves a specific identity
// @Route "GET /{identity_id}"
func (s *UserIdentityService) GetIdentity(ctx *request.Context, req *domain.GetIdentityRequest) (*domain.UserIdentity, error) {
	identity, err := s.userIdentityStore.Get(ctx, req.TenantID, req.UserID, req.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	return identity, nil
}

// ListIdentities lists all identities for a user
// @Route "GET /"
func (s *UserIdentityService) ListIdentities(ctx *request.Context, req *domain.ListIdentitiesRequest) ([]*domain.UserIdentity, error) {
	// Validate user exists
	_, err := s.userStore.Get(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	identities, err := s.userIdentityStore.List(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}

	return identities, nil
}

// UpdateIdentity updates an existing identity
// @Route "PUT /{identity_id}"
func (s *UserIdentityService) UpdateIdentity(ctx *request.Context, req *domain.UpdateUserIdentityRequest) (*domain.UserIdentity, error) {
	// Get existing identity
	identity, err := s.userIdentityStore.Get(ctx, req.TenantID, req.UserID, req.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	// Update fields
	if req.Email != "" {
		identity.Email = req.Email
	}
	if req.Username != "" {
		identity.Username = req.Username
	}
	identity.Verified = req.Verified
	if req.Metadata != nil {
		identity.Metadata = req.Metadata
	}
	identity.UpdatedAt = time.Now()

	if err := s.userIdentityStore.Update(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to update identity: %w", err)
	}

	return identity, nil
}
