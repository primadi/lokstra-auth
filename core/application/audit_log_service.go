package application

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// AuditLogService handles audit log operations
// @RouterService name="audit-log-service", prefix="${api-auth-prefix:/api/auth}/core/audit/logs", middlewares=["recovery", "request_logger", "auth"]
type AuditLogService struct {
	// @Inject "@store.audit-log-store"
	Store repository.AuditLogStore
}

// CreateAuditLog creates a new audit log entry
// @Route "POST /"
func (s *AuditLogService) CreateAuditLog(ctx *request.Context, req *domain.CreateAuditLogRequest) (*domain.AuditLog, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	log := &domain.AuditLog{
		TenantID:     req.TenantID,
		AppID:        req.AppID,
		UserID:       req.UserID,
		SessionID:    req.SessionID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Method:       req.Method,
		Path:         req.Path,
		StatusCode:   req.StatusCode,
		RequestBody:  req.RequestBody,
		ResponseBody: req.ResponseBody,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		Source:       req.Source,
		Success:      req.Success,
		ErrorMessage: req.ErrorMessage,
		Metadata:     req.Metadata,
	}

	if err := s.Store.Create(context.Background(), log); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return log, nil
}

// GetAuditLog retrieves an audit log by ID
// @Route "GET /{id}"
func (s *AuditLogService) GetAuditLog(ctx *request.Context, id int64) (*domain.AuditLog, error) {
	log, err := s.Store.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return log, nil
}

// ListAuditLogs lists audit logs with filters
// @Route "GET /"
func (s *AuditLogService) ListAuditLogs(ctx *request.Context, req *domain.ListAuditLogsRequest) (*ListAuditLogsResponse, error) {
	// Set default limit if not specified
	if req.Limit == 0 {
		req.Limit = 100
	}

	logs, err := s.Store.List(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	count, err := s.Store.Count(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return &ListAuditLogsResponse{
		Logs:   logs,
		Total:  count,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// ListAuditLogsResponse represents the response for listing audit logs
type ListAuditLogsResponse struct {
	Logs   []*domain.AuditLog `json:"logs"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// CleanupOldAuditLogs deletes old audit logs
// @Route "POST /cleanup"
func (s *AuditLogService) CleanupOldAuditLogs(ctx *request.Context, req *CleanupAuditLogsRequest) (*CleanupAuditLogsResponse, error) {
	if req.DaysToKeep <= 0 {
		req.DaysToKeep = 90 // Default to 90 days
	}

	deletedCount, err := s.Store.CleanupOld(context.Background(), req.DaysToKeep)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	return &CleanupAuditLogsResponse{
		DeletedCount: deletedCount,
		DaysKept:     req.DaysToKeep,
	}, nil
}

// CleanupAuditLogsRequest represents a request to cleanup old audit logs
type CleanupAuditLogsRequest struct {
	DaysToKeep int `json:"days_to_keep" validate:"min=1"`
}

// CleanupAuditLogsResponse represents the response for cleanup operation
type CleanupAuditLogsResponse struct {
	DeletedCount int `json:"deleted_count"`
	DaysKept     int `json:"days_kept"`
}

// LogAction is a helper to log an action (can be called from other services)
func (s *AuditLogService) LogAction(ctx context.Context, req *domain.CreateAuditLogRequest) error {
	log := &domain.AuditLog{
		TenantID:     req.TenantID,
		AppID:        req.AppID,
		UserID:       req.UserID,
		SessionID:    req.SessionID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Method:       req.Method,
		Path:         req.Path,
		StatusCode:   req.StatusCode,
		RequestBody:  req.RequestBody,
		ResponseBody: req.ResponseBody,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		Source:       req.Source,
		Success:      req.Success,
		ErrorMessage: req.ErrorMessage,
		Metadata:     req.Metadata,
	}

	return s.Store.Create(ctx, log)
}
