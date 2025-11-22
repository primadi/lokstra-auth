package application

import "github.com/primadi/lokstra-auth/00_core/infrastructure/repository"

// Register triggers package initialization for code generation
func Register() {
	// This function is called during initialization
	// to ensure the package is loaded for annotation processing
	repository.Register()
}
