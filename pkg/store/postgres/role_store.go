package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for role operations
const (
	roleColumns = `id, tenant_id, app_id, name, description, status, metadata, created_at, updated_at, deleted_at`

	queryRoleInsert = `
		INSERT INTO roles (
			id, tenant_id, app_id, name, description, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryRoleSelect = `SELECT ` + roleColumns + ` FROM roles`

	queryRoleGet       = queryRoleSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND id = $3 AND deleted_at IS NULL`
	queryRoleGetByName = queryRoleSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND name = $3 AND deleted_at IS NULL`
	queryRoleList      = queryRoleSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND deleted_at IS NULL ORDER BY created_at DESC`

	queryRoleUpdate = `
		UPDATE roles
		SET name = $1, description = $2, status = $3, metadata = $4, updated_at = NOW()
		WHERE tenant_id = $5 AND app_id = $6 AND id = $7 AND deleted_at IS NULL`

	queryRoleDelete = `
		UPDATE roles
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3 AND deleted_at IS NULL`
)

// @Service "pg-role-store"
type pgRoleStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.RoleStore].
func (p *pgRoleStore) Create(ctx context.Context, role *rbacdomain.Role) error {
	metadata, err := json.Marshal(role.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryRoleInsert,
		role.ID, role.TenantID, role.AppID, role.Name, role.Description,
		role.Status, metadata,
	)
	return err
}

// Get implements [store.RoleStore].
func (p *pgRoleStore) Get(ctx context.Context, tenantID, appID, roleID string) (*rbacdomain.Role, error) {
	role := &rbacdomain.Role{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryRoleGet, tenantID, appID, roleID).Scan(
		&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description,
		&role.Status, &metadata, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("role not found: tenant=%s, app=%s, role=%s", tenantID, appID, roleID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalRoleFields(role, metadata); err != nil {
		return nil, err
	}

	return role, nil
}

// GetByName implements [store.RoleStore].
func (p *pgRoleStore) GetByName(ctx context.Context, tenantID, appID, name string) (*rbacdomain.Role, error) {
	role := &rbacdomain.Role{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryRoleGetByName, tenantID, appID, name).Scan(
		&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description,
		&role.Status, &metadata, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("role not found: tenant=%s, app=%s, name=%s", tenantID, appID, name)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalRoleFields(role, metadata); err != nil {
		return nil, err
	}

	return role, nil
}

// Update implements [store.RoleStore].
func (p *pgRoleStore) Update(ctx context.Context, role *rbacdomain.Role) error {
	metadata, err := json.Marshal(role.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryRoleUpdate,
		role.Name, role.Description, role.Status, metadata,
		role.TenantID, role.AppID, role.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found: tenant=%s, app=%s, role=%s", role.TenantID, role.AppID, role.ID)
	}
	return nil
}

// Delete implements [store.RoleStore].
func (p *pgRoleStore) Delete(ctx context.Context, tenantID, appID, roleID string) error {
	result, err := p.dbPool.Exec(ctx, queryRoleDelete, tenantID, appID, roleID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found: tenant=%s, app=%s, role=%s", tenantID, appID, roleID)
	}
	return nil
}

// List implements [store.RoleStore].
func (p *pgRoleStore) List(ctx context.Context, tenantID, appID string) ([]*rbacdomain.Role, error) {
	rows, err := p.dbPool.Query(ctx, queryRoleList, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanRoles(rows)
}

var _ store.RoleStore = (*pgRoleStore)(nil)

// unmarshalRoleFields unmarshals JSON fields into role struct
func (p *pgRoleStore) unmarshalRoleFields(role *rbacdomain.Role, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		role.Metadata = &m
	}
	return nil
}

func (p *pgRoleStore) scanRoles(rows serviceapi.Rows) ([]*rbacdomain.Role, error) {
	roles := make([]*rbacdomain.Role, 0, 10)

	for rows.Next() {
		role := &rbacdomain.Role{}
		var metadata []byte

		err := rows.Scan(
			&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description,
			&role.Status, &metadata, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalRoleFields(role, metadata); err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, rows.Err()
}
