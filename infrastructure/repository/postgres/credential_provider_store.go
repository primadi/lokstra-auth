package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// PostgresCredentialProviderStore is a PostgreSQL implementation of CredentialProviderStore
// @Service "postgres-credential-provider-store"
type PostgresCredentialProviderStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.CredentialProviderStore = (*PostgresCredentialProviderStore)(nil)

// Create creates a new credential provider
func (s *PostgresCredentialProviderStore) Create(ctx context.Context, provider *domain.CredentialProvider) error {
	if err := provider.Validate(); err != nil {
		return err
	}

	// Serialize config and metadata
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var metadataJSON []byte
	if provider.Metadata != nil {
		metadataJSON, err = json.Marshal(*provider.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO credential_providers (
			id, tenant_id, app_id, type, name, description,
			status, config, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Use NULL for empty appID
	var appID *string
	if provider.AppID != "" {
		appID = &provider.AppID
	}

	_, err = s.dbPool.Exec(
		ctx,
		query,
		provider.ID,
		provider.TenantID,
		appID,
		provider.Type,
		provider.Name,
		provider.Description,
		provider.Status,
		configJSON,
		metadataJSON,
		provider.CreatedAt,
		provider.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create credential provider: %w", err)
	}

	return nil
}

// Get retrieves a provider by ID
func (s *PostgresCredentialProviderStore) Get(ctx context.Context, tenantID, providerID string) (*domain.CredentialProvider, error) {
	query := `
		SELECT 
			id, tenant_id, COALESCE(app_id, ''), type, name, description,
			status, config, metadata, created_at, updated_at
		FROM credential_providers
		WHERE tenant_id = $1 AND id = $2
	`

	provider := &domain.CredentialProvider{}
	var configJSON, metadataJSON []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, providerID).Scan(
		&provider.ID,
		&provider.TenantID,
		&provider.AppID,
		&provider.Type,
		&provider.Name,
		&provider.Description,
		&provider.Status,
		&configJSON,
		&metadataJSON,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrProviderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential provider: %w", err)
	}

	// Deserialize config
	if err = json.Unmarshal(configJSON, &provider.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		provider.Metadata = &metadata
	}

	return provider, nil
}

// Update updates an existing provider
func (s *PostgresCredentialProviderStore) Update(ctx context.Context, provider *domain.CredentialProvider) error {
	if err := provider.Validate(); err != nil {
		return err
	}

	// Serialize config and metadata
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var metadataJSON []byte
	if provider.Metadata != nil {
		metadataJSON, err = json.Marshal(*provider.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE credential_providers
		SET 
			name = $1,
			description = $2,
			status = $3,
			config = $4,
			metadata = $5,
			updated_at = $6
		WHERE tenant_id = $7 AND id = $8
	`

	result, err := s.dbPool.Exec(
		ctx,
		query,
		provider.Name,
		provider.Description,
		provider.Status,
		configJSON,
		metadataJSON,
		provider.UpdatedAt,
		provider.TenantID,
		provider.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update credential provider: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrProviderNotFound
	}

	return nil
}

// Delete deletes a provider
func (s *PostgresCredentialProviderStore) Delete(ctx context.Context, tenantID, providerID string) error {
	query := `
		DELETE FROM credential_providers
		WHERE tenant_id = $1 AND id = $2
	`

	result, err := s.dbPool.Exec(ctx, query, tenantID, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete credential provider: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrProviderNotFound
	}

	return nil
}

// List lists all providers for tenant (optionally filtered by app)
func (s *PostgresCredentialProviderStore) List(ctx context.Context, tenantID, appID string) ([]*domain.CredentialProvider, error) {
	var query string
	var args []any

	if appID != "" {
		// Return providers for specific app OR tenant-level providers
		query = `
			SELECT 
				id, tenant_id, COALESCE(app_id, ''), type, name, description,
				status, config, metadata, created_at, updated_at
			FROM credential_providers
			WHERE tenant_id = $1 AND (app_id = $2 OR app_id IS NULL)
			ORDER BY created_at DESC
		`
		args = []any{tenantID, appID}
	} else {
		// Return only tenant-level providers (app_id IS NULL)
		query = `
			SELECT 
				id, tenant_id, COALESCE(app_id, ''), type, name, description,
				status, config, metadata, created_at, updated_at
			FROM credential_providers
			WHERE tenant_id = $1 AND app_id IS NULL
			ORDER BY created_at DESC
		`
		args = []any{tenantID}
	}

	rows, err := s.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credential providers: %w", err)
	}
	defer rows.Close()

	var providers []*domain.CredentialProvider

	for rows.Next() {
		provider := &domain.CredentialProvider{}
		var configJSON, metadataJSON []byte

		err := rows.Scan(
			&provider.ID,
			&provider.TenantID,
			&provider.AppID,
			&provider.Type,
			&provider.Name,
			&provider.Description,
			&provider.Status,
			&configJSON,
			&metadataJSON,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential provider: %w", err)
		}

		// Deserialize config
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			provider.Metadata = &metadata
		}

		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credential providers: %w", err)
	}

	return providers, nil
}

// ListByType lists all providers of a specific type for tenant+app
func (s *PostgresCredentialProviderStore) ListByType(ctx context.Context, tenantID, appID string, providerType domain.ProviderType) ([]*domain.CredentialProvider, error) {
	var query string
	var args []any

	if appID != "" {
		// Return providers for specific app OR tenant-level providers
		query = `
			SELECT 
				id, tenant_id, COALESCE(app_id, ''), type, name, description,
				status, config, metadata, created_at, updated_at
			FROM credential_providers
			WHERE tenant_id = $1 AND type = $2 AND (app_id = $3 OR app_id IS NULL)
			ORDER BY created_at DESC
		`
		args = []any{tenantID, providerType, appID}
	} else {
		// Return only tenant-level providers
		query = `
			SELECT 
				id, tenant_id, COALESCE(app_id, ''), type, name, description,
				status, config, metadata, created_at, updated_at
			FROM credential_providers
			WHERE tenant_id = $1 AND type = $2 AND app_id IS NULL
			ORDER BY created_at DESC
		`
		args = []any{tenantID, providerType}
	}

	rows, err := s.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credential providers by type: %w", err)
	}
	defer rows.Close()

	var providers []*domain.CredentialProvider

	for rows.Next() {
		provider := &domain.CredentialProvider{}
		var configJSON, metadataJSON []byte

		err := rows.Scan(
			&provider.ID,
			&provider.TenantID,
			&provider.AppID,
			&provider.Type,
			&provider.Name,
			&provider.Description,
			&provider.Status,
			&configJSON,
			&metadataJSON,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential provider: %w", err)
		}

		// Deserialize config
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			provider.Metadata = &metadata
		}

		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credential providers: %w", err)
	}

	return providers, nil
}

// Exists checks if a provider exists
func (s *PostgresCredentialProviderStore) Exists(ctx context.Context, tenantID, providerID string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM credential_providers
		WHERE tenant_id = $1 AND id = $2
	`

	return s.dbPool.IsExists(ctx, query, tenantID, providerID)
}
