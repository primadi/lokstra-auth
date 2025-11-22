package repository

import (
	"github.com/primadi/lokstra/lokstra_registry"
)

// Register registers all credential repository stores as services
func Register() {
	// Register credential stores for dependency injection
	lokstra_registry.RegisterService("credential-user-store", NewInMemoryUserStore())
	lokstra_registry.RegisterService("user-provider", NewInMemoryUserStore()) // Alias for user provider
	lokstra_registry.RegisterService("apikey-store", NewInMemoryAPIKeyStore())
	lokstra_registry.RegisterService("credential-validator", NewSimpleValidator())
	lokstra_registry.RegisterService("simple-validator", NewSimpleValidator()) // Alias
}
