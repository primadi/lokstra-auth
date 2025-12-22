package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// AuditLogStore implements repository.AuditLogStore for PostgreSQL
// @Service "postgres-audit-log-store"
type AuditLogStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

// Create creates a new audit log entry
func (s *AuditLogStore) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			tenant_id, app_id, user_id, session_id,
			action, resource_type, resource_id,
			method, path, status_code,
			request_body, response_body,
			ip_address, user_agent, source,
			success, error_message, metadata,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19
		) RETURNING id`

	var requestBodyJSON, responseBodyJSON, metadataJSON any
	if log.RequestBody != nil {
		data, _ := json.Marshal(log.RequestBody)
		requestBodyJSON = data
	}
	if log.ResponseBody != nil {
		data, _ := json.Marshal(log.ResponseBody)
		responseBodyJSON = data
	}
	if log.Metadata != nil {
		data, _ := json.Marshal(log.Metadata)
		metadataJSON = data
	}

	err := s.dbPool.QueryRow(ctx, query,
		log.TenantID, log.AppID, log.UserID, log.SessionID,
		log.Action, log.ResourceType, log.ResourceID,
		log.Method, log.Path, log.StatusCode,
		requestBodyJSON, responseBodyJSON,
		log.IPAddress, log.UserAgent, log.Source,
		log.Success, log.ErrorMessage, metadataJSON,
		time.Now(),
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// Get retrieves an audit log by ID
func (s *AuditLogStore) Get(ctx context.Context, id int64) (*domain.AuditLog, error) {
	query := `
		SELECT 
			id, tenant_id, app_id, user_id, session_id,
			action, resource_type, resource_id,
			method, path, status_code,
			request_body, response_body,
			ip_address, user_agent, source,
			success, error_message, metadata,
			created_at
		FROM audit_logs
		WHERE id = $1`

	log := &domain.AuditLog{}
	var requestBodyJSON, responseBodyJSON, metadataJSON sql.NullString

	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.TenantID, &log.AppID, &log.UserID, &log.SessionID,
		&log.Action, &log.ResourceType, &log.ResourceID,
		&log.Method, &log.Path, &log.StatusCode,
		&requestBodyJSON, &responseBodyJSON,
		&log.IPAddress, &log.UserAgent, &log.Source,
		&log.Success, &log.ErrorMessage, &metadataJSON,
		&log.CreatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrAuditLogNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Parse JSONB fields
	if requestBodyJSON.Valid {
		var data map[string]any
		if err := json.Unmarshal([]byte(requestBodyJSON.String), &data); err == nil {
			log.RequestBody = &data
		}
	}
	if responseBodyJSON.Valid {
		var data map[string]any
		if err := json.Unmarshal([]byte(responseBodyJSON.String), &data); err == nil {
			log.ResponseBody = &data
		}
	}
	if metadataJSON.Valid {
		var data map[string]any
		if err := json.Unmarshal([]byte(metadataJSON.String), &data); err == nil {
			log.Metadata = &data
		}
	}

	return log, nil
}

// List lists audit logs with filters
func (s *AuditLogStore) List(ctx context.Context, req *domain.ListAuditLogsRequest) ([]*domain.AuditLog, error) {
	query := `
		SELECT 
			id, tenant_id, app_id, user_id, session_id,
			action, resource_type, resource_id,
			method, path, status_code,
			request_body, response_body,
			ip_address, user_agent, source,
			success, error_message, metadata,
			created_at
		FROM audit_logs
		WHERE 1=1`

	args := []any{}
	argPos := 1

	if req.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argPos)
		args = append(args, *req.TenantID)
		argPos++
	}
	if req.AppID != nil {
		query += fmt.Sprintf(" AND app_id = $%d", argPos)
		args = append(args, *req.AppID)
		argPos++
	}
	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *req.UserID)
		argPos++
	}
	if req.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, *req.Action)
		argPos++
	}
	if req.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argPos)
		args = append(args, *req.ResourceType)
		argPos++
	}
	if req.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argPos)
		args = append(args, *req.ResourceID)
		argPos++
	}
	if req.Source != nil {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, *req.Source)
		argPos++
	}
	if req.Success != nil {
		query += fmt.Sprintf(" AND success = $%d", argPos)
		args = append(args, *req.Success)
		argPos++
	}
	if req.FromDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *req.FromDate)
		argPos++
	}
	if req.ToDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *req.ToDate)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.Limit)
		argPos++
	}
	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, req.Offset)
		argPos++
	}

	rows, err := s.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	logs := []*domain.AuditLog{}
	for rows.Next() {
		log := &domain.AuditLog{}
		var requestBodyJSON, responseBodyJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&log.ID, &log.TenantID, &log.AppID, &log.UserID, &log.SessionID,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.Method, &log.Path, &log.StatusCode,
			&requestBodyJSON, &responseBodyJSON,
			&log.IPAddress, &log.UserAgent, &log.Source,
			&log.Success, &log.ErrorMessage, &metadataJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Parse JSONB fields (skip for list to reduce overhead)
		if metadataJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(metadataJSON.String), &data); err == nil {
				log.Metadata = &data
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// Count counts audit logs matching filters
func (s *AuditLogStore) Count(ctx context.Context, req *domain.ListAuditLogsRequest) (int64, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`

	args := []any{}
	argPos := 1

	if req.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argPos)
		args = append(args, *req.TenantID)
		argPos++
	}
	if req.AppID != nil {
		query += fmt.Sprintf(" AND app_id = $%d", argPos)
		args = append(args, *req.AppID)
		argPos++
	}
	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *req.UserID)
		argPos++
	}
	if req.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, *req.Action)
		argPos++
	}
	if req.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argPos)
		args = append(args, *req.ResourceType)
		argPos++
	}
	if req.Source != nil {
		query += fmt.Sprintf(" AND source = $%d", argPos)
		args = append(args, *req.Source)
		argPos++
	}
	if req.Success != nil {
		query += fmt.Sprintf(" AND success = $%d", argPos)
		args = append(args, *req.Success)
		argPos++
	}
	if req.FromDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *req.FromDate)
		argPos++
	}
	if req.ToDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *req.ToDate)
		argPos++
	}

	var count int64
	err := s.dbPool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// CleanupOld deletes audit logs older than specified days
func (s *AuditLogStore) CleanupOld(ctx context.Context, daysToKeep int) (int, error) {
	query := `SELECT cleanup_old_audit_logs($1)`

	var deletedCount int
	err := s.dbPool.QueryRow(ctx, query, daysToKeep).Scan(&deletedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	return deletedCount, nil
}
