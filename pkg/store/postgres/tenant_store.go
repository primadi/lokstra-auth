package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for tenant operations
const (
	tenantColumns = `id, name, domain, owner_id, db_dsn, db_schema, status,
	                 config, settings, metadata, created_at, updated_at, deleted_at`

	queryTenantInsert = `
		INSERT INTO tenants (
			id, name, domain, owner_id, db_dsn, db_schema, status,
			config, settings, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	queryTenantSelect = `SELECT ` + tenantColumns + ` FROM tenants`

	queryTenantGet       = queryTenantSelect + ` WHERE id = $1`
	queryTenantGetByName = queryTenantSelect + ` WHERE name = $1`
	queryTenantList      = queryTenantSelect + ` WHERE deleted_at IS NULL ORDER BY created_at DESC`

	queryTenantUpdate = `
		UPDATE tenants
		SET name = $1, domain = $2, owner_id = $3, db_dsn = $4, db_schema = $5, status = $6,
		    config = $7, settings = $8, metadata = $9, updated_at = NOW()
		WHERE id = $10`

	queryTenantDelete = `UPDATE tenants SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	queryTenantExists = `SELECT 1 FROM tenants WHERE id = $1`
)

// @Service "pg-tenant-store"
type pgTenantStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.TenantStore].
func (p *pgTenantStore) Create(ctx context.Context, tenant *domaincore.Tenant) error {
	metadata, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return err
	}
	settings, err := json.Marshal(tenant.Settings)
	if err != nil {
		return err
	}
	config, err := json.Marshal(tenant.Config)
	if err != nil {
		return err
	}

	_, err = p.dbPool.Exec(ctx, queryTenantInsert,
		tenant.ID, tenant.Name, tenant.Domain, tenant.OwnerID, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		config, settings, metadata,
	)
	return err
}

// Delete implements [store.TenantStore].
func (p *pgTenantStore) Delete(ctx context.Context, tenantID string) error {
	_, err := p.dbPool.Exec(ctx, queryTenantDelete, tenantID)
	return err
}

// Exists implements [store.TenantStore].
func (p *pgTenantStore) Exists(ctx context.Context, tenantID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryTenantExists, tenantID)
}

// Get implements [store.TenantStore].
func (p *pgTenantStore) Get(ctx context.Context, tenantID string) (*domaincore.Tenant, error) {
	tenant := &domaincore.Tenant{}
	var metadata, settings, config []byte

	err := p.dbPool.QueryRow(ctx, queryTenantGet, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.OwnerID, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalTenantFields(tenant, metadata, settings, config); err != nil {
		return nil, err
	}

	return tenant, nil
}

// GetByName implements [store.TenantStore].
func (p *pgTenantStore) GetByName(ctx context.Context, name string) (*domaincore.Tenant, error) {
	tenant := &domaincore.Tenant{}
	var metadata, settings, config []byte

	err := p.dbPool.QueryRow(ctx, queryTenantGetByName, name).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.OwnerID, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("tenant not found with name: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalTenantFields(tenant, metadata, settings, config); err != nil {
		return nil, err
	}

	return tenant, nil
}

// List implements [store.TenantStore].
func (p *pgTenantStore) List(ctx context.Context) ([]*domaincore.Tenant, error) {
	rows, err := p.dbPool.Query(ctx, queryTenantList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanTenants(rows)
}

// Update implements [store.TenantStore].
func (p *pgTenantStore) Update(ctx context.Context, tenant *domaincore.Tenant) error {
	metadata, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return err
	}
	settings, err := json.Marshal(tenant.Settings)
	if err != nil {
		return err
	}
	config, err := json.Marshal(tenant.Config)
	if err != nil {
		return err
	}

	result, err := p.dbPool.Exec(ctx, queryTenantUpdate,
		tenant.Name, tenant.Domain, tenant.OwnerID, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		config, settings, metadata, tenant.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}
	return nil
}

var _ store.TenantStore = (*pgTenantStore)(nil)

// unmarshalTenantFields unmarshals JSON fields into tenant struct
func (p *pgTenantStore) unmarshalTenantFields(tenant *domaincore.Tenant, metadata, settings, config []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		tenant.Metadata = &m
	}
	if len(settings) > 0 {
		tenant.Settings = &domaincore.TenantSettings{}
		if err := json.Unmarshal(settings, tenant.Settings); err != nil {
			return fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}
	if len(config) > 0 {
		tenant.Config = &domaincore.TenantConfig{}
		if err := json.Unmarshal(config, tenant.Config); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return nil
}

func (p *pgTenantStore) scanTenants(rows serviceapi.Rows) ([]*domaincore.Tenant, error) {
	tenants := make([]*domaincore.Tenant, 0, 10)

	for rows.Next() {
		tenant := &domaincore.Tenant{}
		var metadata, settings, config []byte

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.OwnerID, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
			&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalTenantFields(tenant, metadata, settings, config); err != nil {
			return nil, err
		}

		tenants = append(tenants, tenant)
	}

	return tenants, rows.Err()
}
