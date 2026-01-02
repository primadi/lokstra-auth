# Migration Management - SysAdmin Guide

## Problem

Jika semua service instances menggunakan `skip-migrations: true` untuk fast startup, bagaimana cara update database schema?

## Solutions

### Solution 1: CLI Migration Tool (Recommended)

**Command-line tool** untuk menjalankan migrations secara manual.

#### Build Migration Binary

```bash
# Build migration tool
go build -o migrate ./cmd/migrate

# Or build for production
CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate
```

#### Usage

**1. Check Migration Status**

```bash
./migrate --status
```

Output:
```
╔═══════════════════════════════════════════════╗
║      LOKSTRA AUTH - Migration Manager         ║
║         Database Schema Updater               ║
╚═══════════════════════════════════════════════╝

╔════════════════════════════════════════════════════╗
║   MIGRATION STATUS - db_main                     ║
╚════════════════════════════════════════════════════╝

✅ APPLIED  000: bootstrap (at 2025-11-27 10:00:00, took 1.2s)
✅ APPLIED  001: subject_rbac (at 2025-11-27 10:00:01, took 850ms)
⏳ PENDING  007: add_new_feature

Total: 3 migrations (2 applied, 1 pending)
```

**2. Run Pending Migrations**

```bash
./migrate
```

Output:
```
╔════════════════════════════════════════════════════╗
║   MIGRATION STATUS - db_main                     ║
╚════════════════════════════════════════════════════╝

⏳ PENDING  007: add_new_feature

⚠️  About to run pending migrations. Continue? [y/N]: y

🔒 Acquired migration lock
🔄 Running migration 007: add_new_feature
   ✅ Completed in 450ms

╔════════════════════════════════════════════════════╗
║     ALL MIGRATIONS COMPLETED SUCCESSFULLY!         ║
╚════════════════════════════════════════════════════╝

✅ Migration completed successfully!
```

**3. Custom Migration Path**

```bash
./migrate --path ./custom/migrations
```

**4. Disable Advisory Lock** (single instance only)

```bash
./migrate --safe=false
```

#### Docker/Kubernetes Usage

**Docker:**

```bash
# Run migration in separate container
docker run --rm \
  -v $(pwd)/config:/app/config \
  lokstra-auth:latest \
  ./migrate
```

**Kubernetes Job:**

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: lokstra-auth-migration
  annotations:
    # Run before deployment
    "helm.sh/hook": pre-upgrade
    "helm.sh/hook-weight": "-5"
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: migrate
        image: lokstra-auth:latest
        command: ["./migrate"]
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: lokstra-auth-config
```

---

### Solution 2: HTTP API Endpoints (For SysAdmin UI)

**REST API** untuk manage migrations via web interface.

#### Setup

Add migration API handler in your main.go:

```go
package main

import (
    "github.com/primadi/lokstra-auth/deployment"
    "github.com/primadi/lokstra/core/app"
)

func main() {
    // ... existing setup ...
    
    // Create migration API handler
    migrationAPI := deployment.NewMigrationAPIHandler("./deployment/migrations")
    
    // Register admin routes (with strong authentication!)
    adminAPI := app.Group("/admin")
    adminAPI.Use(RequireSuperAdminMiddleware()) // ⚠️ IMPORTANT!
    
    adminAPI.GET("/migrations/status", migrationAPI.GetStatus)
    adminAPI.POST("/migrations/run", migrationAPI.RunMigrations)
    adminAPI.GET("/migrations/history", migrationAPI.GetMigrationHistory)
    
    // ... start server ...
}
```

#### API Endpoints

**1. GET /admin/migrations/status** - Check migration status

Request:
```bash
curl -X GET http://localhost:8080/admin/migrations/status \
  -H "Authorization: Bearer <super_admin_token>"
