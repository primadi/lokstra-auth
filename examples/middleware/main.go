package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	lokstraauth "github.com/primadi/lokstra-auth"
	"github.com/primadi/lokstra-auth/01_credential/basic"
	"github.com/primadi/lokstra-auth/02_token/jwt"
	"github.com/primadi/lokstra-auth/03_subject/simple"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
	"github.com/primadi/lokstra-auth/middleware"
	"github.com/primadi/lokstra/core/request"
)

func main() {
	fmt.Println("=== Lokstra Auth - Middleware Example ===")
	fmt.Println()

	// ========== Setup Auth Runtime ==========
	userProvider := basic.NewInMemoryUserProvider()

	// Add test users
	adminHash, _ := basic.HashPassword("admin123")
	userProvider.AddUser(&basic.User{
		ID:           "admin-001",
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: adminHash,
	})

	devHash, _ := basic.HashPassword("dev123")
	userProvider.AddUser(&basic.User{
		ID:           "dev-001",
		Username:     "developer",
		Email:        "dev@example.com",
		PasswordHash: devHash,
	})

	// Build Auth runtime
	auth := lokstraauth.NewBuilder().
		WithAuthenticator("basic", basic.NewAuthenticator(
			userProvider,
			basic.NewValidator(basic.DefaultValidatorConfig()),
		)).
		WithTokenManager(jwt.NewManager(
			jwt.DefaultConfig("my-secret-key-change-in-production"),
		)).
		WithSubjectResolver(simple.NewResolver()).
		WithIdentityContextBuilder(simple.NewContextBuilder(
			simple.NewStaticRoleProvider(map[string][]string{
				"admin-001": {"admin", "developer"},
				"dev-001":   {"developer"},
			}),
			simple.NewStaticPermissionProvider(map[string][]string{
				"admin-001": {"read:users", "write:users", "delete:users"},
				"dev-001":   {"read:code", "write:code"},
			}),
			simple.NewStaticGroupProvider(map[string][]string{}),
			simple.NewStaticProfileProvider(map[string]map[string]any{}),
		)).
		WithAuthorizer(rbac.NewEvaluator(map[string][]string{
			"admin":     {"*"}, // Admin has all permissions
			"developer": {"read:code", "write:code", "read:users"},
		})).
		EnableRefreshToken().
		Build()

	// ========== Setup Lokstra Application ==========
	r := lokstra.NewRouter("auth-router")

	// ========== Routes: Public (No Authentication) ==========

	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type userInfo struct {
		ID    string   `json:"id"`
		Email string   `json:"email"`
		Roles []string `json:"roles"`
	}

	type loginResponse struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		User         userInfo `json:"user"`
	}

	// Login endpoint
	r.POST("/login", func(c *request.Context, r *loginRequest) error {
		// Authenticate
		loginResp, err := auth.Login(c, &lokstraauth.LoginRequest{
			Credentials: &basic.BasicCredentials{
				Username: r.Username,
				Password: r.Password,
			},
		})
		if err != nil {
			return c.Api.Unauthorized(err.Error())
		}

		return c.Api.Ok(&loginResponse{
			AccessToken:  loginResp.AccessToken.Value,
			RefreshToken: loginResp.RefreshToken.Value,
			User: userInfo{
				ID:    loginResp.Identity.Subject.ID,
				Email: loginResp.Identity.Subject.Principal,
				Roles: loginResp.Identity.Roles,
			},
		})
	})

	// ========== Setup Middlewares ==========
	authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		Auth: auth,
	})

	// ========== Routes: Protected (Requires Authentication) ==========

	// Profile endpoint - requires authentication only
	r.GET("/profile",
		func(c *request.Context) error {
			identity, _ := middleware.GetIdentity(c)

			return c.Api.Ok(map[string]any{
				"user_id":     identity.Subject.ID,
				"email":       identity.Subject.Principal,
				"roles":       identity.Roles,
				"permissions": identity.Permissions,
				"groups":      identity.Groups,
			})
		},
		authMiddleware.Handler(),
	)

	// ========== Routes: Role-Based ==========

	// Admin dashboard - requires 'admin' role
	r.GET("/admin/dashboard",
		func(c *request.Context) error {
			identity, _ := middleware.GetIdentity(c)

			return c.Api.Ok(map[string]any{
				"message": "Welcome to admin dashboard",
				"user":    identity.Subject.Principal,
			})
		},
		authMiddleware.Handler(),
		middleware.RequireRole(auth, "admin"),
	)

	// ========== Routes: Permission-Based ==========

	// Users list - requires 'read:users' permission
	r.GET("/api/users",
		func(c *request.Context) error {
			return c.Api.Ok(map[string]any{
				"users": []map[string]any{
					{"id": "admin-001", "username": "admin", "role": "admin"},
					{"id": "dev-001", "username": "developer", "role": "developer"},
				},
			})
		},
		authMiddleware.Handler(),
		middleware.RequirePermission(auth, "read:users"),
	)

	// Delete user - requires 'delete:users' permission
	r.DELETE("/api/users/:id",
		func(c *request.Context) error {
			userID := c.Req.PathParam("id", "")

			return c.Api.Ok(map[string]any{
				"message": "User deleted",
				"user_id": userID,
			})
		},
		authMiddleware.Handler(),
		middleware.RequirePermission(auth, "delete:users"),
	)

	// ========== Routes: Multiple Permissions ==========

	// Code repository endpoint - requires ANY: read:code OR write:code
	anyPermMiddleware := middleware.NewAnyPermissionMiddleware(
		auth,
		[]string{"read:code", "write:code"},
	)

	r.GET("/api/code",
		func(c *request.Context) error {
			return c.Resp.Json(map[string]any{
				"repository": "lokstra-auth",
				"files":      []string{"main.go", "auth.go", "middleware.go"},
			})
		},
		authMiddleware.Handler(),
		anyPermMiddleware.Handler(),
	)

	// ========== Routes: Multiple Roles ==========

	// Special endpoint - requires ANY: admin OR team-lead role
	anyRoleMiddleware := middleware.NewAnyRoleMiddleware(
		auth,
		[]string{"admin", "team-lead"},
	)

	r.POST("/api/deploy",
		func(c *request.Context) error {
			identity, _ := middleware.GetIdentity(c)

			return c.Resp.Json(map[string]any{
				"message":    "Deployment initiated",
				"started_by": identity.Subject.Principal,
			})
		},
		authMiddleware.Handler(),
		anyRoleMiddleware.Handler(),
	)

	// ========== Routes: Optional Authentication ==========

	// Public content with optional auth
	optionalAuthMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
		Auth:     auth,
		Optional: true, // Don't fail if no token provided
	})

	r.GET("/public/content",
		func(c *request.Context) error {
			identity, authenticated := middleware.GetIdentity(c)

			response := map[string]any{
				"content": "This is public content",
			}

			if authenticated {
				response["user"] = identity.Subject.Principal
				response["message"] = "You are logged in"
			} else {
				response["message"] = "You are viewing as guest"
			}

			return c.Api.Ok(response)
		},
		optionalAuthMiddleware.Handler(),
	)

	// ========== Start Server ==========
	fmt.Println("✓ Server starting on http://localhost:3000")
	fmt.Println()
	fmt.Println("Test endpoints:")
	fmt.Println("1. POST /login - Login with username/password")
	fmt.Println("   {\"username\": \"admin\", \"password\": \"admin123\"}")
	fmt.Println()
	fmt.Println("2. GET /profile - View your profile (requires auth)")
	fmt.Println("   Header: Authorization: Bearer <token>")
	fmt.Println()
	fmt.Println("3. GET /admin/dashboard - Admin only")
	fmt.Println("4. GET /api/users - Requires read:users permission")
	fmt.Println("5. DELETE /api/users/:id - Requires delete:users permission")
	fmt.Println("6. GET /api/code - Requires read:code OR write:code")
	fmt.Println("7. GET /public/content - Public content (optional auth)")
	fmt.Println()

	app := lokstra.NewApp("app1", ":3000", r)

	app.PrintStartInfo()
	if err := app.Run(30 * time.Second); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
