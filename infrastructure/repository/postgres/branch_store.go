package postgres

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-branch-store"
type PostgresBranchStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.BranchStore = (*PostgresBranchStore)(nil)

func (s *PostgresBranchStore) Create(ctx context.Context, branch *domain.Branch) error {
	metadata, _ := json.Marshal(branch.Metadata)
	settings, _ := json.Marshal(branch.Settings)
	address, _ := json.Marshal(branch.Address)
	contact, _ := json.Marshal(branch.Contact)

	query := `
		INSERT INTO branches (
			id, tenant_id, app_id, name, type, status,
			address, contact, settings, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.dbPool.Exec(ctx, query,
		branch.ID, branch.TenantID, branch.AppID, branch.Name, branch.Type, branch.Status,
		address, contact, settings, metadata, branch.CreatedAt, branch.UpdatedAt,
	)

	return err
}

func (s *PostgresBranchStore) Get(ctx context.Context, tenantID, appID, branchID string) (*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	branch := &domain.Branch{}
	var metadata, settings, address, contact []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, appID, branchID).Scan(
		&branch.ID, &branch.TenantID, &branch.AppID, &branch.Name, &branch.Type, &branch.Status,
		&address, &contact, &settings, &metadata, &branch.CreatedAt, &branch.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("branch not found: %s in app %s", branchID, appID)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(address, &branch.Address)
	json.Unmarshal(contact, &branch.Contact)
	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		branch.Metadata = &m
	}
	if len(settings) > 0 {
		branch.Settings = &domain.BranchSettings{}
		json.Unmarshal(settings, branch.Settings)
	}

	return branch, nil
}

func (s *PostgresBranchStore) Update(ctx context.Context, branch *domain.Branch) error {
	metadata, _ := json.Marshal(branch.Metadata)
	settings, _ := json.Marshal(branch.Settings)
	address, _ := json.Marshal(branch.Address)
	contact, _ := json.Marshal(branch.Contact)

	query := `
		UPDATE branches
		SET name = $1, type = $2, status = $3,
		    address = $4, contact = $5, settings = $6, metadata = $7, updated_at = $8
		WHERE tenant_id = $9 AND app_id = $10 AND id = $11
	`

	result, err := s.dbPool.Exec(ctx, query,
		branch.Name, branch.Type, branch.Status,
		address, contact, settings, metadata, branch.UpdatedAt,
		branch.TenantID, branch.AppID, branch.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("branch not found: %s in app %s", branch.ID, branch.AppID)
	}
	return nil
}

func (s *PostgresBranchStore) Delete(ctx context.Context, tenantID, appID, branchID string) error {
	query := `DELETE FROM branches WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	_, err := s.dbPool.Exec(ctx, query, tenantID, appID, branchID)
	return err
}

func (s *PostgresBranchStore) List(ctx context.Context, tenantID, appID string) ([]*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND app_id = $2
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanBranches(rows)
}

func (s *PostgresBranchStore) Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error) {
	query := `SELECT 1 FROM branches WHERE tenant_id = $1 AND app_id = $2 AND id = $3`
	return s.dbPool.IsExists(ctx, query, tenantID, appID, branchID)
}

func (s *PostgresBranchStore) ListByType(ctx context.Context, tenantID, appID string, branchType domain.BranchType) ([]*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND app_id = $2 AND type = $3
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID, branchType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanBranches(rows)
}

func (s *PostgresBranchStore) scanBranches(rows serviceapi.Rows) ([]*domain.Branch, error) {
	branches := make([]*domain.Branch, 0)

	for rows.Next() {
		branch := &domain.Branch{}
		var metadata, settings, address, contact []byte

		err := rows.Scan(
			&branch.ID, &branch.TenantID, &branch.AppID, &branch.Name, &branch.Type, &branch.Status,
			&address, &contact, &settings, &metadata, &branch.CreatedAt, &branch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(address, &branch.Address)
		json.Unmarshal(contact, &branch.Contact)
		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			branch.Metadata = &m
		}
		if len(settings) > 0 {
			branch.Settings = &domain.BranchSettings{}
			json.Unmarshal(settings, branch.Settings)
		}

		branches = append(branches, branch)
	}

	return branches, rows.Err()
}
