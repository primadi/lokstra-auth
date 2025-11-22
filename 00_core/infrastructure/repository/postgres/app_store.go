package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra-auth/00_core/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

type PostgresAppStore struct {
	dbPool serviceapi.DbPoolWithSchema
}

func NewAppStore(dbPool serviceapi.DbPoolWithSchema) *PostgresAppStore {
	return &PostgresAppStore{dbPool: dbPool}
}

func (s *PostgresAppStore) Create(ctx context.Context, app *domain.App) error {
	metadata, _ := json.Marshal(app.Metadata)
	config, _ := json.Marshal(app.Config)

	query := `
		INSERT INTO apps (
			id, tenant_id, name, type, status, config, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		app.ID, app.TenantID, app.Name, app.Type, app.Status, config, metadata,
		app.CreatedAt, app.UpdatedAt,
	)
	return err
}

func (s *PostgresAppStore) Get(ctx context.Context, tenantID, appID string) (*domain.App, error) {
	query := `
		SELECT id, tenant_id, name, type, status, config, metadata,
		       created_at, updated_at
		FROM apps
		WHERE tenant_id = $1 AND id = $2
	`

	app := &domain.App{}
	var metadata, config []byte
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	err = cn.QueryRow(ctx, query, tenantID, appID).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status, &config, &metadata,
		&app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("app not found: %s in tenant %s", appID, tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		app.Metadata = &m
	}
	if len(config) > 0 {
		app.Config = &domain.AppConfig{}
		json.Unmarshal(config, app.Config)
	}

	return app, nil
}

func (s *PostgresAppStore) Update(ctx context.Context, app *domain.App) error {
	metadata, _ := json.Marshal(app.Metadata)
	config, _ := json.Marshal(app.Config)

	query := `
		UPDATE apps
		SET name = $1, type = $2, status = $3, config = $4, metadata = $5, updated_at = $6
		WHERE tenant_id = $7 AND id = $8
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	result, err := cn.Exec(ctx, query,
		app.Name, app.Type, app.Status, config, metadata, app.UpdatedAt, app.TenantID, app.ID,
	)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("app not found: %s in tenant %s", app.ID, app.TenantID)
	}
	return nil
}

func (s *PostgresAppStore) Delete(ctx context.Context, tenantID, appID string) error {
	query := `DELETE FROM apps WHERE tenant_id = $1 AND id = $2`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, tenantID, appID)
	return err
}

func (s *PostgresAppStore) List(ctx context.Context, tenantID string) ([]*domain.App, error) {
	query := `
		SELECT id, tenant_id, name, type, status, config, metadata,
		       created_at, updated_at
		FROM apps
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanApps(rows)
}

func (s *PostgresAppStore) GetByName(ctx context.Context, tenantID, name string) (*domain.App, error) {
	query := `
		SELECT id, tenant_id, name, type, status, config, metadata,
		       created_at, updated_at
		FROM apps
		WHERE tenant_id = $1 AND name = $2
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	app := &domain.App{}
	var metadata, config []byte

	err = cn.QueryRow(ctx, query, tenantID, name).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status, &config, &metadata,
		&app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("app not found with name: %s in tenant %s", name, tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		app.Metadata = &m
	}
	if len(config) > 0 {
		app.Config = &domain.AppConfig{}
		json.Unmarshal(config, app.Config)
	}

	return app, nil
}

func (s *PostgresAppStore) Exists(ctx context.Context, tenantID, appID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM apps WHERE tenant_id = $1 AND id = $2)`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var exists bool
	err = cn.QueryRow(ctx, query, tenantID, appID).Scan(&exists)
	return exists, err
}

func (s *PostgresAppStore) ListByType(ctx context.Context, tenantID string, appType domain.AppType) ([]*domain.App, error) {
	query := `
		SELECT id, tenant_id, name, type, status, config, metadata,
		       created_at, updated_at
		FROM apps
		WHERE tenant_id = $1 AND type = $2
		ORDER BY created_at DESC
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanApps(rows)
}

func (s *PostgresAppStore) scanApps(rows serviceapi.Rows) ([]*domain.App, error) {
	apps := make([]*domain.App, 0)

	for rows.Next() {
		app := &domain.App{}
		var metadata, config []byte

		err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status, &config, &metadata,
			&app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			app.Metadata = &m
		}
		if len(config) > 0 {
			app.Config = &domain.AppConfig{}
			json.Unmarshal(config, app.Config)
		}

		apps = append(apps, app)
	}

	return apps, rows.Err()
}
