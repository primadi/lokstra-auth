package services

import (
	authmiddleware "github.com/primadi/lokstra-auth/middleware"
	"github.com/primadi/lokstra/core/request"
)

// DemoService provides demo endpoints for testing authentication and authorization
// @RouterService name="demo-service", prefix="/api/demo"
type DemoService struct{}

// ============================================================================
// PUBLIC ENDPOINTS (No Authentication Required)
// ============================================================================

// GetPublic handles public endpoint (no auth required)
// @Route "GET /public"
func (s *DemoService) GetPublic(ctx *request.Context) (map[string]any, error) {
	return map[string]any{
		"message": "This is a public endpoint, no authentication required",
		"public":  true,
	}, nil
}

// ============================================================================
// PROTECTED ENDPOINTS (Authentication Required)
// ============================================================================

// ProtectedService provides protected endpoints requiring authentication
// @RouterService name="protected-service", prefix="/api/protected", middlewares=["recovery", "request_logger", "auth"]
type ProtectedService struct{}

// GetInfo returns authenticated user information
// @Route "GET /info"
func (s *ProtectedService) GetInfo(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)

	return map[string]any{
		"message":     "This endpoint requires authentication",
		"user_id":     identity.Subject.ID,
		"tenant_id":   identity.TenantID,
		"app_id":      identity.AppID,
		"roles":       identity.Roles,
		"permissions": identity.Permissions,
	}, nil
}

// GetProfile returns user profile with complete identity
// @Route "GET /profile"
func (s *ProtectedService) GetProfile(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)

	return map[string]any{
		"user_id":     identity.Subject.ID,
		"tenant_id":   identity.TenantID,
		"app_id":      identity.AppID,
		"username":    identity.Subject.Principal,
		"email":       identity.Subject.Attributes["email"],
		"roles":       identity.Roles,
		"permissions": identity.Permissions,
		"groups":      identity.Groups,
		"profile":     identity.Profile,
	}, nil
}

// ============================================================================
// ADMIN ENDPOINTS (Requires Admin Role)
// ============================================================================

// AdminService provides admin-only endpoints
// @RouterService name="admin-service", prefix="/api/admin", middlewares=["recovery", "request_logger"]
type AdminService struct{}

// GetDashboard returns admin dashboard information
// @Route "GET /dashboard", ["auth", "role role=admin"]
func (s *AdminService) GetDashboard(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)

	return map[string]any{
		"message": "Welcome, admin!",
		"user_id": identity.Subject.ID,
		"roles":   identity.Roles,
	}, nil
}

// ============================================================================
// EDITOR ENDPOINTS (Requires Editor or Admin Role)
// ============================================================================

// EditorService provides editor endpoints
// @RouterService name="editor-service", prefix="/api/editor", middlewares=["recovery", "request_logger"]
type EditorService struct{}

// GetContent returns editor content
// @Route "GET /content", ["auth", "any-role roles=editor,admin"]
func (s *EditorService) GetContent(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)

	return map[string]any{
		"message": "You have editor access",
		"user_id": identity.Subject.ID,
		"roles":   identity.Roles,
	}, nil
}

// ============================================================================
// DOCUMENT ENDPOINTS (Permission-Based Access Control)
// ============================================================================

// DocumentService provides document management endpoints with permission checks
// @RouterService name="document-service", prefix="/api/documents", middlewares=["recovery", "request_logger"]
type DocumentService struct{}

// GetDocument reads a document (requires document:read permission)
// @Route "GET /{id}", ["auth", "permission permission=document:read"]
func (s *DocumentService) GetDocument(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)
	documentID := ctx.Req.PathParam("id", "")

	return map[string]any{
		"message":     "Document read access granted",
		"document_id": documentID,
		"user_id":     identity.Subject.ID,
		"permissions": identity.Permissions,
	}, nil
}

// UpdateDocument writes a document (requires document:write permission)
// @Route "POST /{id}", ["auth", "permission permission=document:write"]
func (s *DocumentService) UpdateDocument(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)
	documentID := ctx.Req.PathParam("id", "")

	return map[string]any{
		"message":     "Document write access granted",
		"document_id": documentID,
		"user_id":     identity.Subject.ID,
		"permissions": identity.Permissions,
	}, nil
}

// DeleteDocument deletes a document (requires document:delete permission)
// @Route "DELETE /{id}", ["auth", "permission permission=document:delete"]
func (s *DocumentService) DeleteDocument(ctx *request.Context) (map[string]any, error) {
	identity := authmiddleware.MustGetIdentity(ctx)
	documentID := ctx.Req.PathParam("id", "")

	return map[string]any{
		"message":     "Document delete access granted",
		"document_id": documentID,
		"user_id":     identity.Subject.ID,
		"permissions": identity.Permissions,
	}, nil
}
