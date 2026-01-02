# Migration Runner - Microservices Deployment Guide

## Problem: Race Conditions & Slow Startup

Dalam microservices architecture, ada 2 masalah besar saat menjalankan migrations di startup:

### Problem 1: Race Condition (Multiple Instances)

```
┌──────────────┐
│ Instance 1   │──┐
├──────────────┤  │
│ Instance 2   │──┼──> Migration Runner ──> ❌ CONFLICT!
├──────────────┤  │     (Both try to run same migration)
│ Instance 3   │──┘
└──────────────┘
```

**Konsekuensi:**
- Duplicate migration execution
- Database errors (table already exists)
- Inconsistent migration state
- Service startup failures

### Problem 2: Slow Startup

```
Start → Check DB (2s) → Run Migrations (10s) → Ready
                                                 ↑
                                            Startup: 12s!
```

**Konsekuensi:**
- Slow scaling (container startup delay)
- Health check failures
- Kubernetes readiness probe timeout
- Poor user experience during deployment

## Solutions

### Solution 1: Advisory Lock (Recommended for Most Cases)

PostgreSQL advisory locks untuk memastikan hanya 1 instance yang run migrations.

**Cara Kerja:**
```
Instance 1: pg_try_advisory_lock(123456789) → ✅ true  → Run migrations
Instance 2: pg_try_advisory_lock(123456789) → ❌ false → Skip & continue
Instance 3: pg_try_advisory_lock(123456789) → ❌ false → Skip & continue
```

**Code:**

```go
// Use safe mode (with advisory lock)
err := deployment.RunMigrationsWithBootstrapSafe(
    "./deployment/migrations",
    deployment.BootstrapConfig{
        EnableAutoBootstrap: true,
    },
)
```

**Atau dengan explicit config:**

```go
runner := deployment.NewMigrationRunnerWithConfig(
    deployment.MigrationRunnerConfig{
        PoolName:       "db_main",
        MigrationsPath: "./deployment/migrations",
        SkipIfRunning:  true,  // Skip if another instance is running
        UseLock:        true,  // Use PostgreSQL advisory lock
    },
)

if err := runner.Initialize(); err != nil {
    log.Fatal(err)
}

if err := runner.RunPending(); err != nil {
    log.Fatal(err)
}
```

**Keuntungan:**
- ✅ Automatic coordination antar instances
- ✅ No external dependencies (menggunakan PostgreSQL built-in)
- ✅ Fast recovery (lock auto-released on disconnect)
- ✅ Simple implementation

**Kekurangan:**
- ⚠️ Semua instances tetap harus check database (ada overhead)
- ⚠️ Instance pertama delay startup saat run migrations

### Solution 2: Skip Migrations (Fastest Startup)

Jalankan migrations secara terpisah, tidak di startup aplikasi.

**Environment Variables:**

```bash
# Skip migrations completely
export SKIP_MIGRATIONS=true

# Skip bootstrap
export SKIP_BOOTSTRAP=true
```

**Code sudah support otomatis:**

```go
// main.go automatically checks environment variables
if !deployment.SkipMigrations() {
    // Run migrations
} else {
    log.Println("⏭️  SKIP_MIGRATIONS=true - Skipping database migrations")
}
```

**Deployment Flow:**

```bash
# Step 1: Run migrations once (separate job/pod)
docker run myapp migrate

# Step 2: Start multiple service instances (no migrations)
SKIP_MIGRATIONS=true docker run myapp serve
SKIP_MIGRATIONS=true docker run myapp serve
SKIP_MIGRATIONS=true docker run myapp serve
```

**Keuntungan:**
- ✅ VERY FAST startup (<1s)
- ✅ No race conditions
- ✅ Clear separation of concerns
- ✅ Better for Kubernetes (init containers)

**Kekurangan:**
- ⚠️ Requires separate migration job/script
- ⚠️ More complex deployment process

### Solution 3: External Migration Tool

Gunakan dedicated migration tool yang dijalankan manual atau via CI/CD.

**Tools:**
- `golang-migrate` - CLI migration tool
- `Flyway` - Database migration tool
- `Liquibase` - Database schema change management

**Example dengan golang-migrate:**

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path ./deployment/migrations \
        -database "postgres://user:pass@localhost:5432/lokstra_auth?sslmode=disable" \
        up
