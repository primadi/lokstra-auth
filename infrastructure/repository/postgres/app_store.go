package postgres

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-app-store"
type PostgresAppStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.AppStore = (*PostgresAppStore)(nil)

func (s *PostgresAppStore) Create(ctx context.Context, app *domain.App) error {
	metadata, _ := json.Marshal(app.Metadata)
	config, _ := json.Marshal(app.Config)

	query := `
		INSERT INTO apps (
			id, tenant_id, name, type, status, config, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.dbPool.Exec(ctx, query,
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

	err := s.dbPool.QueryRow(ctx, query, tenantID, appID).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status, &config, &metadata,
		&app.CreatedAt, &app.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
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

	result, err := s.dbPool.Exec(ctx, query,
		app.Name, app.Type, app.Status, config, metadata, app.UpdatedAt, app.TenantID, app.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("app not found: %s in tenant %s", app.ID, app.TenantID)
	}
	return nil
}

func (s *PostgresAppStore) Delete(ctx context.Context, tenantID, appID string) error {
	query := `DELETE FROM apps WHERE tenant_id = $1 AND id = $2`
	_, err := s.dbPool.Exec(ctx, query, tenantID, appID)
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
	rows, err := s.dbPool.Query(ctx, query, tenantID)
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

	app := &domain.App{}
	var metadata, config []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, name).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status, &config, &metadata,
		&app.CreatedAt, &app.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
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
	query := `SELECT 1 FROM apps WHERE tenant_id = $1 AND id = $2`
	return s.dbPool.IsExists(ctx, query, tenantID, appID)
}

func (s *PostgresAppStore) ListByType(ctx context.Context, tenantID string, appType domain.AppType) ([]*domain.App, error) {
	query := `
		SELECT id, tenant_id, name, type, status, config, metadata,
		       created_at, updated_at
		FROM apps
		WHERE tenant_id = $1 AND type = $2
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appType)
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
