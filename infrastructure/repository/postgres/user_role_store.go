package postgres

import (
	"context"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-role-store"
type PostgresUserRoleStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.UserRoleStore = (*PostgresUserRoleStore)(nil)

func (s *PostgresUserRoleStore) AssignRole(ctx context.Context, userRole *domain.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, tenant_id, app_id, role_id, granted_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, tenant_id, app_id, role_id)
		DO UPDATE SET granted_at = $5, revoked_at = $6
	`

	_, err := s.dbPool.Exec(ctx, query,
		userRole.UserID, userRole.TenantID, userRole.AppID, userRole.RoleID,
		userRole.GrantedAt, userRole.RevokedAt,
	)
	return err
}

func (s *PostgresUserRoleStore) RevokeRole(ctx context.Context, tenantID, appID, userID, roleID string) error {
	query := `
		UPDATE user_roles
		SET revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND user_id = $4 AND role_id = $5
	`

	_, err := s.dbPool.Exec(ctx, query, time.Now(), tenantID, appID, userID, roleID)
	return err
}

func (s *PostgresUserRoleStore) ListUserRoles(ctx context.Context, tenantID, appID, userID string) ([]*domain.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.app_id, r.name, r.description, r.status,
		       r.metadata, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.tenant_id = $1 AND ur.app_id = $2 AND ur.user_id = $3
		  AND ur.revoked_at IS NULL
		ORDER BY ur.granted_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		var metadata []byte

		err := rows.Scan(
			&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description, &role.Status,
			&metadata, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			role.Metadata = &m
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (s *PostgresUserRoleStore) ListRoleUsers(ctx context.Context, tenantID, appID, roleID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM user_roles
		WHERE tenant_id = $1 AND app_id = $2 AND role_id = $3
		  AND revoked_at IS NULL
		ORDER BY granted_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (s *PostgresUserRoleStore) HasRole(ctx context.Context, tenantID, appID, userID, roleID string) (bool, error) {
	query := `
		SELECT 1
		FROM user_roles
		WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND role_id = $4
			AND revoked_at IS NULL
	`

	return s.dbPool.IsExists(ctx, query, tenantID, appID, userID, roleID)
}
