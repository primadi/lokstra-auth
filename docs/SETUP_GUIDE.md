# Platform Setup Guide - Step by Step

## 🎯 Overview

Panduan lengkap untuk setup platform **lokstra-auth** dari awal (fresh install).

---

## 📋 Prerequisites

### 1. Software Requirements

```bash
✅ PostgreSQL 14+
✅ Go 1.21+
✅ Git
✅ pgAdmin (optional, untuk GUI)
```

### 2. PostgreSQL Installation

**Windows:**
```powershell
# Download dari https://www.postgresql.org/download/windows/
# Atau via chocolatey:
choco install postgresql

# Verify installation
psql --version
```

**Linux:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**macOS:**
```bash
brew install postgresql@14
brew services start postgresql@14
```

---

## 🚀 Quick Start (Automated)

### **Option A: Fully Automated Setup** ✅ **RECOMMENDED**

Cukup 3 langkah:

```bash
# 1. Clone repository
git clone https://github.com/primadi/lokstra-auth
cd lokstra-auth/examples/01_deployment

# 2. Create database (manual - hanya sekali)
psql -U postgres -c "CREATE DATABASE lokstra_auth;"

# 3. Set environment variables
export SUPER_ADMIN_PASSWORD="YourSecurePassword123!"
export DB_DSN="postgres://postgres:postgres@localhost:5432/lokstra_auth?sslmode=disable"

# 4. Run application (migrations run automatically)
go run main.go
```

**Output yang diharapkan:**
```
╔════════════════════════════════════════════════════╗
║            MIGRATION STATUS                        ║
╚════════════════════════════════════════════════════╝

⏳ PENDING  000: bootstrap
⏳ PENDING  001: subject_rbac
⏳ PENDING  002: authz_policies
⏳ PENDING  006: tenant_ownership

Total: 4 migrations (0 applied, 4 pending)

╔════════════════════════════════════════════════════╗
║         DATABASE MIGRATION RUNNER                  ║
╚════════════════════════════════════════════════════╝

📋 Found 4 pending migration(s)

🔄 Running migration 000: bootstrap
   ✅ Completed in 124ms

🔄 Running migration 001: subject_rbac
   ✅ Completed in 89ms

🔄 Running migration 002: authz_policies
   ✅ Completed in 67ms

🔄 Running migration 006: tenant_ownership
   ✅ Completed in 156ms

╔════════════════════════════════════════════════════╗
║     ALL MIGRATIONS COMPLETED SUCCESSFULLY!         ║
╚════════════════════════════════════════════════════╝

╔════════════════════════════════════════════════════╗
║         CHECKING BOOTSTRAP STATUS                  ║
╚════════════════════════════════════════════════════╝

🚀 Database is empty - Starting bootstrap process...
  📋 Creating system tenant...
  📱 Creating admin console app...
  👤 Creating super admin user...
  🔑 Setting super admin password...
  🔗 Granting app access to super admin...
  👑 Creating super admin role...
  🎭 Assigning super admin role...
  ✨ Granting all permissions...
✅ Bootstrap completed successfully!

╔════════════════════════════════════════════════════╗
║           SUPER ADMIN CREDENTIALS                  ║
╠════════════════════════════════════════════════════╣
║ Tenant ID:  system                                 ║
║ App ID:     admin-console                          ║
║ Username:   admin                                  ║
║ Email:      admin@localhost                        ║
╠════════════════════════════════════════════════════╣
║ ⚠️  IMPORTANT: Change the super admin password!   ║
╚════════════════════════════════════════════════════╝

Server started on :8080
```

**✅ Done! Platform siap digunakan.**

---

## 📝 Detailed Setup (Manual Control)

### **Option B: Step-by-Step Manual Setup**

Jika ingin kontrol penuh atas setiap step:

#### **Step 1: Create Database**

**Via psql:**
```bash
psql -U postgres
```

```sql
-- Create database
CREATE DATABASE lokstra_auth;

-- Create user (optional, for production)
CREATE USER lokstra_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE lokstra_auth TO lokstra_user;

-- Verify
\l lokstra_auth
\q
```

**Via pgAdmin:**
1. Open pgAdmin
2. Connect to PostgreSQL server
3. Right-click "Databases" → "Create" → "Database"
4. Name: `lokstra_auth`
5. Owner: `postgres` (or custom user)
6. Click "Save"

#### **Step 2: Configure Application**

