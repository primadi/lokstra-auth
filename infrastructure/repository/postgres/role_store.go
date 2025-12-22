package postgres

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// PostgresRoleStore implements RoleStore with PostgreSQL
// @Service "postgres-role-store"
type PostgresRoleStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.RoleStore = (*PostgresRoleStore)(nil)

func (s *PostgresRoleStore) Create(ctx context.Context, role *domain.Role) error {
	metadata, _ := json.Marshal(role.Metadata)

	query := `
		INSERT INTO roles (
			id, tenant_id, app_id, name, description, status,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.dbPool.Exec(ctx, query,
		role.ID, role.TenantID, role.AppID, role.Name, role.Description, role.Status,
		metadata, role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (s *PostgresRoleStore) Get(ctx context.Context, tenantID, appID, roleID string) (*domain.Role, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, status,
		       metadata, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	role := &domain.Role{}
	var metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, appID, roleID).Scan(
		&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description, &role.Status,
		&metadata, &role.CreatedAt, &role.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		role.Metadata = &m
	}

	return role, nil
}

func (s *PostgresRoleStore) GetByName(ctx context.Context, tenantID, appID, name string) (*domain.Role, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, status,
		       metadata, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND app_id = $2 AND name = $3
	`

	role := &domain.Role{}
	var metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, appID, name).Scan(
		&role.ID, &role.TenantID, &role.AppID, &role.Name, &role.Description, &role.Status,
		&metadata, &role.CreatedAt, &role.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		role.Metadata = &m
	}

	return role, nil
}

func (s *PostgresRoleStore) Update(ctx context.Context, role *domain.Role) error {
	metadata, _ := json.Marshal(role.Metadata)

	query := `
		UPDATE roles
		SET name = $1, description = $2, status = $3,
		    metadata = $4, updated_at = $5
		WHERE tenant_id = $6 AND app_id = $7 AND id = $8
	`

	result, err := s.dbPool.Exec(ctx, query,
		role.Name, role.Description, role.Status,
		metadata, role.UpdatedAt, role.TenantID, role.AppID, role.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}
	return nil
}

func (s *PostgresRoleStore) Delete(ctx context.Context, tenantID, appID, roleID string) error {
	query := `DELETE FROM roles WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	_, err := s.dbPool.Exec(ctx, query, tenantID, appID, roleID)
	return err
}

func (s *PostgresRoleStore) List(ctx context.Context, tenantID, appID string) ([]*domain.Role, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, status,
		       metadata, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND app_id = $2
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID)
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

func (s *PostgresRoleStore) ListWithFilters(ctx context.Context, filters *domain.ListRolesRequest) ([]*domain.Role, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, status,
		       metadata, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND app_id = $2
	`
	args := []any{filters.TenantID, filters.AppID}
	argIdx := 3

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

	rows, err := s.dbPool.Query(ctx, query, args...)
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
