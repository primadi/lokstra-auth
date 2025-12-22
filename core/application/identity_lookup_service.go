package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// IdentityLookupService handles identity lookup and OAuth2 flow operations
// @RouterService name="identity-lookup-service", prefix="${api-auth-prefix:/api/auth}/core/tenants/{tenant_id}/identities", middlewares=["recovery", "request_logger", "auth"]
type IdentityLookupService struct {
	// @Inject "@store.user-identity-store"
	userIdentityStore repository.UserIdentityStore
	// @Inject "@store.user-store"
	userStore repository.UserStore
}

// FindUserByProvider finds a user by their external provider identity
// @Route "GET /find-by-provider/{provider}/{provider_id}"
func (s *IdentityLookupService) FindUserByProvider(ctx *request.Context, req *domain.FindUserByProviderRequest) (*domain.UserWithIdentity, error) {
	user, err := s.userIdentityStore.FindUserByProvider(ctx, req.TenantID, req.Provider, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("user not found for provider identity: %w", err)
	}

	// Get identity
	identity, err := s.userIdentityStore.GetByProvider(ctx, req.TenantID, user.ID, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	return &domain.UserWithIdentity{
		User:       user,
		Identity:   identity,
		WasCreated: false,
	}, nil
}

// GetOrCreateUserByProvider finds or creates a user from external provider identity
// This is the main endpoint for OAuth2/SAML login flows
// @Route "POST /get-or-create"
func (s *IdentityLookupService) GetOrCreateUserByProvider(ctx *request.Context, req *domain.GetOrCreateUserRequest) (*domain.UserWithIdentity, error) {
	tenantID := req.TenantID
	provider := req.Provider
	providerID := req.ProviderID
	email := req.Email
	username := req.Username
	verified := req.Verified
	metadata := req.Metadata

	// Try to find existing user
	user, err := s.userIdentityStore.FindUserByProvider(ctx, tenantID, provider, providerID)
	if err == nil && user != nil {
		// User exists, get identity
		identity, err := s.userIdentityStore.GetByProvider(ctx, tenantID, user.ID, provider)
		if err != nil {
			return nil, fmt.Errorf("failed to get identity: %w", err)
		}
		return &domain.UserWithIdentity{
			User:       user,
			Identity:   identity,
			WasCreated: false,
		}, nil
	}

	// User doesn't exist, create new user
	newUser := &domain.User{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Username: username,
		Email:    email,
		Status:   domain.UserStatusActive,
		Metadata: metadata,
	}

	if err := s.userStore.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Link provider identity
	identity := &domain.UserIdentity{
		ID:         uuid.New().String(),
		UserID:     newUser.ID,
		TenantID:   tenantID,
		Provider:   provider,
		ProviderID: providerID,
		Email:      email,
		Username:   username,
		Verified:   verified,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userIdentityStore.Create(ctx, identity); err != nil {
		// Rollback: delete user
		_ = s.userStore.Delete(ctx, tenantID, newUser.ID)
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	return &domain.UserWithIdentity{
		User:       newUser,
		Identity:   identity,
		WasCreated: true,
	}, nil
}