Create `.env` file:
```bash
cd examples/01_deployment
cat > .env << EOF
# Database Configuration
DB_DSN=postgres://postgres:postgres@localhost:5432/lokstra_auth?sslmode=disable

# Super Admin Password (for bootstrap)
SUPER_ADMIN_PASSWORD=Admin123!@#

# Server Configuration
SERVER_PORT=8080
EOF
```

Or set environment variables:
```bash
export DB_DSN="postgres://postgres:postgres@localhost:5432/lokstra_auth?sslmode=disable"
export SUPER_ADMIN_PASSWORD="Admin123!@#"
```

#### **Step 3: Configure config files**

Edit `config/db_main.yaml`:
```yaml
db_main:
  dsn: ${DB_DSN:postgres://postgres:postgres@localhost:5432/lokstra_auth?sslmode=disable}
  schema: lokstra_auth
```

Edit `config/api-server.yaml`:
```yaml
api-server:
  port: ${SERVER_PORT:8080}
  host: ${SERVER_HOST:0.0.0.0}
```

#### **Step 4: Run Migrations (Automated)**

Migrations run automatically when you start the app:

```bash
cd examples/01_deployment
go run main.go
```

The migration runner will:
1. ✅ Check which migrations are pending
2. ✅ Run migrations in order (000, 001, 002, 006)
3. ✅ Track applied migrations in `schema_migrations` table
4. ✅ Bootstrap system if database is empty
5. ✅ Start the server

#### **Step 5: Verify Setup**

**Check migrations:**
```sql
psql -U postgres -d lokstra_auth -c "
SELECT version, name, applied_at, duration_ms 
FROM schema_migrations 
ORDER BY version;
"
```

Expected output:
```
 version |      name           |     applied_at      | duration_ms 
---------+---------------------+---------------------+-------------
 000     | bootstrap           | 2025-11-27 10:30:15 |         124
 001     | subject_rbac        | 2025-11-27 10:30:16 |          89
 002     | authz_policies      | 2025-11-27 10:30:17 |          67
 006     | tenant_ownership    | 2025-11-27 10:30:18 |         156
```

**Check bootstrap tenant:**
```sql
psql -U postgres -d lokstra_auth -c "
SELECT id, name, owner_id, status 
FROM tenants;
"
```

Expected:
```
   id   |         name           | owner_id    | status 
--------+------------------------+-------------+--------
 system | System Administrator   | super-admin | active
```

**Check super admin:**
```sql
psql -U postgres -d lokstra_auth -c "
SELECT id, tenant_id, username, email, is_tenant_owner 
FROM users 
WHERE tenant_id = 'system';
"
```

Expected:
```
     id      | tenant_id | username |      email      | is_tenant_owner 
-------------+-----------+----------+-----------------+-----------------
 super-admin | system    | admin    | admin@localhost | true
```

---

## 🔑 First Login

### Test Super Admin Login

```http
POST http://localhost:8080/api/auth/cred/basic/authenticate
Content-Type: application/json

{
  "auth_context": {
    "tenant_id": "system",
    "app_id": "admin-console"
  },
  "username": "admin",
  "password": "Admin123!@#"
}
```

Expected response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "tenant_id": "system",
  "app_id": "admin-console",
  "user_id": "super-admin"
}
```

### Create First Tenant

```http
POST http://localhost:8080/api/auth/core/tenants
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "owner_id": "usr-alice",
  "db_dsn": "postgres://localhost/acme_db",
  "db_schema": "acme",
  "settings": {
    "max_users": 100,
    "max_apps": 10
  }
}
```

---

## 🔧 Troubleshooting

### Issue 1: Cannot connect to PostgreSQL

**Error:**
```
failed to acquire connection: connection refused
```

**Solution:**
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list | grep postgresql  # macOS
Get-Service postgresql*  # Windows PowerShell

# Start if not running
sudo systemctl start postgresql  # Linux
brew services start postgresql@14  # macOS
net start postgresql-x64-14  # Windows
```

### Issue 2: Database does not exist

**Error:**
```
database "lokstra_auth" does not exist
```

**Solution:**
```bash
psql -U postgres -c "CREATE DATABASE lokstra_auth;"
```

### Issue 3: Permission denied

**Error:**
```
permission denied for table tenants
```

**Solution:**
```sql
-- Grant permissions to user
GRANT ALL PRIVILEGES ON DATABASE lokstra_auth TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA lokstra_auth TO your_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA lokstra_auth TO your_user;
```

### Issue 4: Migration failed

**Error:**
```
migration 001 failed: syntax error at or near "..."
```

