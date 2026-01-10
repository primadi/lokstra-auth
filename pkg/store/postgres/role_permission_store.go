package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for role_permission operations
const (
	queryRolePermissionInsert = `
		INSERT INTO role_permissions (
			role_id, tenant_id, app_id, permission_id
		) VALUES ($1, $2, $3, $4)`

	queryRolePermissionRevoke = `
		UPDATE role_permissions
		SET revoked_at = NOW()
		WHERE role_id = $1 AND tenant_id = $2 AND app_id = $3 AND permission_id = $4 AND revoked_at IS NULL`

	queryRolePermissionHasPermission = `
		SELECT 1 FROM role_permissions
		WHERE role_id = $1 AND tenant_id = $2 AND app_id = $3 AND permission_id = $4 AND revoked_at IS NULL`

	queryRolePermissionListByRole = `
		SELECT ` + permissionColumns + `
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1 AND rp.tenant_id = $2 AND rp.app_id = $3 AND rp.revoked_at IS NULL AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC`

	queryRolePermissionListRolesByPermission = `
		SELECT role_id
		FROM role_permissions
		WHERE permission_id = $1 AND tenant_id = $2 AND app_id = $3 AND revoked_at IS NULL
		ORDER BY granted_at DESC`
)

// @Service "pg-role-permission-store"
type pgRolePermissionStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// AssignPermission implements [store.RolePermissionStore].
func (p *pgRolePermissionStore) AssignPermission(ctx context.Context, rolePermission *rbacdomain.RolePermission) error {
	_, err := p.dbPool.Exec(ctx, queryRolePermissionInsert,
		rolePermission.RoleID, rolePermission.TenantID, rolePermission.AppID, rolePermission.PermissionID,
	)
	return err
}

// RevokePermission implements [store.RolePermissionStore].
func (p *pgRolePermissionStore) RevokePermission(ctx context.Context, tenantID, appID, roleID, permissionID string) error {
	result, err := p.dbPool.Exec(ctx, queryRolePermissionRevoke, roleID, tenantID, appID, permissionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role permission not found: tenant=%s, app=%s, role=%s, permission=%s", tenantID, appID, roleID, permissionID)
	}
	return nil
}

// ListRolePermissions implements [store.RolePermissionStore].
func (p *pgRolePermissionStore) ListRolePermissions(ctx context.Context, tenantID, appID, roleID string) ([]*rbacdomain.Permission, error) {
	rows, err := p.dbPool.Query(ctx, queryRolePermissionListByRole, roleID, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPermissions(rows)
}

// ListPermissionRoles implements [store.RolePermissionStore].
func (p *pgRolePermissionStore) ListPermissionRoles(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	rows, err := p.dbPool.Query(ctx, queryRolePermissionListRolesByPermission, permissionID, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roleIDs := make([]string, 0, 10)
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, rows.Err()
}

// HasPermission implements [store.RolePermissionStore].
func (p *pgRolePermissionStore) HasPermission(ctx context.Context, tenantID, appID, roleID, permissionID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryRolePermissionHasPermission, roleID, tenantID, appID, permissionID)
}

var _ store.RolePermissionStore = (*pgRolePermissionStore)(nil)

func (p *pgRolePermissionStore) scanPermissions(rows serviceapi.Rows) ([]*rbacdomain.Permission, error) {
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

		if len(metadata) > 0 {
			var m map[string]any
			if err := json.Unmarshal(metadata, &m); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, rows.Err()
}
