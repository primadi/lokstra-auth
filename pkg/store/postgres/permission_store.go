package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for permission operations
const (
	permissionColumns = `id, tenant_id, app_id, name, description, resource, action, status, metadata, created_at, updated_at, deleted_at`

	queryPermissionInsert = `
		INSERT INTO permissions (
			id, tenant_id, app_id, name, description, resource, action, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryPermissionSelect = `SELECT ` + permissionColumns + ` FROM permissions`

	queryPermissionGet       = queryPermissionSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND id = $3 AND deleted_at IS NULL`
	queryPermissionGetByName = queryPermissionSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND name = $3 AND deleted_at IS NULL`
	queryPermissionList      = queryPermissionSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND deleted_at IS NULL ORDER BY created_at DESC`

	queryPermissionUpdate = `
		UPDATE permissions
		SET name = $1, description = $2, resource = $3, action = $4, status = $5, metadata = $6, updated_at = NOW()
		WHERE tenant_id = $7 AND app_id = $8 AND id = $9 AND deleted_at IS NULL`

	queryPermissionDelete = `
		UPDATE permissions
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3 AND deleted_at IS NULL`
)

// @Service "pg-permission-store"
type pgPermissionStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.PermissionStore].
func (p *pgPermissionStore) Create(ctx context.Context, permission *rbacdomain.Permission) error {
	metadata, err := json.Marshal(permission.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryPermissionInsert,
		permission.ID, permission.TenantID, permission.AppID, permission.Name, permission.Description,
		permission.Resource, permission.Action, permission.Status, metadata,
	)
	return err
}

// Get implements [store.PermissionStore].
func (p *pgPermissionStore) Get(ctx context.Context, tenantID, appID, permissionID string) (*rbacdomain.Permission, error) {
	permission := &rbacdomain.Permission{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryPermissionGet, tenantID, appID, permissionID).Scan(
		&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name, &permission.Description,
		&permission.Resource, &permission.Action, &permission.Status, &metadata,
		&permission.CreatedAt, &permission.UpdatedAt, &permission.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("permission not found: tenant=%s, app=%s, permission=%s", tenantID, appID, permissionID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalPermissionFields(permission, metadata); err != nil {
		return nil, err
	}

	return permission, nil
}

// GetByName implements [store.PermissionStore].
func (p *pgPermissionStore) GetByName(ctx context.Context, tenantID, appID, name string) (*rbacdomain.Permission, error) {
	permission := &rbacdomain.Permission{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryPermissionGetByName, tenantID, appID, name).Scan(
		&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name, &permission.Description,
		&permission.Resource, &permission.Action, &permission.Status, &metadata,
		&permission.CreatedAt, &permission.UpdatedAt, &permission.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("permission not found: tenant=%s, app=%s, name=%s", tenantID, appID, name)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalPermissionFields(permission, metadata); err != nil {
		return nil, err
	}

	return permission, nil
}

// Update implements [store.PermissionStore].
func (p *pgPermissionStore) Update(ctx context.Context, permission *rbacdomain.Permission) error {
	metadata, err := json.Marshal(permission.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryPermissionUpdate,
		permission.Name, permission.Description, permission.Resource, permission.Action,
		permission.Status, metadata, permission.TenantID, permission.AppID, permission.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission not found: tenant=%s, app=%s, permission=%s", permission.TenantID, permission.AppID, permission.ID)
	}
	return nil
}

// Delete implements [store.PermissionStore].
func (p *pgPermissionStore) Delete(ctx context.Context, tenantID, appID, permissionID string) error {
	result, err := p.dbPool.Exec(ctx, queryPermissionDelete, tenantID, appID, permissionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission not found: tenant=%s, app=%s, permission=%s", tenantID, appID, permissionID)
	}
	return nil
}

// List implements [store.PermissionStore].
func (p *pgPermissionStore) List(ctx context.Context, tenantID, appID string) ([]*rbacdomain.Permission, error) {
	rows, err := p.dbPool.Query(ctx, queryPermissionList, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPermissions(rows)
}

var _ store.PermissionStore = (*pgPermissionStore)(nil)

// unmarshalPermissionFields unmarshals JSON fields into permission struct
func (p *pgPermissionStore) unmarshalPermissionFields(permission *rbacdomain.Permission, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		permission.Metadata = &m
	}
	return nil
}

func (p *pgPermissionStore) scanPermissions(rows serviceapi.Rows) ([]*rbacdomain.Permission, error) {
	permissions := make([]*rbacdomain.Permission, 0, 10)

	for rows.Next() {
		permission := &rbacdomain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name, &permission.Description,
			&permission.Resource, &permission.Action, &permission.Status, &metadata,
			&permission.CreatedAt, &permission.UpdatedAt, &permission.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalPermissionFields(permission, metadata); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, rows.Err()
}
