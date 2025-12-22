# Migration Strategy - Global DB vs Tenant DB

## Overview

Migration Runner di lokstra-auth mendukung 2 tipe database migrations:

1. **Global DB Migrations** - Authentication & tenant management schema
2. **Tenant DB Migrations** - Business data schema (per tenant)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GLOBAL DATABASE                          │
│                   (lokstra_auth schema)                     │
├─────────────────────────────────────────────────────────────┤
│  Schema: lokstra_auth                                       │
│  Pool Name: "db_main"                                     │
│                                                             │
│  Tables:                                                    │
│  - tenants              (ALL tenants)                       │
│  - users                (ALL users, all tenants)            │
│  - apps                 (ALL applications)                  │
│  - roles                (ALL roles)                         │
│  - permissions          (ALL permissions)                   │
│  - user_roles           (ALL role assignments)              │
│  - user_permissions     (ALL permission grants)             │
│  - credentials          (ALL login credentials)             │
│  - policies             (ALL authorization policies)        │
│  - schema_migrations    (Global migration tracking)         │
│                                                             │
│  Purpose: Single source of truth for authentication         │
│  Migrations: deployment/migrations/*.sql                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ References tenant.db_dsn
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              TENANT DATABASE (per tenant)                   │
│                (Business data schema)                       │
├─────────────────────────────────────────────────────────────┤
│  Schema: tenant.db_schema (e.g., "public", "acme")          │
│  Pool Name: "tenant-db-{tenant_id}"                         │
│  DSN: tenant.db_dsn (from tenants table)                    │
│                                                             │
│  Tables: (Business-specific, varies by application)         │
│  - documents                                                │
│  - orders                                                   │
│  - products                                                 │
│  - invoices                                                 │
│  - custom_tables                                            │
│  - schema_migrations    (Tenant migration tracking)         │
│                                                             │
│  Purpose: Isolated business data per tenant                 │
│  Migrations: app-specific migration files                   │
└─────────────────────────────────────────────────────────────┘
```

## Migration Tracking

Migration tracking adalah **per database** (bukan per schema):

### Global DB Migration Tracking

```sql
-- In db_main database
CREATE TABLE schema_migrations (
    version VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
    duration_ms INTEGER,
    checksum VARCHAR(64)
);
```

**Tracks:**
- `000_bootstrap.sql` - Super admin setup
- `001_subject_rbac.sql` - Roles & permissions
- `002_authz_policies.sql` - Authorization policies
- `006_tenant_ownership.sql` - Ownership constraints
- `db_schema.sql` - Core tables

### Tenant DB Migration Tracking

```sql
-- In each tenant-specific database
CREATE TABLE {tenant_schema}.schema_migrations (
    version VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
    duration_ms INTEGER,
    checksum VARCHAR(64)
);
```

**Tracks:**
- Business-specific migrations
- Application schema changes
- Custom table migrations

**Important:** Each tenant database has its own `schema_migrations` table, tracking only migrations applied to that specific tenant's database.

## Usage

### 1. Global DB Migrations (Authentication)

Run migrations for the global authentication database:

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    // Run global DB migrations (authentication schema)
    err := deployment.RunMigrationsWithBootstrap(
        "../../../deployment/migrations",
        deployment.BootstrapConfig{
            EnableAutoBootstrap: true,
        },
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

**What it does:**
1. Creates `schema_migrations` table in db_main
2. Scans `deployment/migrations/*.sql` files
3. Runs pending migrations (000, 001, 002, 006, db_schema)
4. Records applied migrations
5. Runs bootstrap if database is empty

**Pool:** Uses `"db_main"` pool from lokstra_registry

### 2. Tenant DB Migrations (Business Data)

Run migrations for a specific tenant's database:

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    // First, register tenant database pool
    // (This would typically be done when tenant is created)
    
    // Run tenant-specific migrations
    runner := deployment.NewTenantDBMigrationRunner(
        "acme-corp",                    // tenant ID
        "./migrations/tenant-schema",   // path to tenant migrations
    )
    
    if err := runner.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    runner.PrintStatus()
    
    if err := runner.RunPending(); err != nil {
        log.Fatal(err)
    }
}
```

**What it does:**
1. Creates `schema_migrations` table in tenant's database
2. Scans tenant migration files
3. Runs pending migrations
4. Records applied migrations

**Pool:** Uses `"tenant-db-{tenant_id}"` pool from lokstra_registry

### 3. Custom Database Migrations

Run migrations for any database pool:

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    // Run migrations for custom pool
    runner := deployment.NewMigrationRunner(
        "my-custom-pool",       // any registered pool name
        "./migrations/custom",  // migration files
    )
    
    if err := runner.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    if err := runner.RunPending(); err != nil {
        log.Fatal(err)
    }
}
```

## Migration Isolation

### Database Level

Each database has its own `schema_migrations` table:

```
PostgreSQL Server
├── lokstra_auth (db_main)
│   └── schema_migrations (tracks global auth migrations)
│
├── acme_corp_db (tenant-db-acme-corp)
│   └── schema_migrations (tracks acme corp migrations)
│
└── widgets_inc_db (tenant-db-widgets-inc)
    └── schema_migrations (tracks widgets inc migrations)
```

**Why separate?**
- Each database can be on different PostgreSQL servers
- Each tenant can have different schema versions
- Independent migration rollback per tenant
- No cross-contamination of migration states

### Schema Level (Within Database)

Within a database, you can have multiple schemas:

```
lokstra_auth database:
├── lokstra_auth schema (authentication tables)
│   ├── tenants
│   ├── users
│   └── schema_migrations (global tracking)
│
└── public schema (other application data)
    └── (other tables)
```

**Note:** `schema_migrations` table is created in the **default schema** of the database pool.

## Best Practices

### ✅ DO

1. **Run global migrations first**
   ```go
   // Before any tenant operations
   RunMigrationsWithBootstrap("./deployment/migrations", cfg)
   ```

2. **Run tenant migrations on tenant creation**
   ```go
   // When creating new tenant
   func CreateTenant(tenant *Tenant) error {
       // 1. Create tenant in global DB
       // 2. Register tenant DB pool
       // 3. Run tenant migrations
       runner := NewTenantDBMigrationRunner(tenant.ID, "./tenant-migrations")
       return runner.RunPending()
   }
   ```

3. **Use separate migration directories**
   ```
   deployment/
   ├── migrations/           ← Global auth migrations
   │   ├── 000_bootstrap.sql
   │   ├── 001_subject_rbac.sql
   │   └── db_schema.sql
   │
   └── tenant-migrations/    ← Tenant business migrations
       ├── 001_create_documents.sql
       ├── 002_create_orders.sql
       └── 003_create_products.sql
   ```

4. **Version migrations independently**
   - Global: `000`, `001`, `002`, etc.
   - Tenant: `001`, `002`, `003`, etc. (can reuse numbers)

### ❌ DON'T

1. **Don't mix auth and business migrations**
   ```go
   // ❌ WRONG - Auth tables in tenant DB
   CREATE TABLE tenants (...);  // Should be in db_main!
   
   // ✅ CORRECT - Business tables in tenant DB
   CREATE TABLE documents (...);
   ```

2. **Don't share migration tracking**
   ```go
   // ❌ WRONG - Trying to use global schema_migrations for tenant
   // Each database needs its own tracking!
   ```

3. **Don't run tenant migrations before global**
   ```go
   // ❌ WRONG order
   RunTenantMigrations()
   RunGlobalMigrations()  // Too late!
   
   // ✅ CORRECT order
   RunGlobalMigrations()
   RunTenantMigrations()
   ```

## Migration Files

### Global DB Migration Example

File: `deployment/migrations/000_bootstrap.sql`

```sql
-- This runs in db_main (lokstra_auth schema)
-- Creates super admin for initial access

INSERT INTO tenants (id, name, status, metadata)
VALUES (
    'system',
    'System Tenant',
    'active',
    '{"is_bootstrap": true}'::jsonb
);

INSERT INTO users (tenant_id, id, username, email, status)
VALUES (
    'system',
    'admin',
    'admin',
    'admin@localhost',
    'active'
);

-- ... more global auth setup
```

### Tenant DB Migration Example

File: `tenant-migrations/001_create_documents.sql`

```sql
-- This runs in tenant-specific database
-- Creates business tables for document management

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    owner_id VARCHAR(50) NOT NULL,  -- references user in db_main
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_owner ON documents(owner_id);

-- ... more tenant-specific tables
```

## API Reference

### NewMigrationRunner

```go
func NewMigrationRunner(poolName string, migrationsPath string) *MigrationRunner
```

**Parameters:**
- `poolName`: Registered pool name in lokstra_registry
- `migrationsPath`: Directory containing .sql migration files

**Returns:** MigrationRunner instance

### NewGlobalDBMigrationRunner

```go
func NewGlobalDBMigrationRunner(migrationsPath string) *MigrationRunner
```

**Convenience function** for global authentication database.

**Parameters:**
- `migrationsPath`: Directory containing global migration files

**Returns:** MigrationRunner configured for "db_main" pool

### NewTenantDBMigrationRunner

```go
func NewTenantDBMigrationRunner(tenantID string, migrationsPath string) *MigrationRunner
```

**Convenience function** for tenant-specific database.

**Parameters:**
- `tenantID`: Tenant identifier
- `migrationsPath`: Directory containing tenant migration files

**Returns:** MigrationRunner configured for "tenant-db-{tenantID}" pool

**Note:** Tenant database pool must be registered first!

### RunMigrationsWithBootstrap

```go
func RunMigrationsWithBootstrap(migrationsPath string, bootstrapCfg BootstrapConfig) error
```

**High-level function** that runs global migrations + bootstrap.

**Parameters:**
- `migrationsPath`: Global migration directory
- `bootstrapCfg`: Bootstrap configuration

**Returns:** Error if migrations or bootstrap fail

## Summary

| Aspect | Global DB | Tenant DB |
|--------|-----------|-----------|
| **Database** | Single (lokstra_auth) | Multiple (one per tenant) |
| **Schema** | lokstra_auth | tenant.db_schema |
| **Pool Name** | `"db_main"` | `"tenant-db-{tenant_id}"` |
| **Migration Tracking** | schema_migrations in db_main | schema_migrations in each tenant db |
| **Migration Files** | deployment/migrations/*.sql | tenant-migrations/*.sql |
| **Purpose** | Authentication, tenants, users | Business data, documents, orders |
| **When to Run** | Application startup | Tenant creation |
| **Function** | `NewGlobalDBMigrationRunner()` | `NewTenantDBMigrationRunner()` |

**Answer:** Migration Runner berlaku **per DATABASE** (bukan per schema). Setiap database punya `schema_migrations` table sendiri untuk tracking migrations yang sudah diapply.
