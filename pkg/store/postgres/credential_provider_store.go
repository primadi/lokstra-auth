package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for credential provider operations
const (
	credentialProviderColumns = `id, tenant_id, app_id, type, name, description, status, config, metadata, created_at, updated_at`

	queryCredentialProviderInsert = `
		INSERT INTO credential_providers (
			id, tenant_id, app_id, type, name, description, status, config, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryCredentialProviderSelect = `SELECT ` + credentialProviderColumns + ` FROM credential_providers`

	queryCredentialProviderGet        = queryCredentialProviderSelect + ` WHERE tenant_id = $1 AND id = $2`
	queryCredentialProviderList       = queryCredentialProviderSelect + ` WHERE tenant_id = $1 AND (app_id = $2 OR app_id IS NULL) ORDER BY created_at DESC`
	queryCredentialProviderListByType = queryCredentialProviderSelect + ` WHERE tenant_id = $1 AND (app_id = $2 OR app_id IS NULL) AND type = $3 ORDER BY created_at DESC`
	queryCredentialProviderExists     = `SELECT 1 FROM credential_providers WHERE tenant_id = $1 AND id = $2`

	queryCredentialProviderUpdate = `
		UPDATE credential_providers
		SET type = $1, name = $2, description = $3, status = $4, config = $5, metadata = $6, updated_at = NOW()
		WHERE tenant_id = $7 AND id = $8`

	queryCredentialProviderDelete = `DELETE FROM credential_providers WHERE tenant_id = $1 AND id = $2`
)

// @Service "pg-credential-provider-store"
type pgCredentialProviderStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) Create(ctx context.Context, provider *domaincore.CredentialProvider) error {
	config, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	metadata, err := json.Marshal(provider.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryCredentialProviderInsert,
		provider.ID, provider.TenantID, provider.AppID, provider.Type, provider.Name,
		provider.Description, provider.Status, config, metadata,
	)
	return err
}

// Get implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) Get(ctx context.Context, tenantID, providerID string) (*domaincore.CredentialProvider, error) {
	provider := &domaincore.CredentialProvider{}
	var config, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryCredentialProviderGet, tenantID, providerID).Scan(
		&provider.ID, &provider.TenantID, &provider.AppID, &provider.Type, &provider.Name,
		&provider.Description, &provider.Status, &config, &metadata,
		&provider.CreatedAt, &provider.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("credential provider not found: tenant=%s, provider=%s", tenantID, providerID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalProviderFields(provider, config, metadata); err != nil {
		return nil, err
	}

	return provider, nil
}

// Update implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) Update(ctx context.Context, provider *domaincore.CredentialProvider) error {
	config, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	metadata, err := json.Marshal(provider.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryCredentialProviderUpdate,
		provider.Type, provider.Name, provider.Description, provider.Status, config, metadata,
		provider.TenantID, provider.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("credential provider not found: tenant=%s, provider=%s", provider.TenantID, provider.ID)
	}
	return nil
}

// Delete implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) Delete(ctx context.Context, tenantID, providerID string) error {
	result, err := p.dbPool.Exec(ctx, queryCredentialProviderDelete, tenantID, providerID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("credential provider not found: tenant=%s, provider=%s", tenantID, providerID)
	}
	return nil
}

// List implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) List(ctx context.Context, tenantID, appID string) ([]*domaincore.CredentialProvider, error) {
	rows, err := p.dbPool.Query(ctx, queryCredentialProviderList, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanProviders(rows)
}

// ListByType implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) ListByType(ctx context.Context, tenantID, appID string, providerType domaincore.ProviderType) ([]*domaincore.CredentialProvider, error) {
	rows, err := p.dbPool.Query(ctx, queryCredentialProviderListByType, tenantID, appID, providerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanProviders(rows)
}

// Exists implements [store.CredentialProviderStore].
func (p *pgCredentialProviderStore) Exists(ctx context.Context, tenantID, providerID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryCredentialProviderExists, tenantID, providerID)
}

var _ store.CredentialProviderStore = (*pgCredentialProviderStore)(nil)

// unmarshalProviderFields unmarshals JSON fields into provider struct
func (p *pgCredentialProviderStore) unmarshalProviderFields(provider *domaincore.CredentialProvider, config, metadata []byte) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &provider.Config); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		provider.Metadata = &m
	}
	return nil
}

func (p *pgCredentialProviderStore) scanProviders(rows serviceapi.Rows) ([]*domaincore.CredentialProvider, error) {
	providers := make([]*domaincore.CredentialProvider, 0, 10)

	for rows.Next() {
		provider := &domaincore.CredentialProvider{}
		var config, metadata []byte

		err := rows.Scan(
			&provider.ID, &provider.TenantID, &provider.AppID, &provider.Type, &provider.Name,
			&provider.Description, &provider.Status, &config, &metadata,
			&provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalProviderFields(provider, config, metadata); err != nil {
			return nil, err
		}

		providers = append(providers, provider)
	}

	return providers, rows.Err()
}
