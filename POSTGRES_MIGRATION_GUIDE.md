# Migration Guide: In-Memory to PostgreSQL Stores

This guide helps you migrate from in-memory stores to PostgreSQL stores in your lokstra-auth application.

## Why Migrate to PostgreSQL?

| Requirement | In-Memory | PostgreSQL | Recommendation |
|-------------|-----------|------------|----------------|
| Development/Testing | ✅ Perfect | ⚠️ Overkill | Use In-Memory |
| Production | ❌ Not suitable | ✅ Recommended | Use PostgreSQL |
| Data Persistence | ❌ Lost on restart | ✅ Persistent | Use PostgreSQL |
| Multi-Instance | ❌ No sharing | ✅ Shared DB | Use PostgreSQL |
| Large Datasets | ❌ Memory limits | ✅ Scalable | Use PostgreSQL |
| Backup/Recovery | ❌ Not possible | ✅ Standard tools | Use PostgreSQL |

## Prerequisites

1. PostgreSQL server (version 12+)
2. Database created for lokstra-auth
3. Database credentials
4. `github.com/lib/pq` driver installed

```bash
go get github.com/lib/pq
```

## Step-by-Step Migration

### Step 1: Prepare PostgreSQL Database

**Option A: Local PostgreSQL**
```bash
# Create database
createdb lokstra_auth

# Verify connection
psql -d lokstra_auth -c "SELECT version();"
```

**Option B: Docker PostgreSQL**
```bash
docker run --name lokstra-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=lokstra_auth \
  -p 5432:5432 \
  -d postgres:15
```

**Option C: Cloud PostgreSQL**
Use your cloud provider's PostgreSQL service (AWS RDS, Azure Database, GCP Cloud SQL, etc.)

### Step 2: Initialize Database Schema

**Using Go migration tool:**

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra-auth/infrastructure/repository"
)

func main() {
    cfg := repository.DBConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "your_password",
        DBName:   "lokstra_auth",
        SSLMode:  "disable",
    }
    
    db, err := repository.NewPostgresConnection(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    if err := repository.MigrateDatabase(context.Background(), db); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Database initialized successfully!")
}
```

**Or using psql directly:**

```bash
psql -d lokstra_auth -f core/infrastructure/repository/db_schema.sql
```

### Step 3: Update Your Application Code

#### Before (In-Memory Stores)

```go
package main

import (
    "github.com/primadi/lokstra-auth/infrastructure/repository"
)

func main() {
    // Create in-memory stores
    tenantStore := repository.NewInMemoryTenantStore()
    appStore := repository.NewInMemoryAppStore()
    branchStore := repository.NewInMemoryBranchStore()
    userStore := repository.NewInMemoryUserStore()
    userAppStore := repository.NewInMemoryUserAppStore()
    appKeyStore := repository.NewInMemoryAppKeyStore()
    
    // Use stores in services...
    tenantService := application.NewTenantService(tenantStore)
    // ... etc
}
```

#### After (PostgreSQL Stores)

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/primadi/lokstra-auth/infrastructure/repository"
)

func main() {
    // Configure database connection
    cfg := repository.DBConfig{
        Host:     os.Getenv("DB_HOST"),
        Port:     getEnvInt("DB_PORT", 5432),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
        SSLMode:  os.Getenv("DB_SSL_MODE"),
    }
    
    // Connect to database
    db, err := repository.NewPostgresConnection(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Run migrations
    ctx := context.Background()
    if err := repository.MigrateDatabase(ctx, db); err != nil {
        log.Fatal(err)
    }
    
    // Create PostgreSQL stores
    factory := repository.NewPostgresStoreFactory(db)
    stores := factory.CreateAllStores()
    
    // Use stores in services (same interface!)
    tenantService := application.NewTenantService(stores.Tenant)
    appService := application.NewAppService(stores.App, stores.Tenant)
    userService := application.NewUserService(stores.User, stores.Tenant)
    // ... etc
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

### Step 4: Environment Configuration

Create a `.env` file or set environment variables:

```bash
# .env file
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=lokstra_auth
DB_SSL_MODE=disable  # use 'require' in production
```

**Load environment variables** (using a package like `godotenv`):

```go
import "github.com/joho/godotenv"

func init() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }
}
```

### Step 5: Migrate Existing Data (if any)

If you have data in in-memory stores that needs to be migrated:

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra-auth/infrastructure/repository"
)

func migrateData() {
    // Old in-memory stores
    oldTenantStore := repository.NewInMemoryTenantStore()
    oldAppStore := repository.NewInMemoryAppStore()
    // ... load data into old stores ...
    
    // New PostgreSQL stores
    db, _ := repository.NewPostgresConnection(cfg)
    factory := repository.NewPostgresStoreFactory(db)
    newStores := factory.CreateAllStores()
    
    ctx := context.Background()
    
    // Migrate tenants
    tenants, _ := oldTenantStore.List(ctx)
    for _, tenant := range tenants {
        if err := newStores.Tenant.Create(ctx, tenant); err != nil {
            log.Printf("Failed to migrate tenant %s: %v", tenant.ID, err)
        }
    }
    
    // Migrate apps
    // ... similar for other entities ...
}
```

### Step 6: Update Tests

If you have tests using in-memory stores, you can:

