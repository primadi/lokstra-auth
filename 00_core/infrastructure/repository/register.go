package repository

import (
	"github.com/primadi/lokstra-auth/00_core/infrastructure/repository/postgres"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/services/dbpool_manager"
)

// Register registers all repository stores as services
func Register() {

	// Register DBPool Manager
	dbPoolM := dbpool_manager.NewPgxPoolManager()
	lokstra_registry.RegisterService("dbpool-manager", dbPoolM)

	// Get DSN and schema for global-db
	dsn := lokstra_registry.GetConfig("global-db.dsn", "")
	if dsn == "" {
		panic("global-db.dsn configuration is required")
	}
	schema := lokstra_registry.GetConfig("global-db.schema", "public")

	// Set DSN and Schema for global-db
	dbPoolM.SetNamedDsn("global-db", dsn, schema)

	// Get the global-db pool
	globalDb, err := dbPoolM.GetNamedPool("global-db")
	if err != nil {
		panic(err)
	}

	// Register the global-db pool as a service
	lokstra_registry.RegisterService("global-db", globalDb)

	// Register each repository store as a service with the global-db pool
	lokstra_registry.RegisterService("tenant-store", postgres.NewTenantStore(globalDb))
	lokstra_registry.RegisterService("app-store", postgres.NewAppStore(globalDb))
	lokstra_registry.RegisterService("branch-store", postgres.NewBranchStore(globalDb))
	lokstra_registry.RegisterService("user-store", postgres.NewUserStore(globalDb))
	lokstra_registry.RegisterService("app-key-store", postgres.NewAppKeyStore(globalDb))
	lokstra_registry.RegisterService("user-app-store", postgres.NewUserAppStore(globalDb))
}
