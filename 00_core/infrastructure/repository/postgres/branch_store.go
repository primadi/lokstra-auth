package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

type PostgresBranchStore struct {
	dbPool     serviceapi.DbPoolWithSchema
	schemaName string
}

func NewBranchStore(dbPool serviceapi.DbPoolWithSchema) *PostgresBranchStore {
	return &PostgresBranchStore{dbPool: dbPool}
}

func (s *PostgresBranchStore) Create(ctx context.Context, branch *domain.Branch) error {
	metadata, _ := json.Marshal(branch.Metadata)
	settings, _ := json.Marshal(branch.Settings)
	address, _ := json.Marshal(branch.Address)
	contact, _ := json.Marshal(branch.Contact)

	query := `
		INSERT INTO branches (
			id, tenant_id, app_id, code, name, type, status,
			address, contact, settings, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		branch.ID, branch.TenantID, branch.AppID, branch.Code, branch.Name, branch.Type, branch.Status,
		address, contact, settings, metadata, branch.CreatedAt, branch.UpdatedAt,
	)

	return err
}

func (s *PostgresBranchStore) Get(ctx context.Context, tenantID, appID, branchID string) (*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, code, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	branch := &domain.Branch{}
	var metadata, settings, address, contact []byte

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	err = cn.QueryRow(ctx, query, tenantID, appID, branchID).Scan(
		&branch.ID, &branch.TenantID, &branch.AppID, &branch.Code, &branch.Name, &branch.Type, &branch.Status,
		&address, &contact, &settings, &metadata, &branch.CreatedAt, &branch.UpdatedAt,
	)

	if err == sql.ErrNoRows {
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
		SET code = $1, name = $2, type = $3, status = $4,
		    address = $5, contact = $6, settings = $7, metadata = $8, updated_at = $9
		WHERE tenant_id = $10 AND app_id = $11 AND id = $12
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	result, err := cn.Exec(ctx, query,
		branch.Code, branch.Name, branch.Type, branch.Status,
		address, contact, settings, metadata, branch.UpdatedAt,
		branch.TenantID, branch.AppID, branch.ID,
	)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("branch not found: %s in app %s", branch.ID, branch.AppID)
	}
	return nil
}

func (s *PostgresBranchStore) Delete(ctx context.Context, tenantID, appID, branchID string) error {
	query := `DELETE FROM branches WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, tenantID, appID, branchID)
	return err
}

func (s *PostgresBranchStore) List(ctx context.Context, tenantID, appID string) ([]*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, code, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
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

	return s.scanBranches(rows)
}

func (s *PostgresBranchStore) Exists(ctx context.Context, tenantID, appID, branchID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM branches WHERE tenant_id = $1 AND app_id = $2 AND id = $3)`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var exists bool
	err = cn.QueryRow(ctx, query, tenantID, appID, branchID).Scan(&exists)

	return exists, err
}

func (s *PostgresBranchStore) ListByType(ctx context.Context, tenantID, appID string, branchType domain.BranchType) ([]*domain.Branch, error) {
	query := `
		SELECT id, tenant_id, app_id, code, name, type, status,
		       address, contact, settings, metadata, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND app_id = $2 AND type = $3
		ORDER BY created_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, branchType)
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
			&branch.ID, &branch.TenantID, &branch.AppID, &branch.Code, &branch.Name, &branch.Type, &branch.Status,
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
