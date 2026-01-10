package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for user_role operations
const (
	queryUserRoleInsert = `
		INSERT INTO user_roles (
			user_id, tenant_id, app_id, role_id
		) VALUES ($1, $2, $3, $4)`

	queryUserRoleRevoke = `
		UPDATE user_roles
		SET revoked_at = NOW()
		WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND role_id = $4 AND revoked_at IS NULL`

	queryUserRoleHasRole = `SELECT 1 FROM user_roles WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND role_id = $4 AND revoked_at IS NULL`

	queryUserRoleListRolesByUser = `
		SELECT ` + roleColumns + `
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.tenant_id = $1 AND ur.app_id = $2 AND ur.user_id = $3 AND ur.revoked_at IS NULL AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`

	queryUserRoleListUsersByRole = `
		SELECT user_id
		FROM user_roles
		WHERE tenant_id = $1 AND app_id = $2 AND role_id = $3 AND revoked_at IS NULL
		ORDER BY granted_at DESC`
)

// @Service "pg-user-role-store"
type pgUserRoleStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// AssignRole implements [store.UserRoleStore].
func (p *pgUserRoleStore) AssignRole(ctx context.Context, userRole *rbacdomain.UserRole) error {
	_, err := p.dbPool.Exec(ctx, queryUserRoleInsert,
		userRole.UserID, userRole.TenantID, userRole.AppID, userRole.RoleID,
	)
	return err
}

// RevokeRole implements [store.UserRoleStore].
func (p *pgUserRoleStore) RevokeRole(ctx context.Context, tenantID, appID, userID, roleID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserRoleRevoke, tenantID, appID, userID, roleID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user role not found: tenant=%s, app=%s, user=%s, role=%s", tenantID, appID, userID, roleID)
	}
	return nil
}

// ListUserRoles implements [store.UserRoleStore].
func (p *pgUserRoleStore) ListUserRoles(ctx context.Context, tenantID, appID, userID string) ([]*rbacdomain.Role, error) {
	rows, err := p.dbPool.Query(ctx, queryUserRoleListRolesByUser, tenantID, appID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanRoles(rows)
}

// ListRoleUsers implements [store.UserRoleStore].
func (p *pgUserRoleStore) ListRoleUsers(ctx context.Context, tenantID, appID, roleID string) ([]string, error) {
	rows, err := p.dbPool.Query(ctx, queryUserRoleListUsersByRole, tenantID, appID, roleID)
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

// HasRole implements [store.UserRoleStore].
func (p *pgUserRoleStore) HasRole(ctx context.Context, tenantID, appID, userID, roleID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryUserRoleHasRole, tenantID, appID, userID, roleID)
}

var _ store.UserRoleStore = (*pgUserRoleStore)(nil)

func (p *pgUserRoleStore) scanRoles(rows serviceapi.Rows) ([]*rbacdomain.Role, error) {
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

		if len(metadata) > 0 {
			var m map[string]any
			if err := json.Unmarshal(metadata, &m); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			role.Metadata = &m
		}

		roles = append(roles, role)
	}

	return roles, rows.Err()
}
