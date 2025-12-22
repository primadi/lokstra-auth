package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// PostgresPermissionStore implements PermissionStore with PostgreSQL
// @Service "postgres-permission-store"
type PostgresPermissionStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.PermissionStore = (*PostgresPermissionStore)(nil)

func (s *PostgresPermissionStore) Create(ctx context.Context, permission *domain.Permission) error {
	metadata, _ := json.Marshal(permission.Metadata)

	query := `
		INSERT INTO permissions (
			id, tenant_id, app_id, name, description, resource, action, status,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		permission.ID, permission.TenantID, permission.AppID, permission.Name,
		permission.Description, permission.Resource, permission.Action, permission.Status,
		metadata, permission.CreatedAt, permission.UpdatedAt,
	)
	return err
}

func (s *PostgresPermissionStore) Get(ctx context.Context, tenantID, appID, permissionID string) (*domain.Permission, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, resource, action, status,
		       metadata, created_at, updated_at
		FROM permissions
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	permission := &domain.Permission{}
	var metadata []byte

	err = cn.QueryRow(ctx, query, tenantID, appID, permissionID).Scan(
		&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
		&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
		&metadata, &permission.CreatedAt, &permission.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		permission.Metadata = &m
	}

	return permission, nil
}

func (s *PostgresPermissionStore) GetByName(ctx context.Context, tenantID, appID, name string) (*domain.Permission, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, resource, action, status,
		       metadata, created_at, updated_at
		FROM permissions
		WHERE tenant_id = $1 AND app_id = $2 AND name = $3
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	permission := &domain.Permission{}
	var metadata []byte

	err = cn.QueryRow(ctx, query, tenantID, appID, name).Scan(
		&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
		&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
		&metadata, &permission.CreatedAt, &permission.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		permission.Metadata = &m
	}

	return permission, nil
}

func (s *PostgresPermissionStore) Update(ctx context.Context, permission *domain.Permission) error {
	metadata, _ := json.Marshal(permission.Metadata)

	query := `
		UPDATE permissions
		SET name = $1, description = $2, resource = $3, action = $4, status = $5,
		    metadata = $6, updated_at = $7
		WHERE tenant_id = $8 AND app_id = $9 AND id = $10
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	result, err := cn.Exec(ctx, query,
		permission.Name, permission.Description, permission.Resource, permission.Action, permission.Status,
		metadata, permission.UpdatedAt, permission.TenantID, permission.AppID, permission.ID,
	)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPermissionNotFound
	}
	return nil
}

func (s *PostgresPermissionStore) Delete(ctx context.Context, tenantID, appID, permissionID string) error {
	query := `DELETE FROM permissions WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, tenantID, appID, permissionID)
	return err
}

func (s *PostgresPermissionStore) List(ctx context.Context, tenantID, appID string) ([]*domain.Permission, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, resource, action, status,
		       metadata, created_at, updated_at
		FROM permissions
		WHERE tenant_id = $1 AND app_id = $2
		ORDER BY created_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
			&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
			&metadata, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (s *PostgresPermissionStore) ListWithFilters(ctx context.Context, filters *domain.ListPermissionsRequest) ([]*domain.Permission, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, resource, action, status,
		       metadata, created_at, updated_at
		FROM permissions
		WHERE tenant_id = $1 AND app_id = $2
	`
	args := []any{filters.TenantID, filters.AppID}
	argIdx := 3

	if filters.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argIdx)
		args = append(args, *filters.Resource)
		argIdx++
	}

	if filters.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filters.Action)
		argIdx++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filters.Status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filters.Offset)
	}

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
			&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
			&metadata, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// ============================================================
// PostgreSQL RolePermission Store
// ============================================================

type PostgresRolePermissionStore struct {
	dbPool          serviceapi.DbPool
	permissionStore repository.PermissionStore
}

var _ repository.RolePermissionStore = (*PostgresRolePermissionStore)(nil)

func NewRolePermissionStore(cfg map[string]any, permissionStore repository.PermissionStore) *PostgresRolePermissionStore {
	dbPool := cfg["db_main"].(serviceapi.DbPool)
	return &PostgresRolePermissionStore{
		dbPool:          dbPool,
		permissionStore: permissionStore,
	}
}

func (s *PostgresRolePermissionStore) AssignPermission(ctx context.Context, rolePermission *domain.RolePermission) error {
	query := `
		INSERT INTO role_permissions (role_id, tenant_id, app_id, permission_id, granted_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (role_id, tenant_id, app_id, permission_id)
		DO UPDATE SET granted_at = $5, revoked_at = $6
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		rolePermission.RoleID, rolePermission.TenantID, rolePermission.AppID, rolePermission.PermissionID,
		rolePermission.GrantedAt, rolePermission.RevokedAt,
	)
	return err
}

func (s *PostgresRolePermissionStore) RevokePermission(ctx context.Context, tenantID, appID, roleID, permissionID string) error {
	query := `
		UPDATE role_permissions
		SET revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND role_id = $4 AND permission_id = $5
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, time.Now(), tenantID, appID, roleID, permissionID)
	return err
}

