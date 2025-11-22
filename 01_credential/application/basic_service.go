package application

import (
	"fmt"

	"github.com/primadi/lokstra-auth/01_credential/domain"
	"github.com/primadi/lokstra-auth/01_credential/domain/basic"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/hasher"
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/repository"
	"github.com/primadi/lokstra-auth/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// BasicAuthService handles basic authentication (username/password) via HTTP
// @RouterService name="basic-auth-service", prefix="/api/cred/basic", middlewares=["recovery", "request-logger"]
type BasicAuthService struct {
	// @Inject "basic-authenticator"
	Authenticator *service.Cached[domain.Authenticator]
	// @Inject "user-provider"
	UserProvider *service.Cached[repository.UserProvider]
	// @Inject "credential-validator"
	Validator *service.Cached[repository.CredentialValidator]
	// @Inject "credential-config-resolver"
	ConfigResolver *service.Cached[*ConfigResolver]
}

// Login authenticates user with username/password
// @Route "POST /login"
func (s *BasicAuthService) Login(ctx *request.Context, req *basic.LoginRequest) (*basic.LoginResponse, error) {
	// Check if basic auth is enabled
	if !s.ConfigResolver.MustGet().IsBasicEnabled(ctx, req.TenantID, req.AppID) {
		return &basic.LoginResponse{
			Success: false,
			Error:   "basic authentication is not enabled for this application",
		}, nil
	}

	// Build auth context
	authCtx := &domain.AuthContext{
		TenantID:  req.TenantID,
		AppID:     req.AppID,
		BranchID:  req.BranchID,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	// Build credentials
	creds := &basic.Credentials{
		Username: req.Username,
		Password: req.Password,
	}

	// Authenticate
	result, err := s.Authenticator.MustGet().Authenticate(ctx, authCtx, creds)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	if !result.Success {
		return &basic.LoginResponse{
			Success: false,
			Error:   result.Error.Error(),
		}, nil
	}

	// TODO: Generate token using 02_token layer
	// For now, return success with user info
	user, err := s.UserProvider.MustGet().GetUserByID(ctx, result.TenantID, result.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &basic.LoginResponse{
		Success:     true,
		AccessToken: "TODO_GENERATE_TOKEN", // Will be handled by 02_token layer
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		User:        basic.ToUserInfo(user),
	}, nil
}

// Register creates a new user account
// @Route "POST /register"
func (s *BasicAuthService) Register(ctx *request.Context, req *basic.RegisterRequest) (*basic.RegisterResponse, error) {
	// Check if basic auth is enabled
	if !s.ConfigResolver.MustGet().IsBasicEnabled(ctx, req.TenantID, "") {
		return &basic.RegisterResponse{
			Success: false,
			Error:   "basic authentication is not enabled for this tenant",
		}, nil
	}

	// Get effective config for validation rules
	config := s.ConfigResolver.MustGet().GetBasicConfig(ctx, req.TenantID, "")

	// Validate username length
	if len(req.Username) < config.MinUsernameLength {
		return &basic.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("username must be at least %d characters", config.MinUsernameLength),
		}, nil
	}
	if len(req.Username) > config.MaxUsernameLength {
		return &basic.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("username must not exceed %d characters", config.MaxUsernameLength),
		}, nil
	}

	// Validate password length
	if len(req.Password) < config.MinPasswordLength {
		return &basic.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("password must be at least %d characters", config.MinPasswordLength),
		}, nil
	}

	// Validate username
	if err := s.Validator.MustGet().ValidateUsername(req.Username); err != nil {
		return &basic.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid username: %s", err.Error()),
		}, nil
	}

	// Validate password complexity
	if err := s.Validator.MustGet().ValidatePassword(req.Password); err != nil {
		return &basic.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid password: %s", err.Error()),
		}, nil
	}

	// Hash password
	passwordHash, err := hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &basic.User{
		ID:           utils.GenerateID("usr"),
		TenantID:     req.TenantID,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		Disabled:     false,
		Metadata:     req.Metadata,
	}

	if err := s.UserProvider.MustGet().CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &basic.RegisterResponse{
		Success: true,
		User:    basic.ToUserInfo(user),
	}, nil
}

// ChangePassword changes user password
// @Route "POST /change-password"
func (s *BasicAuthService) ChangePassword(ctx *request.Context, req *basic.ChangePasswordRequest) error {
	// Get user
	user, err := s.UserProvider.MustGet().GetUserByID(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if !hasher.VerifyPassword(user.PasswordHash, req.OldPassword) {
		return fmt.Errorf("invalid old password")
	}

	// Validate new password
	if err := s.Validator.MustGet().ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}

	// Hash new password
	newHash, err := hasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = newHash
	if err := s.UserProvider.MustGet().UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