```

**Keuntungan:**
- ✅ Industry-standard tools
- ✅ Advanced features (rollback, version tracking)
- ✅ Independent dari aplikasi
- ✅ CI/CD integration

**Kekurangan:**
- ⚠️ External dependency
- ⚠️ Learning curve
- ⚠️ More complex deployment

## Deployment Strategies

### Strategy 1: Monolith (Single Instance)

**Use:** Regular migration runner

```go
// No special configuration needed
err := deployment.RunMigrationsWithBootstrap(
    "./deployment/migrations",
    deployment.BootstrapConfig{
        EnableAutoBootstrap: true,
    },
)
```

**Pros:**
- Simple
- No race conditions
- Straightforward

### Strategy 2: Microservices (Multiple Instances) - Advisory Lock

**Use:** Safe mode dengan advisory lock

```go
// Automatically handles multiple instances
err := deployment.RunMigrationsWithBootstrapSafe(
    "./deployment/migrations",
    deployment.BootstrapConfig{
        EnableAutoBootstrap: true,
    },
)
```

**Flow:**
```
Instance 1: Acquires lock → Runs migrations (10s) → Releases lock → Ready
Instance 2: Lock busy → Skips migrations (0s) → Ready (fast!)
Instance 3: Lock busy → Skips migrations (0s) → Ready (fast!)
```

**Pros:**
- Automatic coordination
- No external dependencies
- Simple implementation

**Cons:**
- All instances check database
- First instance delays startup

### Strategy 3: Kubernetes - Init Container

**Use:** Separate migration job, skip migrations in app

**kubernetes.yaml:**

```yaml
# Migration Job (runs once before deployment)
apiVersion: batch/v1
kind: Job
metadata:
  name: lokstra-auth-migration
spec:
  template:
    spec:
      containers:
      - name: migration
        image: lokstra-auth:latest
        command: ["./migrate"]  # Custom migration script
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
      restartPolicy: OnFailure

---
# Application Deployment (multiple replicas, no migrations)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-auth
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: lokstra-auth:latest
        env:
        - name: SKIP_MIGRATIONS
          value: "true"
        - name: SKIP_BOOTSTRAP
          value: "true"
        ports:
        - containerPort: 8080
```

**Pros:**
- ✅ Fast startup (<1s)
- ✅ No race conditions
- ✅ Kubernetes-native
- ✅ Clear separation

**Cons:**
- More complex deployment
- Requires init container/job

### Strategy 4: Kubernetes - Init Container (Inline)

**Use:** Init container dengan advisory lock

**kubernetes.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-auth
spec:
  replicas: 3
  template:
    spec:
      # Init container runs migrations (only first pod succeeds)
      initContainers:
      - name: migration
        image: lokstra-auth:latest
        command: ["./migrate-safe"]  # Uses advisory lock
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
      
      # Main container (all pods)
      containers:
      - name: app
        image: lokstra-auth:latest
        env:
        - name: SKIP_MIGRATIONS
          value: "true"
        ports:
        - containerPort: 8080
```

**Pros:**
- Automatic migration on deploy
- Advisory lock prevents conflicts
- Simple workflow

**Cons:**
- Init container overhead

### Strategy 5: CI/CD Pipeline

**Use:** Migrations run in CI/CD before deployment

**GitHub Actions Example:**

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Run Migrations
        run: |
          ./migrate.sh
        env:
          DB_DSN: ${{ secrets.DB_DSN }}
  
  deploy:
    needs: migrate  # Wait for migrations
    runs-on: ubuntu-latest
    steps:
      - name: Deploy Application
        run: |
          kubectl set image deployment/lokstra-auth \
            app=lokstra-auth:${{ github.sha }}
        env:
          SKIP_MIGRATIONS: "true"
```

**Pros:**
- ✅ Migrations before deployment
- ✅ Fast application startup
- ✅ Version control
- ✅ Rollback capability

**Cons:**
- CI/CD dependency
- More complex pipeline

## Environment Variables Reference

| Variable | Values | Default | Description |
|----------|--------|---------|-------------|
| `SKIP_MIGRATIONS` | `true`, `false`, `1`, `0`, `yes`, `no` | `false` | Skip database migrations completely |
| `SKIP_BOOTSTRAP` | `true`, `false`, `1`, `0`, `yes`, `no` | `false` | Skip bootstrap (super admin creation) |
| `SUPER_ADMIN_PASSWORD` | string | - | Password for bootstrap super admin |

## Code Examples

### Example 1: Development (Monolith)

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    deployment.SetupGlobalDB()
    
    // Simple migration runner
    err := deployment.RunMigrationsWithBootstrap(
        "./deployment/migrations",
        deployment.BootstrapConfig{
            EnableAutoBootstrap: true,
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    startServer()
}
```