func (s *PostgresRolePermissionStore) ListRolePermissions(ctx context.Context, tenantID, appID, roleID string) ([]*domain.Permission, error) {
	query := `
		SELECT p.id, p.tenant_id, p.app_id, p.name, p.description, p.resource, p.action, p.status,
		       p.metadata, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.tenant_id = $1 AND rp.app_id = $2 AND rp.role_id = $3
		  AND rp.revoked_at IS NULL
		ORDER BY rp.granted_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
			&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
			&metadata, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (s *PostgresRolePermissionStore) ListPermissionRoles(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	query := `
		SELECT role_id
		FROM role_permissions
		WHERE tenant_id = $1 AND app_id = $2 AND permission_id = $3
		  AND revoked_at IS NULL
		ORDER BY granted_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, permissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}

func (s *PostgresRolePermissionStore) HasPermission(ctx context.Context, tenantID, appID, roleID, permissionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM role_permissions
			WHERE tenant_id = $1 AND app_id = $2 AND role_id = $3 AND permission_id = $4
			  AND revoked_at IS NULL
		)
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var exists bool
	err = cn.QueryRow(ctx, query, tenantID, appID, roleID, permissionID).Scan(&exists)
	return exists, err
}

// ============================================================
// PostgreSQL UserPermission Store
// ============================================================

type PostgresUserPermissionStore struct {
	dbPool          serviceapi.DbPool
	permissionStore repository.PermissionStore
}

var _ repository.UserPermissionStore = (*PostgresUserPermissionStore)(nil)

func NewUserPermissionStore(cfg map[string]any, permissionStore repository.PermissionStore) *PostgresUserPermissionStore {
	dbPool := cfg["db_main"].(serviceapi.DbPool)
	return &PostgresUserPermissionStore{
		dbPool:          dbPool,
		permissionStore: permissionStore,
	}
}

func (s *PostgresUserPermissionStore) AssignPermission(ctx context.Context, userPermission *domain.UserPermission) error {
	query := `
		INSERT INTO user_permissions (user_id, tenant_id, app_id, permission_id, granted_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, tenant_id, app_id, permission_id)
		DO UPDATE SET granted_at = $5, revoked_at = $6
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		userPermission.UserID, userPermission.TenantID, userPermission.AppID, userPermission.PermissionID,
		userPermission.GrantedAt, userPermission.RevokedAt,
	)
	return err
}

func (s *PostgresUserPermissionStore) RevokePermission(ctx context.Context, tenantID, appID, userID, permissionID string) error {
	query := `
		UPDATE user_permissions
		SET revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND user_id = $4 AND permission_id = $5
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, time.Now(), tenantID, appID, userID, permissionID)
	return err
}

func (s *PostgresUserPermissionStore) ListUserPermissions(ctx context.Context, tenantID, appID, userID string) ([]*domain.Permission, error) {
	query := `
		SELECT p.id, p.tenant_id, p.app_id, p.name, p.description, p.resource, p.action, p.status,
		       p.metadata, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN user_permissions up ON p.id = up.permission_id
		WHERE up.tenant_id = $1 AND up.app_id = $2 AND up.user_id = $3
		  AND up.revoked_at IS NULL
		ORDER BY up.granted_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
			&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
			&metadata, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (s *PostgresUserPermissionStore) ListUserPermissionsWithRoles(ctx context.Context, tenantID, appID, userID string) ([]*domain.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.tenant_id, p.app_id, p.name, p.description, p.resource, p.action, p.status,
		       p.metadata, p.created_at, p.updated_at
		FROM permissions p
		WHERE (p.tenant_id, p.app_id, p.id) IN (
			-- Direct user permissions
			SELECT tenant_id, app_id, permission_id
			FROM user_permissions
			WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3
			  AND revoked_at IS NULL
			
			UNION
			
			-- Permissions from user roles
			SELECT rp.tenant_id, rp.app_id, rp.permission_id
			FROM role_permissions rp
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.tenant_id = $1 AND ur.app_id = $2 AND ur.user_id = $3
			  AND ur.revoked_at IS NULL
			  AND rp.revoked_at IS NULL
		)
		ORDER BY p.name
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		var metadata []byte

		err := rows.Scan(
			&permission.ID, &permission.TenantID, &permission.AppID, &permission.Name,
			&permission.Description, &permission.Resource, &permission.Action, &permission.Status,
			&metadata, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			permission.Metadata = &m
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (s *PostgresUserPermissionStore) ListPermissionUsers(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM user_permissions
		WHERE tenant_id = $1 AND app_id = $2 AND permission_id = $3
		  AND revoked_at IS NULL
		ORDER BY granted_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, permissionID)
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

func (s *PostgresUserPermissionStore) HasPermission(ctx context.Context, tenantID, appID, userID, permissionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM user_permissions
			WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND permission_id = $4
			  AND revoked_at IS NULL
			
			UNION
			
			SELECT 1
			FROM role_permissions rp
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.tenant_id = $1 AND ur.app_id = $2 AND ur.user_id = $3
			  AND rp.permission_id = $4
			  AND ur.revoked_at IS NULL
			  AND rp.revoked_at IS NULL
		)
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var exists bool
	err = cn.QueryRow(ctx, query, tenantID, appID, userID, permissionID).Scan(&exists)
	return exists, err
}
