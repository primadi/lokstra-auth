package application

import (
	"github.com/primadi/lokstra-auth/01_credential/infrastructure/repository"
	"github.com/primadi/lokstra/lokstra_registry"
)

func Register() {
	// Register credential-config-resolver
	// This will be used by BasicAuthService and APIKeyAuthService
	lokstra_registry.RegisterService("credential-config-resolver", &ConfigResolver{})
	repository.Register()
}
