package middleware

import (
	"strings"

	"github.com/primadi/lokstra-auth/authz/rbac"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Register registers all auth-related middlewares to Lokstra registry
func Register() {
	// Register auth middleware (no params)
	lokstra_registry.RegisterMiddlewareFactory("auth", createAuthMiddleware)

	// Register role middlewares
	lokstra_registry.RegisterMiddlewareFactory("role", createRoleMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("any-role", createAnyRoleMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("all-roles", createAllRolesMiddleware)

	// Register permission middlewares
	lokstra_registry.RegisterMiddlewareFactory("permission", createPermissionMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("any-permission", createAnyPermissionMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("all-permissions", createAllPermissionsMiddleware)
}

// createAuthMiddleware creates auth middleware (no params needed)
func createAuthMiddleware(params map[string]any) func(*request.Context) error {
	mw := NewAuthMiddleware(AuthMiddlewareConfig{})
	return mw.Handler()
}

// createRoleMiddleware creates single role check middleware
// Params: role=admin
func createRoleMiddleware(params map[string]any) func(*request.Context) error {
	role, _ := params["role"].(string)
	if role == "" {
		return forbiddenHandler("missing role parameter")
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)
		if !identity.HasRole(role) {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing required role: " + role,
			})
		}
		return c.Next()
	}
}

// createAnyRoleMiddleware creates any-role check middleware
// Params: roles=admin,editor,manager
func createAnyRoleMiddleware(params map[string]any) func(*request.Context) error {
	rolesStr, _ := params["roles"].(string)
	if rolesStr == "" {
		return forbiddenHandler("missing roles parameter")
	}

	roles := strings.Split(rolesStr, ",")
	for i := range roles {
		roles[i] = strings.TrimSpace(roles[i])
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)
		if !identity.HasAnyRole(roles...) {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing any of required roles: " + rolesStr,
			})
		}
		return c.Next()
	}
}

// createAllRolesMiddleware creates all-roles check middleware
// Params: roles=admin,manager
func createAllRolesMiddleware(params map[string]any) func(*request.Context) error {
	rolesStr, _ := params["roles"].(string)
	if rolesStr == "" {
		return forbiddenHandler("missing roles parameter")
	}

	roles := strings.Split(rolesStr, ",")
	for i := range roles {
		roles[i] = strings.TrimSpace(roles[i])
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)
		if !identity.HasAllRoles(roles...) {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing all required roles: " + rolesStr,
			})
		}
		return c.Next()
	}
}

// createPermissionMiddleware creates single permission check middleware
// Params: permission=document:read
func createPermissionMiddleware(params map[string]any) func(*request.Context) error {
	permission, _ := params["permission"].(string)
	if permission == "" {
		return forbiddenHandler("missing permission parameter")
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)

		// Check using RoleChecker for simple permission check
		checker := rbac.NewPermissionChecker()
		hasPermission, _ := checker.HasPermission(c, identity, permission)

		if !hasPermission {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing required permission: " + permission,
			})
		}
		return c.Next()
	}
}

// createAnyPermissionMiddleware creates any-permission check middleware
// Params: permissions=read,write
func createAnyPermissionMiddleware(params map[string]any) func(*request.Context) error {
	permsStr, _ := params["permissions"].(string)
	if permsStr == "" {
		return forbiddenHandler("missing permissions parameter")
	}

	permissions := strings.Split(permsStr, ",")
	for i := range permissions {
		permissions[i] = strings.TrimSpace(permissions[i])
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)

		checker := rbac.NewPermissionChecker()
		hasPermission, _ := checker.HasAnyPermission(c, identity, permissions...)

		if !hasPermission {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing any of required permissions: " + permsStr,
			})
		}
		return c.Next()
	}
}

// createAllPermissionsMiddleware creates all-permissions check middleware
// Params: permissions=read,write,delete
func createAllPermissionsMiddleware(params map[string]any) func(*request.Context) error {
	permsStr, _ := params["permissions"].(string)
	if permsStr == "" {
		return forbiddenHandler("missing permissions parameter")
	}

	permissions := strings.Split(permsStr, ",")
	for i := range permissions {
		permissions[i] = strings.TrimSpace(permissions[i])
	}

	return func(c *request.Context) error {
		identity := MustGetIdentity(c)

		checker := rbac.NewPermissionChecker()
		hasPermissions, _ := checker.HasAllPermissions(c, identity, permissions...)

		if !hasPermissions {
			return c.Resp.WithStatus(403).Json(map[string]any{
				"error":   "Forbidden",
				"message": "missing all required permissions: " + permsStr,
			})
		}
		return c.Next()
	}
}

// forbiddenHandler returns a handler that always returns forbidden
func forbiddenHandler(message string) func(*request.Context) error {
	return func(c *request.Context) error {
		return c.Resp.WithStatus(403).Json(map[string]any{
			"error":   "Forbidden",
			"message": message,
		})
	}
}
