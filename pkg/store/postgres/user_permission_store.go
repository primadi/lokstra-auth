package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for user_permission operations
const (
	queryUserPermissionInsert = `
		INSERT INTO user_permissions (
			user_id, tenant_id, app_id, permission_id
		) VALUES ($1, $2, $3, $4)`

	queryUserPermissionRevoke = `
		UPDATE user_permissions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND tenant_id = $2 AND app_id = $3 AND permission_id = $4 AND revoked_at IS NULL`

	queryUserPermissionListByUser = `
		SELECT ` + permissionColumns + `
		FROM permissions p
		INNER JOIN user_permissions up ON up.permission_id = p.id
		WHERE up.user_id = $1 AND up.tenant_id = $2 AND up.app_id = $3 AND up.revoked_at IS NULL AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC`

	queryUserPermissionListByUserWithRoles = `
		SELECT DISTINCT ` + permissionColumns + `
		FROM permissions p
		WHERE p.tenant_id = $2 AND p.app_id = $3 AND p.deleted_at IS NULL
		AND (
			-- Direct permissions
			EXISTS (
				SELECT 1 FROM user_permissions up
				WHERE up.permission_id = p.id AND up.user_id = $1 AND up.tenant_id = $2 AND up.app_id = $3 AND up.revoked_at IS NULL
			)
			OR
			-- Permissions via roles
			EXISTS (
				SELECT 1 FROM user_roles ur
				INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
				WHERE rp.permission_id = p.id AND ur.user_id = $1 AND ur.tenant_id = $2 AND ur.app_id = $3 
				AND ur.revoked_at IS NULL AND rp.revoked_at IS NULL
			)
		)
		ORDER BY p.created_at DESC`

	queryUserPermissionListUsersByPermission = `
		SELECT user_id
		FROM user_permissions
		WHERE permission_id = $1 AND tenant_id = $2 AND app_id = $3 AND revoked_at IS NULL
		ORDER BY granted_at DESC`

	queryUserPermissionHasPermission = `
		SELECT EXISTS (
			SELECT 1 FROM permissions p
			WHERE p.id = $4 AND p.tenant_id = $2 AND p.app_id = $3 AND p.deleted_at IS NULL
			AND (
				-- Direct permission
				EXISTS (
					SELECT 1 FROM user_permissions up
					WHERE up.permission_id = p.id AND up.user_id = $1 AND up.tenant_id = $2 AND up.app_id = $3 AND up.revoked_at IS NULL
				)
				OR
				-- Permission via role
				EXISTS (
					SELECT 1 FROM user_roles ur
					INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
					WHERE rp.permission_id = p.id AND ur.user_id = $1 AND ur.tenant_id = $2 AND ur.app_id = $3
					AND ur.revoked_at IS NULL AND rp.revoked_at IS NULL
				)
			)
		)`
)

// @Service "pg-user-permission-store"
type pgUserPermissionStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// AssignPermission implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) AssignPermission(ctx context.Context, userPermission *rbacdomain.UserPermission) error {
	_, err := p.dbPool.Exec(ctx, queryUserPermissionInsert,
		userPermission.UserID, userPermission.TenantID, userPermission.AppID, userPermission.PermissionID,
	)
	return err
}

// RevokePermission implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) RevokePermission(ctx context.Context, tenantID, appID, userID, permissionID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserPermissionRevoke, userID, tenantID, appID, permissionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user permission not found: tenant=%s, app=%s, user=%s, permission=%s", tenantID, appID, userID, permissionID)
	}
	return nil
}

// ListUserPermissions implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) ListUserPermissions(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Permission, error) {
	rows, err := p.dbPool.Query(ctx, queryUserPermissionListByUser, userID, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPermissions(rows)
}

// ListUserPermissionsWithRoles implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) ListUserPermissionsWithRoles(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Permission, error) {
	rows, err := p.dbPool.Query(ctx, queryUserPermissionListByUserWithRoles, userID, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPermissions(rows)
}

// ListPermissionUsers implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) ListPermissionUsers(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	rows, err := p.dbPool.Query(ctx, queryUserPermissionListUsersByPermission, permissionID, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0, 10)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// HasPermission implements [store.UserPermissionStore].
func (p *pgUserPermissionStore) HasPermission(ctx context.Context, tenantID, appID, userID, permissionID string) (bool, error) {
	var hasPermission bool
	err := p.dbPool.QueryRow(ctx, queryUserPermissionHasPermission, userID, tenantID, appID, permissionID).Scan(&hasPermission)
	return hasPermission, err
}

var _ store.UserPermissionStore = (*pgUserPermissionStore)(nil)

func (p *pgUserPermissionStore) scanPermissions(rows serviceapi.Rows) ([]*rbacdomain.Permission, error) {
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
