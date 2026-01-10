package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for branch operations
const (
	branchColumns = `id, tenant_id, app_id, name, type, status, address, contact, settings, metadata, created_at, updated_at, deleted_at`

	queryBranchInsert = `
		INSERT INTO branches (
			id, tenant_id, app_id, name, type, status, address, contact, settings, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	queryBranchSelect = `SELECT ` + branchColumns + ` FROM branches`

	queryBranchGet        = queryBranchSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND id = $3`
	queryBranchList       = queryBranchSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryBranchListByType = queryBranchSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND type = $3 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryBranchExists     = `SELECT 1 FROM branches WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	queryBranchUpdate = `
		UPDATE branches
		SET name = $1, type = $2, status = $3, address = $4, contact = $5, settings = $6, metadata = $7, updated_at = NOW()
		WHERE tenant_id = $8 AND app_id = $9 AND id = $10`

	queryBranchDelete = `UPDATE branches SET deleted_at = NOW(), updated_at = NOW() WHERE tenant_id = $1 AND app_id = $2 AND id = $3`
)

// @Service "pg-branch-store"
type pgBranchStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.BranchStore].
func (p *pgBranchStore) Create(ctx context.Context, branch *domaincore.Branch) error {
	address, err := json.Marshal(branch.Address)
	if err != nil {
		return fmt.Errorf("marshal address: %w", err)
	}
	contact, err := json.Marshal(branch.Contact)
	if err != nil {
		return fmt.Errorf("marshal contact: %w", err)
	}
	settings, err := json.Marshal(branch.Settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	metadata, err := json.Marshal(branch.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryBranchInsert,
		branch.ID, branch.TenantID, branch.AppID, branch.Name, branch.Type, branch.Status,
		address, contact, settings, metadata,
	)
	return err
}

// Get implements [store.BranchStore].
func (p *pgBranchStore) Get(ctx context.Context, tenantID, appID, branchID string) (*domaincore.Branch, error) {
	branch := &domaincore.Branch{}
	var address, contact, settings, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryBranchGet, tenantID, appID, branchID).Scan(
		&branch.ID, &branch.TenantID, &branch.AppID, &branch.Name, &branch.Type, &branch.Status,
		&address, &contact, &settings, &metadata, &branch.CreatedAt, &branch.UpdatedAt, &branch.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("branch not found: tenant=%s, app=%s, branch=%s", tenantID, appID, branchID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalBranchFields(branch, address, contact, settings, metadata); err != nil {
		return nil, err
	}

	return branch, nil
}

// Update implements [store.BranchStore].
func (p *pgBranchStore) Update(ctx context.Context, branch *domaincore.Branch) error {
	address, err := json.Marshal(branch.Address)
	if err != nil {
		return fmt.Errorf("marshal address: %w", err)
	}
	contact, err := json.Marshal(branch.Contact)
	if err != nil {
		return fmt.Errorf("marshal contact: %w", err)
	}
	settings, err := json.Marshal(branch.Settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	metadata, err := json.Marshal(branch.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryBranchUpdate,
		branch.Name, branch.Type, branch.Status, address, contact, settings, metadata,
		branch.TenantID, branch.AppID, branch.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("branch not found: tenant=%s, app=%s, branch=%s", branch.TenantID, branch.AppID, branch.ID)
	}
	return nil
}

// Delete implements [store.BranchStore].
func (p *pgBranchStore) Delete(ctx context.Context, tenantID, appID, branchID string) error {
	result, err := p.dbPool.Exec(ctx, queryBranchDelete, tenantID, appID, branchID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("branch not found: tenant=%s, app=%s, branch=%s", tenantID, appID, branchID)
	}
	return nil
}

// List implements [store.BranchStore].
func (p *pgBranchStore) List(ctx context.Context, tenantID, appID string) ([]*domaincore.Branch, error) {
	rows, err := p.dbPool.Query(ctx, queryBranchList, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanBranches(rows)
}

// ListByType implements [store.BranchStore].
func (p *pgBranchStore) ListByType(ctx context.Context, tenantID, appID string, branchType domaincore.BranchType) ([]*domaincore.Branch, error) {
	rows, err := p.dbPool.Query(ctx, queryBranchListByType, tenantID, appID, branchType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanBranches(rows)
}

// Exists implements [store.BranchStore].
func (p *pgBranchStore) Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryBranchExists, tenantID, appID, branchID)
}

var _ store.BranchStore = (*pgBranchStore)(nil)

// unmarshalBranchFields unmarshals JSON fields into branch struct
func (p *pgBranchStore) unmarshalBranchFields(branch *domaincore.Branch, address, contact, settings, metadata []byte) error {
	if len(address) > 0 {
		branch.Address = &domaincore.BranchAddress{}
		if err := json.Unmarshal(address, branch.Address); err != nil {
			return fmt.Errorf("failed to unmarshal address: %w", err)
		}
	}
	if len(contact) > 0 {
		branch.Contact = &domaincore.BranchContact{}
		if err := json.Unmarshal(contact, branch.Contact); err != nil {
			return fmt.Errorf("failed to unmarshal contact: %w", err)
		}
	}
	if len(settings) > 0 {
		branch.Settings = &domaincore.BranchSettings{}
		if err := json.Unmarshal(settings, branch.Settings); err != nil {
			return fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		branch.Metadata = &m
	}
	return nil
}

func (p *pgBranchStore) scanBranches(rows serviceapi.Rows) ([]*domaincore.Branch, error) {
	branches := make([]*domaincore.Branch, 0, 10)

	for rows.Next() {
		branch := &domaincore.Branch{}
		var address, contact, settings, metadata []byte

		err := rows.Scan(
			&branch.ID, &branch.TenantID, &branch.AppID, &branch.Name, &branch.Type, &branch.Status,
			&address, &contact, &settings, &metadata, &branch.CreatedAt, &branch.UpdatedAt, &branch.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalBranchFields(branch, address, contact, settings, metadata); err != nil {
			return nil, err
		}

		branches = append(branches, branch)
	}

	return branches, rows.Err()
}