```

Response:
```json
{
  "all_migrations": [
    {
      "version": "000",
      "name": "bootstrap",
      "status": "completed",
      "applied_at": "2025-11-27T10:00:00Z",
      "duration": 1200000000
    },
    {
      "version": "007",
      "name": "add_new_feature",
      "status": "pending"
    }
  ],
  "pending_migrations": [...],
  "applied_migrations": [...],
  "total": 3,
  "pending_count": 1,
  "applied_count": 2
}
```

**2. POST /admin/migrations/run** - Execute pending migrations

Request:
```bash
curl -X POST http://localhost:8080/admin/migrations/run \
  -H "Authorization: Bearer <super_admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"confirm": true}'
```

Response (Success):
```json
{
  "status": "success",
  "message": "migrations completed successfully",
  "applied_count": 1,
  "applied_migrations": [
    {
      "version": "007",
      "name": "add_new_feature",
      "status": "completed",
      "duration": 450000000
    }
  ]
}
```

Response (No Pending):
```json
{
  "status": "up_to_date",
  "message": "no pending migrations"
}
```

Response (Error):
```json
{
  "error": "migration failed",
  "details": "table already exists"
}
```

**3. GET /admin/migrations/history** - View migration history

Request:
```bash
curl -X GET http://localhost:8080/admin/migrations/history \
  -H "Authorization: Bearer <super_admin_token>"
```

Response:
```json
{
  "total": 2,
  "migrations": [
    {
      "version": "000",
      "name": "bootstrap",
      "status": "completed",
      "applied_at": "2025-11-27T10:00:00Z",
      "duration": 1200000000
    },
    {
      "version": "001",
      "name": "subject_rbac",
      "status": "completed",
      "applied_at": "2025-11-27T10:00:01Z",
      "duration": 850000000
    }
  ]
}
```

#### Security Considerations

⚠️ **CRITICAL:** Migration endpoints MUST be protected with strong authentication!

```go
func RequireSuperAdminMiddleware() middleware.Middleware {
    return func(c *request.Context) error {
        // Verify token
        token := c.R.Header.Get("Authorization")
        if token == "" {
            return c.Resp.WithStatus(401).Json(map[string]string{
                "error": "unauthorized",
            })
        }
        
        // Verify super admin role
        identity := GetIdentityFromToken(token)
        if !identity.IsSuperAdmin() {
            return c.Resp.WithStatus(403).Json(map[string]string{
                "error": "forbidden - super admin only",
            })
        }
        
        return c.Next()
    }
}
```

**Additional Security:**

1. **Rate Limiting** - Prevent abuse
2. **IP Whitelist** - Only allow from internal network
3. **Audit Logging** - Log all migration attempts
4. **Confirmation Required** - Must send `{"confirm": true}`

---

### Solution 3: Kubernetes CronJob (Automatic)

**Scheduled migration** check (auto-update schema).

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: lokstra-auth-migration-check
spec:
  # Run every day at 2 AM
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: migrate
            image: lokstra-auth:latest
            command: ["./migrate"]
            env:
            - name: DB_DSN
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: dsn
```

**Pros:**
- ✅ Automatic migration
- ✅ No manual intervention

**Cons:**
- ⚠️ Less control
- ⚠️ May run at unexpected times

---

### Solution 4: CI/CD Pipeline (Recommended for Production)

**Deploy migrations before application** in CI/CD pipeline.

**GitHub Actions Example:**

```yaml
name: Deploy Production

on:
  push:
    branches: [main]

jobs:
  migrate:
    name: Database Migration
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Build Migration Tool
        run: go build -o migrate ./cmd/migrate
      
      - name: Run Migrations
        run: ./migrate --status && ./migrate
        env:
          DB_DSN: ${{ secrets.PRODUCTION_DB_DSN }}
      
      - name: Verify Migration Success
        run: ./migrate --status
  
  deploy:
    name: Deploy Application
    needs: migrate  # Wait for migrations to complete
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/lokstra-auth \
            app=lokstra-auth:${{ github.sha }}
        env:
          SKIP_MIGRATIONS: "true"  # App instances skip migrations
```

**Pros:**
- ✅ Migrations before deployment
- ✅ Rollback capability
- ✅ Version controlled
- ✅ Audit trail

---

## Deployment Workflows

### Workflow 1: Manual CLI (Simple)

