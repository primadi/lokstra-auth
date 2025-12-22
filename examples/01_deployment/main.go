package main

import (

	// Import all application packages to trigger registration

	authmiddleware "github.com/primadi/lokstra-auth/middleware"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
	// lokstra.SetLogLevel(lokstra.LogLevelDebug)

	// Bootstrap MUST be called first to generate registration code
	if err := lokstra_init.BootstrapAndRun(
		lokstra_init.WithAnnotations(true,
			"../../core",
			"../../credential",
			"../../token",
			"../../rbac",
			"../../authz",
			"../../infrastructure",
			"../../middleware",
			"services",
		),
		// lokstra_init.WithPgSyncMap(true, "db_main"),
		// lokstra_init.WithDbPoolAutoSync(true),
		lokstra_init.WithDbMigrations(true, "migrations"),
		lokstra_init.WithServerInitFunc(func() error {
			logger.LogInfo("╔═══════════════════════════════════════════════╗")
			logger.LogInfo("║      LOKSTRA AUTH - Authentication System     ║")
			logger.LogInfo("║   Multi-Tenant Multi-Database Architecture    ║")
			logger.LogInfo("╚═══════════════════════════════════════════════╝")

			recovery.Register()
			request_logger.Register()
			authmiddleware.Register()

			return nil
		}),
	); err != nil {
		logger.LogPanic("Failed to start Lokstra Auth service: %v", err)
	}
}
