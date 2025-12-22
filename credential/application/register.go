package application

import (
	"github.com/primadi/lokstra-auth/token/jwt"
	"github.com/primadi/lokstra/lokstra_registry"
)

func init() {
	registerTokenManager()
}

func registerTokenManager() {
	// Register JWT token manager with default config
	jwtSecret := lokstra_registry.GetConfig("jwt.secret",
		"default-secret-change-in-production")
	jwtConfig := jwt.DefaultConfig(jwtSecret)
	jwtConfig.EnableRevocation = true

	tokenManager := jwt.NewManager(jwtConfig)
	lokstra_registry.RegisterService("token-manager", tokenManager)
}
