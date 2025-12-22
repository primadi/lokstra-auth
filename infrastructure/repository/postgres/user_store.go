package postgres

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-store"
type PostgresUserStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.UserStore = (*PostgresUserStore)(nil)

func (s *PostgresUserStore) Create(ctx context.Context, user *domain.User) error {
	metadata, _ := json.Marshal(user.Metadata)

	query := `
		INSERT INTO users (
			id, tenant_id, username, email, full_name, password_hash, status,
			metadata, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.dbPool.Exec(ctx, query,
		user.ID, user.TenantID, user.Username, user.Email, user.FullName, user.PasswordHash, user.Status,
		metadata, user.CreatedAt, user.UpdatedAt, user.DeletedAt,
	)
	return err
}

func (s *PostgresUserStore) Get(ctx context.Context,
	tenantID, userID string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, username, email, full_name, password_hash, status,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND id = $2
	`

	user := &domain.User{}
	var metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, userID).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName, &user.PasswordHash, &user.Status,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		user.Metadata = &m
	}

	return user, nil
}

func (s *PostgresUserStore) Update(ctx context.Context, user *domain.User) error {
	metadata, _ := json.Marshal(user.Metadata)

	query := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, password_hash = $4, status = $5,
		    metadata = $6, updated_at = $7, deleted_at = $8
		WHERE tenant_id = $9 AND id = $10
	`

	result, err := s.dbPool.Exec(ctx, query,
		user.Username, user.Email, user.FullName, user.PasswordHash, user.Status,
		metadata, user.UpdatedAt, user.DeletedAt, user.TenantID, user.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s in tenant %s", user.ID, user.TenantID)
	}
	return nil
}

func (s *PostgresUserStore) Delete(ctx context.Context, tenantID, userID string) error {
	query := `DELETE FROM users WHERE tenant_id = $1 AND id = $2`

	_, err := s.dbPool.Exec(ctx, query, tenantID, userID)
	return err
}

func (s *PostgresUserStore) List(ctx context.Context, tenantID string) ([]*domain.User, error) {
	query := `
		SELECT id, tenant_id, username, email, full_name, password_hash, status,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanUsers(rows)
}

func (s *PostgresUserStore) GetByUsername(ctx context.Context,
	tenantID, username string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, username, email, full_name, password_hash, status,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND username = $2
	`

	user := &domain.User{}
	var metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, username).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName, &user.PasswordHash, &user.Status,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found with username: %s in tenant %s", username, tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		user.Metadata = &m
	}

	return user, nil
}

func (s *PostgresUserStore) GetByEmail(ctx context.Context,
	tenantID, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, username, email, full_name, password_hash, status,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`

	user := &domain.User{}
	var metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, email).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName, &user.PasswordHash, &user.Status,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found with email: %s in tenant %s", email, tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		user.Metadata = &m
	}

	return user, nil
}

func (s *PostgresUserStore) Exists(ctx context.Context, tenantID, userID string) (bool, error) {
	query := `SELECT 1 FROM users WHERE tenant_id=$1 AND id=$2`
	return s.dbPool.IsExists(ctx, query, tenantID, userID)
}

func (s *PostgresUserStore) ListByApp(ctx context.Context,
	tenantID, appID string) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.username, u.email, u.full_name, u.password_hash, u.status,
		       u.metadata, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		INNER JOIN user_apps ua ON u.tenant_id = ua.tenant_id AND u.id = ua.user_id
		WHERE u.tenant_id = $1 AND ua.app_id = $2 AND ua.status = 'active'
		ORDER BY u.created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanUsers(rows)
}

func (s *PostgresUserStore) scanUsers(rows serviceapi.Rows) ([]*domain.User, error) {
	users := make([]*domain.User, 0)

	for rows.Next() {
		user := &domain.User{}
		var metadata []byte

		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName, &user.PasswordHash, &user.Status,
			&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			user.Metadata = &m
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

func (s *PostgresUserStore) SetPassword(ctx context.Context,
	tenantID, userID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`

	result, err := s.dbPool.Exec(ctx, query, passwordHash, tenantID, userID)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}

	return nil
}

func (s *PostgresUserStore) RemovePassword(ctx context.Context, tenantID, userID string) error {
	query := `
		UPDATE users
		SET password_hash = NULL, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`

	result, err := s.dbPool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}
	return nil
}