```bash
# Step 1: Sysadmin runs migration
./migrate

# Step 2: Deploy application (all instances skip migrations)
kubectl apply -f deployment.yaml  # SKIP_MIGRATIONS=true
```

### Workflow 2: API-Based (UI)

```
1. Sysadmin logs into Admin Panel
2. Navigate to "Database Migrations"
3. Click "Check Status"
4. Click "Run Migrations" (if pending)
5. Deploy application
```

### Workflow 3: Automated (CI/CD)

```
1. Push code to main branch
2. CI/CD runs migrations automatically
3. CI/CD deploys application (skip migrations)
4. Application instances start fast
```

### Workflow 4: Kubernetes (Pre-Deploy)

```
1. helm upgrade lokstra-auth ./chart
2. Helm pre-upgrade hook runs migration Job
3. Migration completes successfully
4. Helm deploys application pods (skip migrations)
5. Pods start fast (<1s)
```

---

## Best Practices

### ✅ DO

1. **Use CLI tool for production**
   ```bash
   ./migrate
   ```

2. **Run migrations BEFORE deployment**
   ```
   migrate → test → deploy
   ```

3. **Use advisory lock** (default in CLI)
   - Prevents concurrent execution

4. **Backup database before migration**
   ```bash
   pg_dump lokstra_auth > backup.sql
   ./migrate
   ```

5. **Test migrations in staging first**

6. **Use CI/CD for automation**

7. **Monitor migration execution**

### ❌ DON'T

1. **Don't run migrations in application startup** (in production)
   ```go
   // ❌ WRONG in production
   RunMigrations() // Slow startup
   ```

2. **Don't expose migration API publicly**
   ```go
   // ❌ WRONG
   publicAPI.POST("/migrations/run") // Security risk!
   ```

3. **Don't skip confirmation**
   ```go
   // ❌ WRONG
   RunMigrations() // No confirmation
   ```

4. **Don't ignore migration failures**

5. **Don't run migrations manually in production** (without testing)

---

## Configuration

### Application (Skip Migrations)

```yaml
# config/deployment.yaml
skip-migrations: true
skip-bootstrap: true
```

### Migration Tool (config/db_main.yaml)

```yaml
# Config used by ./migrate command
db_main:
  dsn: postgres://user:pass@localhost:5432/lokstra_auth?sslmode=disable
  schema: lokstra_auth
```

---

## Monitoring & Alerts

### Prometheus Metrics

```go
// Add to migration_runner.go
migrationDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "lokstra_migration_duration_seconds",
        Help: "Duration of migrations",
    },
    []string{"version", "status"},
)

migrationCount := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "lokstra_migration_total",
        Help: "Total number of migrations executed",
    },
    []string{"status"},
)
```

### Slack Notifications

```bash
# In CI/CD pipeline
./migrate && \
  curl -X POST https://hooks.slack.com/... \
    -d '{"text": "✅ Database migration completed successfully"}'
```

---

## Troubleshooting

### Problem: Migration stuck (advisory lock)

**Solution:**
```sql
-- Check active locks
SELECT * FROM pg_locks WHERE locktype = 'advisory';

-- Force release
SELECT pg_advisory_unlock_all();
```

### Problem: Migration failed halfway

**Solution:**
```bash
# Check status
./migrate --status

# Fix the failed migration SQL
vim deployment/migrations/007_failed.sql

# Re-run
./migrate
```

### Problem: Need to rollback migration

**Solution:**
Currently not supported. Manual rollback:
```sql
-- Revert changes manually
DROP TABLE new_table;

-- Remove from tracking
DELETE FROM schema_migrations WHERE version = '007';
```

---

## Summary

| Method | Use Case | Startup Time | Complexity |
|--------|----------|--------------|------------|
| **CLI Tool** | Manual/Scheduled | <1s app | Low |
| **HTTP API** | Admin UI | <1s app | Medium |
| **CronJob** | Automatic | <1s app | Medium |
| **CI/CD** | Production | <1s app | High |

**Recommendation:**
- **Development**: Run in app startup
- **Staging**: CLI tool or CI/CD
- **Production**: CI/CD pipeline + CLI tool for emergencies