**Option A: Keep using in-memory for unit tests**
```go
func TestTenantService(t *testing.T) {
    // Use in-memory for fast unit tests
    store := repository.NewInMemoryTenantStore()
    service := application.NewTenantService(store)
    
    // ... test service ...
}
```

**Option B: Use test database**
```go
func TestTenantServiceWithDB(t *testing.T) {
    // Use test database for integration tests
    cfg := repository.DBConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "postgres",
        DBName:   "lokstra_auth_test",
        SSLMode:  "disable",
    }
    
    db, err := repository.NewPostgresConnection(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    
    // Clean database before test
    cleanDatabase(t, db)
    
    factory := repository.NewPostgresStoreFactory(db)
    service := application.NewTenantService(factory.CreateTenantStore())
    
    // ... test service ...
}

func cleanDatabase(t *testing.T, db *sql.DB) {
    _, err := db.Exec(`
        TRUNCATE user_apps, app_keys, branches, users, apps, tenants CASCADE
    `)
    if err != nil {
        t.Fatal(err)
    }
}
```

### Step 7: Deployment

#### Development Deployment

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=dev_password
export DB_NAME=lokstra_auth
export DB_SSL_MODE=disable

# Run application
go run main.go
```

#### Production Deployment

**Using systemd service:**

```ini
# /etc/systemd/system/lokstra-auth.service
[Unit]
Description=Lokstra Auth Service
After=network.target postgresql.service

[Service]
Type=simple
User=lokstra
WorkingDirectory=/opt/lokstra-auth
Environment="DB_HOST=db.example.com"
Environment="DB_PORT=5432"
Environment="DB_USER=lokstra_user"
Environment="DB_PASSWORD=secure_password"
Environment="DB_NAME=lokstra_auth"
Environment="DB_SSL_MODE=require"
ExecStart=/opt/lokstra-auth/app
Restart=always

[Install]
WantedBy=multi-user.target
```

**Using Docker Compose:**

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=lokstra
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=lokstra_auth
      - DB_SSL_MODE=disable
    depends_on:
      - postgres
    ports:
      - "8080:8080"

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_USER=lokstra
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=lokstra_auth
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

## Rollback Plan

If you need to rollback to in-memory stores:

1. **Keep both implementations in code** (use feature flag):

```go
var usePostgres = os.Getenv("USE_POSTGRES") == "true"

func createStores() any {
    if usePostgres {
        db, _ := repository.NewPostgresConnection(cfg)
        factory := repository.NewPostgresStoreFactory(db)
        return factory.CreateAllStores()
    } else {
        return &InMemoryStores{
            Tenant:  repository.NewInMemoryTenantStore(),
            App:     repository.NewInMemoryAppStore(),
            User:    repository.NewInMemoryUserStore(),
            // ... etc
        }
    }
}
```

2. **Switch by changing environment variable**:
```bash
export USE_POSTGRES=false
```

## Performance Tuning

### Connection Pool

```go
db, _ := repository.NewPostgresConnection(cfg)

// Tune for your workload
db.SetMaxOpenConns(50)      // Max concurrent connections
db.SetMaxIdleConns(10)      // Idle connections to keep
db.SetConnMaxLifetime(time.Hour) // Max connection lifetime
```

### Indexes

Add indexes for your specific query patterns:

```sql
-- If you frequently query by tenant status
CREATE INDEX idx_tenants_status ON tenants(status);

-- If you frequently join users and apps
CREATE INDEX idx_user_apps_tenant_user ON user_apps(tenant_id, user_id);
```

### Monitoring

```sql
-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Troubleshooting

### Connection Issues

```
Failed to connect: pq: connection refused
```
**Solution**: Check PostgreSQL is running and firewall allows connections

### Authentication Issues

```
Failed to connect: pq: password authentication failed
```
**Solution**: Verify username and password in configuration

### Schema Issues

```
Failed to create tenant: pq: relation "tenants" does not exist
```
**Solution**: Run migrations to create schema

### Performance Issues

```
Queries are slow
```
**Solution**: 
1. Check and add indexes
2. Use `EXPLAIN ANALYZE` to identify bottlenecks
3. Tune connection pool settings
4. Monitor database resources

## Best Practices

1. **Always use environment variables** for credentials
2. **Enable SSL/TLS** in production (`sslmode=require`)
3. **Set up regular backups** using `pg_dump` or cloud backup tools
4. **Monitor database performance** using pg_stat_statements
5. **Use connection pooling** appropriately
6. **Test migrations** in staging before production
7. **Keep database and application versions in sync**
8. **Use prepared statements** (automatic with database/sql)
9. **Implement health checks** to verify database connectivity
10. **Log slow queries** for optimization

## Migration Checklist

- [ ] PostgreSQL server installed and running
- [ ] Database created
- [ ] Schema initialized (migrations run)
- [ ] Application code updated to use PostgreSQL stores
- [ ] Environment variables configured
- [ ] Tests updated/verified
- [ ] Connection pooling tuned
- [ ] SSL/TLS enabled (production)
- [ ] Backup strategy implemented
- [ ] Monitoring configured
- [ ] Rollback plan documented
- [ ] Team trained on new setup

## Support

For issues during migration:
- Review `POSTGRES_README.md` for detailed documentation
- Check example in `examples/postgres_store_example/`
- Compare implementations in `store_inmemory.go` vs `store_postgres.go`
- Test with the example application first

---

**Remember**: The store interfaces are identical, so your service layer code doesn't need to change!