**Solution:**
```bash
# Check migration file
cat deployment/migrations/001_subject_rbac.sql

# Manually run problematic migration
psql -U postgres -d lokstra_auth -f deployment/migrations/001_subject_rbac.sql

# Check errors in detail
psql -U postgres -d lokstra_auth

# If needed, rollback and re-run
DELETE FROM schema_migrations WHERE version = '001';
```

### Issue 5: Bootstrap password not set

**Error:**
```
super admin password required: set SuperAdminPassword or SUPER_ADMIN_PASSWORD env var
```

**Solution:**
```bash
# Set environment variable
export SUPER_ADMIN_PASSWORD="YourSecurePassword123!"

# Or in PowerShell
$env:SUPER_ADMIN_PASSWORD = "YourSecurePassword123!"

# Or in .env file
echo "SUPER_ADMIN_PASSWORD=YourSecurePassword123!" >> .env
```

---

## 📊 Migration System Details

### How It Works

1. **Migration Files**
   - Located in `deployment/migrations/`
   - Named with version prefix: `000_bootstrap.sql`, `001_subject_rbac.sql`
   - Executed in alphabetical order

2. **Migration Tracking**
   - Table: `schema_migrations`
   - Records: version, name, applied_at, duration_ms
   - Prevents re-running applied migrations

3. **Automatic Execution**
   - Runs on application startup
   - Checks pending migrations
   - Executes in transaction (rollback on error)
   - Records execution time

### Migration Order

```
000_bootstrap.sql          → Core schema (tenants, apps, users)
001_subject_rbac.sql       → RBAC (roles, permissions, user_roles)
002_authz_policies.sql     → Policies (for policy-based authz)
006_tenant_ownership.sql   → Ownership (owner_id, is_tenant_owner)
```

### Checking Migration Status

```bash
# Via application logs (on startup)
go run main.go

# Via SQL query
psql -U postgres -d lokstra_auth -c "
SELECT 
  version,
  name,
  applied_at,
  duration_ms || 'ms' as duration
FROM schema_migrations
ORDER BY version;
"

# Check pending migrations
psql -U postgres -d lokstra_auth -c "
SELECT COUNT(*) as pending_count
FROM (
  SELECT '000' as version UNION
  SELECT '001' UNION
  SELECT '002' UNION
  SELECT '006'
) expected
LEFT JOIN schema_migrations applied ON expected.version = applied.version
WHERE applied.version IS NULL;
"
```

---

## 🎯 Production Deployment Checklist

- [ ] **Database**
  - [ ] PostgreSQL installed and running
  - [ ] Database `lokstra_auth` created
  - [ ] Dedicated user with proper permissions
  - [ ] SSL/TLS enabled for connections
  - [ ] Backup strategy in place

- [ ] **Configuration**
  - [ ] `DB_DSN` configured in secrets manager (not .env)
  - [ ] `SUPER_ADMIN_PASSWORD` set (strong password)
  - [ ] `SERVER_PORT` configured
  - [ ] Production config files ready

- [ ] **Migrations**
  - [ ] All migrations tested in staging
  - [ ] Migration order verified
  - [ ] Rollback plan documented

- [ ] **Bootstrap**
  - [ ] Super admin password changed after first login
  - [ ] Bootstrap tenant suspended after first real tenant
  - [ ] Initial tenant owner assigned

- [ ] **Security**
  - [ ] Environment variables secured
  - [ ] Database credentials rotated
  - [ ] SSL certificates configured
  - [ ] Firewall rules applied

- [ ] **Monitoring**
  - [ ] Application logs configured
  - [ ] Database logs enabled
  - [ ] Metrics collection setup
  - [ ] Alerts configured

---

## 📚 Related Documentation

- [Bootstrap Guide](../docs/BOOTSTRAP.md) - Super admin setup
- [Ownership Architecture](../docs/OWNERSHIP_AND_ROLES.md) - Owner vs Admin
- [Architecture Summary](../ARCHITECTURE_SUMMARY.md) - Complete overview
- [Migration Files](../deployment/migrations/) - SQL migration scripts

---

## 🆘 Support

Jika mengalami masalah:

1. Check logs: `examples/01_deployment/logs/`
2. Verify database: `psql -U postgres -d lokstra_auth`
3. Check migrations: `SELECT * FROM schema_migrations;`
4. Review config: `examples/01_deployment/config/`

---

## 🎉 Success!

Jika semua steps berhasil:

✅ Database created  
✅ Migrations applied automatically  
✅ Bootstrap tenant created  
✅ Super admin ready  
✅ Server running on :8080  

**You can now:**
- Login as platform admin
- Create tenants
- Assign tenant owners
- Manage multi-tenant system

🚀 **Platform is ready for production!**
