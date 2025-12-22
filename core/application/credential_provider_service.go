package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// CredentialProviderService handles credential provider management operations
// @RouterService name="credential-provider-service", prefix="${api-auth-prefix:/api/auth}/core/tenants/{tenant_id}/credential-providers", middlewares=["recovery", "request_logger", "auth"]
type CredentialProviderService struct {
	// @Inject "@store.credential-provider-store"
	providerStore repository.CredentialProviderStore
	// @Inject "@store.tenant-store"
	tenantStore repository.TenantStore
	// @Inject "@store.app-store"
	appStore repository.AppStore
}

// CreateProvider creates a new credential provider
// @Route "POST /"
func (s *CredentialProviderService) CreateProvider(ctx *request.Context, req *domain.CreateProviderRequest) (*domain.CredentialProvider, error) {
	// Validate tenant exists
	tenantExists, err := s.tenantStore.Exists(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant: %w", err)
	}
	if !tenantExists {
		return nil, fmt.Errorf("tenant not found: %s", req.TenantID)
	}

	// Validate app exists (if app_id provided)
	if req.AppID != "" {
		appExists, err := s.appStore.Exists(ctx, req.TenantID, req.AppID)
		if err != nil {
			return nil, fmt.Errorf("failed to check app: %w", err)
		}
		if !appExists {
			return nil, fmt.Errorf("app not found: %s", req.AppID)
		}
	}

	// Create provider
	provider := &domain.CredentialProvider{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		AppID:       req.AppID,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Status:      domain.ProviderStatusActive,
		Config:      req.Config,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := provider.Validate(); err != nil {
		return nil, err
	}

	if err := s.providerStore.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// GetProvider retrieves a provider by ID
// @Route "GET /{provider_id}"
func (s *CredentialProviderService) GetProvider(ctx *request.Context, req *domain.GetProviderRequest) (*domain.CredentialProvider, error) {
	provider, err := s.providerStore.Get(ctx, req.TenantID, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	return provider, nil
}

// UpdateProvider updates an existing provider
// @Route "PUT /{provider_id}"
func (s *CredentialProviderService) UpdateProvider(ctx *request.Context, req *domain.UpdateProviderRequest) (*domain.CredentialProvider, error) {
	// Get existing provider
	provider, err := s.providerStore.Get(ctx, req.TenantID, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Update fields
	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.Description != "" {
		provider.Description = req.Description
	}
	if req.Config != nil {
		provider.Config = req.Config
	}
	if req.Status != "" {
		provider.Status = req.Status
	}
	if req.Metadata != nil {
		provider.Metadata = req.Metadata
	}
	provider.UpdatedAt = time.Now()

	if err := s.providerStore.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	return provider, nil
}

// DeleteProvider deletes a provider
// @Route "DELETE /{provider_id}"
func (s *CredentialProviderService) DeleteProvider(ctx *request.Context, req *domain.DeleteProviderRequest) error {
	// Check provider exists
	_, err := s.providerStore.Get(ctx, req.TenantID, req.ProviderID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	if err := s.providerStore.Delete(ctx, req.TenantID, req.ProviderID); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}

// ListProviders lists all providers for a tenant (optionally filtered by app)
// @Route "GET /"
func (s *CredentialProviderService) ListProviders(ctx *request.Context, req *domain.ListProvidersRequest) ([]*domain.CredentialProvider, error) {
	providers, err := s.providerStore.List(ctx, req.TenantID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Apply filters
	filtered := make([]*domain.CredentialProvider, 0)
	for _, provider := range providers {
		// Filter by type
		if req.Type != "" && provider.Type != req.Type {
			continue
		}
		// Filter by status
		if req.Status != "" && provider.Status != req.Status {
			continue
		}
		filtered = append(filtered, provider)
	}

	return filtered, nil
}

// EnableProvider enables a provider
// @Route "POST /{provider_id}/enable"
func (s *CredentialProviderService) EnableProvider(ctx *request.Context, req *domain.EnableProviderRequest) (*domain.CredentialProvider, error) {
	provider, err := s.providerStore.Get(ctx, req.TenantID, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	provider.Status = domain.ProviderStatusActive
	provider.UpdatedAt = time.Now()

	if err := s.providerStore.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to enable provider: %w", err)
	}

	return provider, nil
}

// DisableProvider disables a provider
// @Route "POST /{provider_id}/disable"
func (s *CredentialProviderService) DisableProvider(ctx *request.Context, req *domain.DisableProviderRequest) (*domain.CredentialProvider, error) {
	provider, err := s.providerStore.Get(ctx, req.TenantID, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	provider.Status = domain.ProviderStatusDisabled
	provider.UpdatedAt = time.Now()

	if err := s.providerStore.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to disable provider: %w", err)
	}

	return provider, nil
}

// GetProvidersByType retrieves all providers of a specific type
// @Route "GET /by-type/{provider_type}"
func (s *CredentialProviderService) GetProvidersByType(ctx *request.Context, req *domain.GetProvidersByTypeRequest) ([]*domain.CredentialProvider, error) {
	providers, err := s.providerStore.ListByType(ctx, req.TenantID, req.AppID, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers by type: %w", err)
	}

	// Filter only active providers
	active := make([]*domain.CredentialProvider, 0)
	for _, provider := range providers {
		if provider.IsActive() {
			active = append(active, provider)
		}
	}

	return active, nil
}

// GetActiveProviderForApp retrieves the active provider for an app
// If no app-specific provider is found, falls back to tenant-level provider
// @Route "GET /active"
func (s *CredentialProviderService) GetActiveProviderForApp(ctx *request.Context, req *domain.GetActiveProviderForAppRequest) (*domain.CredentialProvider, error) {
	tenantID := req.TenantID
	appID := req.AppID
	providerType := req.Type
	// First, try to get app-level provider
	appProviders, err := s.providerStore.ListByType(ctx, tenantID, appID, providerType)
	if err == nil && len(appProviders) > 0 {
		for _, provider := range appProviders {
			if provider.IsActive() {
				return provider, nil
			}
		}
	}

	// Fallback to tenant-level provider (AppID = "")
	tenantProviders, err := s.providerStore.ListByType(ctx, tenantID, "", providerType)
	if err != nil {
		return nil, fmt.Errorf("no active provider found for type %s: %w", providerType, err)
	}

	for _, provider := range tenantProviders {
		if provider.IsActive() {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no active provider found for type: %s", providerType)
}
