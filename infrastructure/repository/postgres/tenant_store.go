package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-tenant-store"
type PostgresTenantStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.TenantStore = (*PostgresTenantStore)(nil)

func (s *PostgresTenantStore) Create(ctx context.Context, tenant *domain.Tenant) error {
	metadata, _ := json.Marshal(tenant.Metadata)
	settings, _ := json.Marshal(tenant.Settings)
	config, _ := json.Marshal(tenant.Config)

	query := `
		INSERT INTO tenants (
			id, name, domain, db_dsn, db_schema, status,
			config, settings, metadata, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.dbPool.Exec(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		config, settings, metadata, tenant.CreatedAt, tenant.UpdatedAt, tenant.DeletedAt,
	)
	return err
}

func (s *PostgresTenantStore) Get(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       config, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1
	`
	tenant := &domain.Tenant{}
	var metadata, settings, config []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		tenant.Metadata = &m
	}
	if len(settings) > 0 {
		tenant.Settings = &domain.TenantSettings{}
		json.Unmarshal(settings, tenant.Settings)
	}
	if len(config) > 0 {
		tenant.Config = &domain.TenantConfig{}
		json.Unmarshal(config, tenant.Config)
	}

	return tenant, nil
}

func (s *PostgresTenantStore) Update(ctx context.Context, tenant *domain.Tenant) error {
	metadata, _ := json.Marshal(tenant.Metadata)
	settings, _ := json.Marshal(tenant.Settings)
	config, _ := json.Marshal(tenant.Config)

	query := `
		UPDATE tenants
		SET name = $1, domain = $2, db_dsn = $3, db_schema = $4, status = $5,
		    config = $6, settings = $7, metadata = $8, updated_at = $9, deleted_at = $10
		WHERE id = $11
	`

	result, err := s.dbPool.Exec(ctx, query,
		tenant.Name, tenant.Domain, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		config, settings, metadata, tenant.UpdatedAt, tenant.DeletedAt, tenant.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}
	return nil
}

func (s *PostgresTenantStore) Delete(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenants WHERE id = $1`

	_, err := s.dbPool.Exec(ctx, query, tenantID)
	return err
}

func (s *PostgresTenantStore) List(ctx context.Context) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       config, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTenants(rows)
}

func (s *PostgresTenantStore) GetByName(ctx context.Context, name string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       config, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE name = $1
	`

	tenant := &domain.Tenant{}
	var metadata, settings, config []byte

	err := s.dbPool.QueryRow(ctx, query, name).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("tenant not found with name: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		tenant.Metadata = &m
	}
	if len(settings) > 0 {
		tenant.Settings = &domain.TenantSettings{}
		json.Unmarshal(settings, tenant.Settings)
	}
	if len(config) > 0 {
		tenant.Config = &domain.TenantConfig{}
		json.Unmarshal(config, tenant.Config)
	}

	return tenant, nil
}

func (s *PostgresTenantStore) Exists(ctx context.Context, tenantID string) (bool, error) {
	query := `SELECT SELECT 1 FROM tenants WHERE id = $1`
	return s.dbPool.IsExists(ctx, query, tenantID)
}

func (s *PostgresTenantStore) scanTenants(rows serviceapi.Rows) ([]*domain.Tenant, error) {
	tenants := make([]*domain.Tenant, 0)

	for rows.Next() {
		tenant := &domain.Tenant{}
		var metadata, settings, config []byte

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
			&config, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			tenant.Metadata = &m
		}
		if len(settings) > 0 {
			tenant.Settings = &domain.TenantSettings{}
			json.Unmarshal(settings, tenant.Settings)
		}
		if len(config) > 0 {
			tenant.Config = &domain.TenantConfig{}
			json.Unmarshal(config, tenant.Config)
		}

		tenants = append(tenants, tenant)
	}

	return tenants, rows.Err()
}