### Example 2: Production (Microservices with Advisory Lock)

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    deployment.SetupGlobalDB()
    
    // Safe migration runner (advisory lock + skip if running)
    err := deployment.RunMigrationsWithBootstrapSafe(
        "./deployment/migrations",
        deployment.BootstrapConfig{
            EnableAutoBootstrap: true,
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    startServer()
}
```

### Example 3: Production (Skip Migrations)

```go
package main

import "github.com/primadi/lokstra-auth/deployment"

func main() {
    deployment.SetupGlobalDB()
    
    // Check environment variable
    if !deployment.SkipMigrations() {
        err := deployment.RunMigrationsWithBootstrapSafe(
            "./deployment/migrations",
            deployment.BootstrapConfig{
                EnableAutoBootstrap: !deployment.SkipBootstrap(),
            },
        )
        if err != nil {
            log.Fatal(err)
        }
    } else {
        log.Println("⏭️  Skipping migrations - database must be migrated separately")
    }
    
    // Start server (fast!)
    startServer()
}
```

### Example 4: Separate Migration Script

**migrate.go:**

```go
package main

import (
    "log"
    "github.com/primadi/lokstra-auth/deployment"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // Load configuration
    lokstra_registry.LoadConfigFromFolder("config")
    
    // Setup database
    deployment.SetupGlobalDB()
    
    // Run migrations only
    runner := deployment.NewGlobalDBMigrationRunnerSafe("./deployment/migrations")
    
    if err := runner.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    runner.PrintStatus()
    
    if err := runner.RunPending(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ Migrations completed successfully")
}
```

**Run:**

```bash
# Build migration binary
go build -o migrate migrate.go

# Run migrations
./migrate

# Then start application with SKIP_MIGRATIONS=true
SKIP_MIGRATIONS=true go run main.go
```

## Best Practices

### ✅ DO

1. **Use advisory lock in microservices**
   ```go
   RunMigrationsWithBootstrapSafe(...)
   ```

2. **Skip migrations for fast startup**
   ```bash
   SKIP_MIGRATIONS=true ./app
   ```

3. **Run migrations in init container (Kubernetes)**
   ```yaml
   initContainers:
   - name: migration
     command: ["./migrate"]
   ```

4. **Separate migration concerns**
   - Migration script: `./migrate`
   - Application: `./app serve`

5. **Test migrations before deployment**
   ```bash
   ./migrate --dry-run
   ```

### ❌ DON'T

1. **Don't run migrations without lock in microservices**
   ```go
   // ❌ WRONG - Race condition!
   RunMigrationsWithBootstrap(...) // in all instances
   ```

2. **Don't mix migration logic in application**
   ```go
   // ❌ WRONG - Tight coupling
   func StartServer() {
       runMigrations()  // Makes server startup slow
       serve()
   }
   ```

3. **Don't ignore migration failures**
   ```go
   // ❌ WRONG
   RunMigrations() // Ignore error
   StartServer()
   
   // ✅ CORRECT
   if err := RunMigrations(); err != nil {
       log.Fatal(err)
   }
   ```

## Performance Comparison

| Strategy | First Instance | Other Instances | Total Startup |
|----------|----------------|-----------------|---------------|
| **No Lock (Race)** | 12s | 12s (ERROR) | ❌ Fails |
| **Advisory Lock** | 12s | 0.5s | ⚠️ Mixed |
| **Skip Migrations** | 0.5s | 0.5s | ✅ Fast |
| **Init Container** | 12s (init) + 0.5s (app) | 0.5s | ✅ Good |

## Troubleshooting

### Problem: Migrations timeout in Kubernetes

**Symptom:**
```
Init container migration exceeded timeout (60s)
```

**Solution:**
Increase init container timeout:
```yaml
initContainers:
- name: migration
  command: ["./migrate"]
  # Increase timeout
  livenessProbe:
    initialDelaySeconds: 120
```

### Problem: Advisory lock not released

**Symptom:**
```
All instances skip migrations - lock stuck
```

**Solution:**
Locks are auto-released on connection close. Manual release:
```sql
-- Check active locks
SELECT * FROM pg_locks WHERE locktype = 'advisory';

-- Force release (if needed)
SELECT pg_advisory_unlock_all();
```

### Problem: Slow startup even with SKIP_MIGRATIONS

**Symptom:**
```
Startup still slow (5s) even with SKIP_MIGRATIONS=true
```

**Solution:**
Database connection pooling might be slow. Check:
1. Database network latency
2. Connection pool size
3. Health check queries

## Summary

| Deployment Type | Recommended Strategy | Startup Time |
|-----------------|---------------------|--------------|
| **Development** | Regular migrations | ~12s |
| **Monolith** | Regular migrations | ~12s |
| **Microservices (Simple)** | Advisory lock | First: 12s, Others: <1s |
| **Microservices (Production)** | Skip migrations + separate job | <1s all |
| **Kubernetes** | Init container | <1s app |
| **CI/CD** | Pipeline migration | <1s app |

**Recommendation:**
- **Development**: Use regular migrations
- **Production**: Use `SKIP_MIGRATIONS=true` + separate migration job
- **Quick Fix**: Use `RunMigrationsWithBootstrapSafe()` for advisory lock
