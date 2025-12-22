# 🚀 Lokstra Auth - Quick Start Guide

Get up and running in 5 minutes!

## Prerequisites

- **PostgreSQL 14+** ([Download](https://www.postgresql.org/download/windows/))
- **Go 1.21+** ([Download](https://go.dev/dl/))

## Option 1: Automated Setup (Recommended)

### Windows

```powershell
cd examples/01_deployment
.\setup.ps1
```

The script will:
1. ✅ Check prerequisites (PostgreSQL, Go)
2. ✅ Create database `lokstra_auth`
3. ✅ Generate configuration files
4. ✅ Set environment variables
5. ✅ Run migrations automatically
6. ✅ Bootstrap super admin
7. ✅ Save credentials to `CREDENTIALS.txt`

### After Setup

Your super admin credentials:
- **Tenant ID**: `system`
- **App ID**: `admin-console`
- **Username**: `admin`
- **Email**: `admin@localhost`
- **Password**: (what you entered)

Check `CREDENTIALS.txt` for full details.

## Option 2: Manual Setup

### Step 1: Create Database

```sql
-- Using pgAdmin or psql
CREATE DATABASE lokstra_auth;
```

### Step 2: Configure

Create `config/db_main.yaml`:

```yaml
db_main:
  dsn: postgres://postgres:yourpassword@localhost:5432/lokstra_auth?sslmode=disable
  schema: lokstra_auth
```

Create `config/api-server.yaml`:

```yaml
api-server:
  port: 8080
  host: 0.0.0.0
  
api-auth-prefix: /api/auth
```

### Step 3: Set Environment Variables

```powershell
# Windows PowerShell
$env:SUPER_ADMIN_PASSWORD="YourStrongPassword123!"
```

```bash
# Linux/Mac
export SUPER_ADMIN_PASSWORD="YourStrongPassword123!"
```

### Step 4: Run

```bash
cd examples/01_deployment
go run main.go
```

You'll see:
```
✅ Running migrations...
✅ Migration db_schema.sql applied
✅ Migration 000_bootstrap.sql applied
✅ Migration 001_subject_rbac.sql applied
✅ Migration 002_authz_policies.sql applied
✅ Migration 006_tenant_ownership.sql applied

🚀 Bootstrap: Creating super admin...
✅ Bootstrap: Super admin created
   Tenant: system
   App: admin-console
   User: admin
   Email: admin@localhost

🌐 API Server starting on :8080
```

## First Steps

### 1. Test Super Admin Login

```bash
curl -X POST http://localhost:8080/api/auth/cred/basic/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "auth_context": {
      "tenant_id": "system",
      "app_id": "admin-console"
    },
    "username": "admin",
    "password": "YourStrongPassword123!"
  }'
```

Response:
```json
{
  "credential_id": "...",
  "user_id": "...",
  "tenant_id": "system",
  "verified": true
}
```

### 2. Generate Access Token

```bash
curl -X POST http://localhost:8080/api/auth/token/jwt/generate \
  -H "Content-Type: application/json" \
  -d '{
    "credential_id": "<from_previous_response>",
    "app_id": "admin-console",
    "duration": 3600
  }'
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 3. Create Your First Tenant

```bash
curl -X POST http://localhost:8080/api/auth/core/tenants \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acme-corp",
    "name": "Acme Corporation",
    "domain": "acme.com",
    "db_dsn": "postgres://localhost:5432/acme_db",
    "db_schema": "public",
    "status": "active"
  }'
```

### 4. Create Tenant Owner

```bash
curl -X POST http://localhost:8080/api/auth/core/tenants/acme-corp/users \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "john-doe",
    "email": "john@acme.com",
    "full_name": "John Doe",
    "username": "john",
    "is_tenant_owner": true,
    "status": "active"
  }'
```

## Architecture Overview

```
┌─────────────────────────────────────────────┐
│         Platform Level                      │
│  ┌─────────────────────────────────┐        │
│  │  System Tenant (Bootstrap)       │        │
│  │  - Super Admin                   │        │
│  │  - Platform Management           │        │
│  └─────────────────────────────────┘        │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│         Tenant Level                        │
│  ┌─────────────────────────────────┐        │
│  │  Tenant: acme-corp               │        │
│  │  - Tenant Owner (1)              │ ◄───── Billing, Ownership
│  │  - Tenant Admin (N)              │ ◄───── Manage Users/Apps
│  │  - Regular Users (N)             │ ◄───── Access Resources
│  └─────────────────────────────────┘        │
└─────────────────────────────────────────────┘
```

## Database Strategy

- **Global DB** (`lokstra_auth`): ALL authentication data
  - Tenants
  - Users (all tenants)
  - Roles, Permissions
  - Policies
  - Credentials

- **Tenant DB** (`tenant.db_dsn`): Business data only
  - Documents
  - Orders
  - Products
  - Custom tables

## Key Features

### ✅ Multi-Tenancy
- Complete tenant isolation
- Shared authentication
- Separate business databases

### ✅ Ownership Management
- 1 owner per tenant
- Transferable ownership
- Audit trail

### ✅ Bootstrap System
- Auto-creates super admin
- Self-disables after first tenant
- Secure by default

### ✅ Credential Providers
- Basic (username/password)
- API Key
- OAuth2
- Passkey (WebAuthn)
- Passwordless (magic link)

### ✅ Authorization
- RBAC (Role-Based Access Control)
- ABAC (Attribute-Based Access Control)
- ACL (Access Control Lists)
- Policy-Based

## Security Best Practices

1. **Change Default Password**
   ```bash
   # After first login, change super admin password
   curl -X PUT http://localhost:8080/api/auth/core/users/admin/password \
     -H "Authorization: Bearer <token>" \
     -d '{"new_password": "NewSecurePassword"}'
   ```

2. **Disable Bootstrap Tenant**
   - Automatically disabled after creating first real tenant
   - Or manually: `UPDATE tenants SET status='inactive' WHERE id='system'`

3. **Use Strong Passwords**
   - Minimum 10 characters
   - Mix of uppercase, lowercase, numbers, symbols

4. **Enable SSL/TLS**
   - Use `sslmode=require` in PostgreSQL DSN
   - Configure HTTPS for API server

5. **Rotate Credentials**
   - API keys: Regular rotation
   - Passwords: Enforce expiry
   - Tokens: Short-lived (1 hour)

## Troubleshooting

### Database Connection Failed

```
Error: failed to connect to database
```

**Solution:**
- Check PostgreSQL is running
- Verify `db_main.dsn` in config
- Test connection: `psql -U postgres -h localhost -d lokstra_auth`

### Migrations Failed

```
Error: migration db_schema.sql failed
```

**Solution:**
- Check migration logs
- Verify database permissions
- Reset migrations: `DELETE FROM schema_migrations;`
- Re-run application

### Bootstrap Failed

```
Error: failed to create super admin
```

**Solution:**
- Check `SUPER_ADMIN_PASSWORD` is set
- Verify migrations completed
- Check logs for detailed error

### Login Failed

```
Error: invalid credentials
```

**Solution:**
- Verify `tenant_id`, `app_id`, `username`, `password`
- Check user status is `active`
- Verify bootstrap completed successfully

## Next Steps

1. **Read Documentation**
   - [Setup Guide](docs/SETUP_GUIDE.md) - Detailed setup
   - [Bootstrap Guide](docs/BOOTSTRAP.md) - Bootstrap options
   - [Ownership Guide](docs/OWNERSHIP_AND_ROLES.md) - Ownership architecture
   - [Architecture Summary](ARCHITECTURE_SUMMARY.md) - System overview

2. **Explore Examples**
   - [HTTP Tests](examples/01_deployment/http-tests/bootstrap/) - API examples
   - [Complete Example](examples/complete/) - Full integration

3. **Integrate Your App**
   - Configure credential providers
   - Set up authorization policies
   - Implement middleware
   - Add custom business logic

## Support

- **Documentation**: See `/docs` folder
- **Examples**: See `/examples` folder
- **Issues**: Check migration logs and troubleshooting guide

---

**Ready to build multi-tenant applications! 🎉**
