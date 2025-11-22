package main

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra"

	// Import all application packages to trigger registration
	coreapp "github.com/primadi/lokstra-auth/00_core/application"

	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
	// Bootstrap MUST be called first to generate registration code
	lokstra.Bootstrap("../../../00_core/application", "../../../01_credential/application")

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║      LOKSTRA AUTH - Authentication System     ║")
	fmt.Println("║   Multi-Tenant Multi-Database Architecture    ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println("")

	lokstra_registry.LoadConfigFromFolder("config")

	// Trigger package loading for @RouterService annotations
	coreapp.Register()
	// credapp.Register()

	// Register middlewares
	recovery.Register()
	request_logger.Register()

	// run server manager
	if err := lokstra_registry.InitAndRunServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server stopped")
}
