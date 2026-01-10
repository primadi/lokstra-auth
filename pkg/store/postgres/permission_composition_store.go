package postgres

import (
	"context"
	"fmt"

	rbacdomain "github.com/primadi/lokstra-auth/pkg/domain/rbac"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for permission_composition operations
const (
	permissionCompositionColumns = `parent_permission_id, child_permission_id, tenant_id, app_id, is_required, priority, metadata, created_at`

	queryPermissionCompositionInsert = `
		INSERT INTO permission_compositions (
			parent_permission_id, child_permission_id, tenant_id, app_id, is_required, priority, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryPermissionCompositionSelect = `SELECT ` + permissionCompositionColumns + ` FROM permission_compositions`

	queryPermissionCompositionDelete = `
		DELETE FROM permission_compositions
		WHERE tenant_id = $1 AND app_id = $2 AND parent_permission_id = $3 AND child_permission_id = $4`

	queryPermissionCompositionListByParent = queryPermissionCompositionSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND parent_permission_id = $3 ORDER BY priority ASC, created_at ASC`
	queryPermissionCompositionListByChild  = queryPermissionCompositionSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND child_permission_id = $3 ORDER BY priority ASC, created_at ASC`
	queryPermissionCompositionExists       = `SELECT 1 FROM permission_compositions WHERE tenant_id = $1 AND app_id = $2 AND parent_permission_id = $3 AND child_permission_id = $4`
)

// @Service "pg-permission-composition-store"
type pgPermissionCompositionStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) Create(ctx context.Context, composition *rbacdomain.PermissionComposition) error {
	metadata, err := json.Marshal(composition.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryPermissionCompositionInsert,
		composition.ParentPermissionID, composition.ChildPermissionID,
		composition.TenantID, composition.AppID, composition.IsRequired,
		composition.Priority, metadata,
	)
	return err
}

// Delete implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) Delete(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) error {
	result, err := p.dbPool.Exec(ctx, queryPermissionCompositionDelete, tenantID, appID, parentPermissionID, childPermissionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("composition not found: tenant=%s, app=%s, parent=%s, child=%s", tenantID, appID, parentPermissionID, childPermissionID)
	}
	return nil
}

// ListByParent implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) ListByParent(ctx context.Context, tenantID, appID, parentPermissionID string) ([]*rbacdomain.PermissionComposition, error) {
	rows, err := p.dbPool.Query(ctx, queryPermissionCompositionListByParent, tenantID, appID, parentPermissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanCompositions(rows)
}

// ListByChild implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) ListByChild(ctx context.Context, tenantID, appID, childPermissionID string) ([]*rbacdomain.PermissionComposition, error) {
	rows, err := p.dbPool.Query(ctx, queryPermissionCompositionListByChild, tenantID, appID, childPermissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanCompositions(rows)
}

// GetEffectivePermissions implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) GetEffectivePermissions(ctx context.Context, tenantID, appID, permissionID string) ([]string, error) {
	// Recursive query to get all child permissions
	query := `
		WITH RECURSIVE permission_tree AS (
			-- Base case: direct children
			SELECT child_permission_id
			FROM permission_compositions
			WHERE tenant_id = $1 AND app_id = $2 AND parent_permission_id = $3
			
			UNION
			
			-- Recursive case: children of children
			SELECT pc.child_permission_id
			FROM permission_compositions pc
			INNER JOIN permission_tree pt ON pc.parent_permission_id = pt.child_permission_id
			WHERE pc.tenant_id = $1 AND pc.app_id = $2
		)
		SELECT DISTINCT child_permission_id FROM permission_tree`

	rows, err := p.dbPool.Query(ctx, query, tenantID, appID, permissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissionIDs := make([]string, 0, 10)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		permissionIDs = append(permissionIDs, id)
	}

	return permissionIDs, rows.Err()
}

// Exists implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) Exists(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryPermissionCompositionExists, tenantID, appID, parentPermissionID, childPermissionID)
}

// HasCircularDependency implements [store.PermissionCompositionStore].
func (p *pgPermissionCompositionStore) HasCircularDependency(ctx context.Context, tenantID, appID, parentPermissionID, childPermissionID string) (bool, error) {
	// Check if adding this relationship would create a cycle
	// i.e., if childPermissionID is already an ancestor of parentPermissionID
	query := `
		WITH RECURSIVE permission_tree AS (
			-- Base case: check if child is parent
			SELECT parent_permission_id
			FROM permission_compositions
			WHERE tenant_id = $1 AND app_id = $2 AND child_permission_id = $3
			
			UNION
			
			-- Recursive case: check parents of parents
			SELECT pc.parent_permission_id
			FROM permission_compositions pc
			INNER JOIN permission_tree pt ON pc.child_permission_id = pt.parent_permission_id
			WHERE pc.tenant_id = $1 AND pc.app_id = $2
		)
		SELECT EXISTS(SELECT 1 FROM permission_tree WHERE parent_permission_id = $4)`

	var hasCircular bool
	err := p.dbPool.QueryRow(ctx, query, tenantID, appID, parentPermissionID, childPermissionID).Scan(&hasCircular)
	return hasCircular, err
}

var _ store.PermissionCompositionStore = (*pgPermissionCompositionStore)(nil)

func (p *pgPermissionCompositionStore) scanCompositions(rows serviceapi.Rows) ([]*rbacdomain.PermissionComposition, error) {
	compositions := make([]*rbacdomain.PermissionComposition, 0, 10)

	for rows.Next() {
		composition := &rbacdomain.PermissionComposition{}
		var metadata []byte

		err := rows.Scan(
			&composition.ParentPermissionID, &composition.ChildPermissionID,
			&composition.TenantID, &composition.AppID, &composition.IsRequired,
			&composition.Priority, &metadata, &composition.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			if err := json.Unmarshal(metadata, &m); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			composition.Metadata = &m
		}

		compositions = append(compositions, composition)
	}

	return compositions, rows.Err()
}
