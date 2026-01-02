package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/rbac/domain"
	"github.com/primadi/lokstra/serviceapi"
)

// PermissionCompositionStore implements repository.PermissionCompositionStore for PostgreSQL
// @Service "postgres-permission-composition-store"
type PermissionCompositionStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

// Create creates a new permission composition
func (s *PermissionCompositionStore) Create(ctx context.Context, composition *domain.PermissionComposition) error {
	query := `
		INSERT INTO permission_compositions (
			parent_permission_id, child_permission_id,
			tenant_id, app_id, is_required, priority,
			metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadataJSON any
	if composition.Metadata != nil {
		data, _ := json.Marshal(composition.Metadata)
		metadataJSON = data
	}

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		composition.ParentPermissionID,
		composition.ChildPermissionID,
		composition.TenantID,
		composition.AppID,
		composition.IsRequired,
		composition.Priority,
		metadataJSON,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create permission composition: %w", err)
	}

	return nil
}

// Delete removes a child permission from compound permission
func (s *PermissionCompositionStore) Delete(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) error {
	query := `
		DELETE FROM permission_compositions
		WHERE parent_permission_id = $1
		AND child_permission_id = $2
		AND tenant_id = $3
		AND app_id = $4`

	result, err := s.dbPool.Exec(ctx, query, parentPermissionID, childPermissionID, tenantID, appID)
	if err != nil {
		return fmt.Errorf("failed to delete permission composition: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPermissionCompositionNotFound
	}

	return nil
}

// ListByParent lists all child permissions for a compound permission
func (s *PermissionCompositionStore) ListByParent(ctx context.Context, tenantID, appID, parentPermissionID string) ([]*domain.PermissionComposition, error) {
	query := `
		SELECT 
			parent_permission_id, child_permission_id,
			tenant_id, app_id, is_required, priority,
			metadata, created_at
		FROM permission_compositions
		WHERE parent_permission_id = $1
		AND tenant_id = $2
		AND app_id = $3
		ORDER BY priority ASC, created_at ASC`

	rows, err := s.dbPool.Query(ctx, query, parentPermissionID, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permission compositions: %w", err)
	}
	defer rows.Close()

	compositions := []*domain.PermissionComposition{}
	for rows.Next() {
		comp := &domain.PermissionComposition{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&comp.ParentPermissionID,
			&comp.ChildPermissionID,
			&comp.TenantID,
			&comp.AppID,
			&comp.IsRequired,
			&comp.Priority,
			&metadataJSON,
			&comp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission composition: %w", err)
		}

		if metadataJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(metadataJSON.String), &data); err == nil {
				comp.Metadata = &data
			}
		}

		compositions = append(compositions, comp)
	}

	return compositions, nil
}

// ListByChild finds all compound permissions that include a specific permission
func (s *PermissionCompositionStore) ListByChild(ctx context.Context, tenantID, appID, childPermissionID string) ([]*domain.PermissionComposition, error) {
	query := `
		SELECT 
			parent_permission_id, child_permission_id,
			tenant_id, app_id, is_required, priority,
			metadata, created_at
		FROM permission_compositions
		WHERE child_permission_id = $1
		AND tenant_id = $2
		AND app_id = $3
		ORDER BY priority ASC, created_at ASC`

	rows, err := s.dbPool.Query(ctx, query, childPermissionID, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permission compositions by child: %w", err)
	}
	defer rows.Close()

	compositions := []*domain.PermissionComposition{}
	for rows.Next() {
		comp := &domain.PermissionComposition{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&comp.ParentPermissionID,
			&comp.ChildPermissionID,
			&comp.TenantID,
			&comp.AppID,
			&comp.IsRequired,
			&comp.Priority,
			&metadataJSON,
			&comp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission composition: %w", err)
		}

		if metadataJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(metadataJSON.String), &data); err == nil {
				comp.Metadata = &data
			}
		}

		compositions = append(compositions, comp)
	}

	return compositions, nil
}

// GetEffectivePermissions recursively resolves all permissions
func (s *PermissionCompositionStore) GetEffectivePermissions(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	// Use recursive CTE to resolve all nested permissions
	query := `
		WITH RECURSIVE permission_tree AS (
			-- Base case: the permission itself
			SELECT $1::varchar AS permission_id, 0 AS level
			
			UNION
			
			-- Recursive case: child permissions
			SELECT 
				pc.child_permission_id AS permission_id,
				pt.level + 1 AS level
			FROM permission_tree pt
			JOIN permission_compositions pc 
				ON pc.parent_permission_id = pt.permission_id
				AND pc.tenant_id = $2
				AND pc.app_id = $3
			WHERE pt.level < 10  -- Prevent infinite loops (max depth 10)
		)
		SELECT DISTINCT permission_id
		FROM permission_tree
		ORDER BY permission_id`

	rows, err := s.dbPool.Query(ctx, query, permissionID, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective permissions: %w", err)
	}
	defer rows.Close()

	permissions := []string{}
	for rows.Next() {
		var permID string
		if err := rows.Scan(&permID); err != nil {
			return nil, fmt.Errorf("failed to scan permission ID: %w", err)
		}
		permissions = append(permissions, permID)
	}

	return permissions, nil
}

// Exists checks if a composition exists
func (s *PermissionCompositionStore) Exists(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM permission_compositions
			WHERE parent_permission_id = $1
			AND child_permission_id = $2
			AND tenant_id = $3
			AND app_id = $4
		)`

	var exists bool
	err := s.dbPool.QueryRow(ctx, query, parentPermissionID, childPermissionID, tenantID, appID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission composition existence: %w", err)
	}

	return exists, nil
}

// HasCircularDependency checks if adding this composition would create a circular dependency
func (s *PermissionCompositionStore) HasCircularDependency(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error) {
	// Check if childPermissionID is already an ancestor of parentPermissionID
	// This would create a cycle: parent -> ... -> child -> parent
	query := `
		WITH RECURSIVE permission_tree AS (
			-- Start from the proposed child
			SELECT $1::varchar AS permission_id, 0 AS level
			
			UNION
			
			-- Traverse up the tree (find parents)
			SELECT 
				pc.parent_permission_id AS permission_id,
				pt.level + 1 AS level
			FROM permission_tree pt
			JOIN permission_compositions pc 
				ON pc.child_permission_id = pt.permission_id
				AND pc.tenant_id = $2
				AND pc.app_id = $3
			WHERE pt.level < 10  -- Prevent infinite loops
		)
		SELECT EXISTS(
			SELECT 1 FROM permission_tree
			WHERE permission_id = $4
		)`

	var hasCircular bool
	err := s.dbPool.QueryRow(ctx, query, childPermissionID, tenantID, appID, parentPermissionID).Scan(&hasCircular)
	if err != nil {
		return false, fmt.Errorf("failed to check circular dependency: %w", err)
	}

	return hasCircular, nil
}
