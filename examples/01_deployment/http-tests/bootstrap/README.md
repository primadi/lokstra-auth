# Bootstrap Testing - Initial System Setup

## 📋 Overview

Test file ini mendemonstrasikan **complete workflow** dari sistem kosong sampai siap production:

1. ✅ Login sebagai Super Admin (bootstrap)
2. ✅ Create tenant pertama
3. ✅ Create app untuk tenant
4. ✅ Create admin user untuk tenant
5. ✅ Setup roles dan permissions
6. ✅ Disable bootstrap tenant
7. ✅ Change super admin password

## 🚀 Prerequisites

### 1. Database Setup

```bash
# Create database
psql -U postgres -c "CREATE DATABASE lokstra_auth;"

# Run migrations
psql -U postgres -d lokstra_auth -f deployment/migrations/db_schema.sql
psql -U postgres -d lokstra_auth -f deployment/migrations/001_subject_rbac.sql
psql -U postgres -d lokstra_auth -f deployment/migrations/002_authz_policies.sql
```

### 2. Set Super Admin Password

```bash
# Option A: Environment variable
export SUPER_ADMIN_PASSWORD="Admin123!@#"

# Option B: .env file
echo "SUPER_ADMIN_PASSWORD=Admin123!@#" > .env
```

### 3. Run Application

```bash
cd examples/01_deployment
go run main.go
```

**Expected Output:**
```
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

╔════════════════════════════════════════════════════════╗
║           SUPER ADMIN CREDENTIALS                      ║
╠════════════════════════════════════════════════════════╣
║ Tenant ID:  system                                     ║
║ App ID:     admin-console                              ║
║ Username:   admin                                      ║
║ Email:      admin@localhost                            ║
╚════════════════════════════════════════════════════════╝
```

## 📝 Test Files

### 00-initial-setup.http

**Complete bootstrap workflow** dengan steps:

#### Phase 1: Super Admin Login
- Login dengan bootstrap credentials
- Verify token validity
- Check super admin identity

#### Phase 2: Create First Tenant
- Create tenant "acme-corp"
- Get tenant details
- List all tenants

#### Phase 3: Setup App
- Create "web-portal" app
- Configure auth methods
- Verify app creation

#### Phase 4: Create Tenant Admin
- Create user "alice"
- Grant app access
- Verify user details

#### Phase 5: Setup Authorization
- Create "Administrator" role
- Create permissions (tenant, user, app management)
- Assign permissions to role
- Assign role to alice

#### Phase 6: Verify Tenant Admin
- Login as alice
- Check permissions
- Test creating new user

#### Phase 7: Cleanup
- Disable bootstrap tenant
- Verify suspension
- Change super admin password

## 🔑 Default Credentials

### Bootstrap Super Admin
```
Tenant ID:  system
App ID:     admin-console
Username:   admin
Email:      admin@localhost
Password:   (from SUPER_ADMIN_PASSWORD env var)
```

### Test Tenant Admin (Created in tests)
```
Tenant ID:  acme-corp
App ID:     web-portal
Username:   alice
Email:      alice@acme.com
Password:   AliceSecure123!@#
```

## ⚡ Quick Start

### VS Code REST Client

1. Install extension: [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

2. Open `00-initial-setup.http`

3. Update variables (if needed):
   ```http
   @bootstrapPassword = Admin123!@#  # Match SUPER_ADMIN_PASSWORD
   @newTenantId = acme-corp
   @newAppId = web-portal
   ```

4. Run requests **in order** (top to bottom)

### Using cURL

```bash
# 1. Login Super Admin
curl -X POST http://localhost:9090/api/auth/cred/basic/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "auth_context": {
      "tenant_id": "system",
      "app_id": "admin-console"
    },
    "username": "admin",
    "password": "Admin123!@#"
  }'

# Save access_token from response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 2. Create Tenant
curl -X POST http://localhost:9090/api/auth/core/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acme-corp",
    "name": "Acme Corporation",
    "settings": {
      "max_users": 100,
      "max_apps": 10
    }
  }'

# 3. Create App
curl -X POST http://localhost:9090/api/auth/core/tenants/acme-corp/apps \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "web-portal",
    "name": "Web Portal",
    "type": "web"
  }'

# ... continue with remaining steps
```

## 🎯 Testing Scenarios

### Scenario 1: First Deployment

**Goal:** Setup production system from scratch

**Steps:**
1. ✅ Run all requests in `00-initial-setup.http`
2. ✅ Verify bootstrap tenant created
3. ✅ Create first real tenant
4. ✅ Setup admin user with full permissions
5. ✅ Disable bootstrap tenant
6. ✅ Change super admin password

**Success Criteria:**
- ✅ Super admin can login
- ✅ Tenant admin can login
- ✅ Tenant admin has correct permissions
- ✅ Bootstrap tenant is suspended
- ✅ Super admin password changed

### Scenario 2: Multi-Tenant Setup

**Goal:** Create multiple tenants

**Steps:**
1. ✅ Login as super admin
2. ✅ Create tenant A (acme-corp)
3. ✅ Create tenant B (globex-inc)
4. ✅ Create tenant C (initech-llc)
5. ✅ Setup admin for each tenant
6. ✅ Verify isolation between tenants

**Success Criteria:**
- ✅ Each tenant has separate users
- ✅ Each tenant has separate apps
- ✅ No cross-tenant access
- ✅ Each tenant admin can only manage own tenant

### Scenario 3: Permission Testing

**Goal:** Verify RBAC works correctly

**Steps:**
1. ✅ Create roles: Admin, Editor, Viewer
2. ✅ Create permissions: read, write, delete
3. ✅ Assign permissions to roles
4. ✅ Create users with different roles
5. ✅ Test access control

**Success Criteria:**
- ✅ Admin has full access
- ✅ Editor can write but not delete
- ✅ Viewer can only read
- ✅ Permission inheritance from roles works

### Scenario 4: Bootstrap Recovery

**Goal:** Test bootstrap disable/re-enable

**Steps:**
1. ✅ Disable bootstrap tenant
2. ✅ Verify login fails
3. ✅ Re-enable via database
4. ✅ Verify login works again

**Database Command:**
```sql
-- Re-enable bootstrap tenant
UPDATE tenants 
SET status = 'active', updated_at = NOW()
WHERE id = 'system';
```

## 🔍 Verification Queries

### Check Bootstrap Status

```sql
-- Check if bootstrap tenant exists
SELECT id, name, status, settings->>'is_bootstrap' as is_bootstrap
FROM tenants
WHERE id = 'system';

-- Check super admin user
SELECT id, username, email, status
FROM users
WHERE tenant_id = 'system' AND id = 'super-admin';

-- Check super admin permissions
SELECT rp.role_id, p.resource, p.action
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.tenant_id = 'system' 
  AND rp.app_id = 'admin-console'
  AND rp.role_id = 'super-admin-role';

-- Check all tenants
SELECT id, name, status, created_at
FROM tenants
ORDER BY created_at;
```

### Verify Tenant Isolation

```sql
-- Count users per tenant
SELECT tenant_id, COUNT(*) as user_count
FROM users
WHERE status = 'active'
GROUP BY tenant_id;

-- Count apps per tenant
SELECT tenant_id, COUNT(*) as app_count
FROM apps
WHERE status = 'active'
GROUP BY tenant_id;

-- Verify no cross-tenant data leak
SELECT DISTINCT tenant_id 
FROM users 
WHERE username = 'alice'; -- Should only return acme-corp
```

## ⚠️ Common Issues

### Issue 1: Bootstrap Already Ran

**Error:** `Database already initialized with tenants`

**Cause:** Bootstrap hanya jalan sekali (database kosong)

**Solution:**
```sql
-- Drop and recreate database (DEV ONLY!)
DROP DATABASE lokstra_auth;
CREATE DATABASE lokstra_auth;
-- Re-run migrations
```

### Issue 2: Password Hash Error

**Error:** `Invalid password hash`

**Cause:** Password hashing belum diimplementasi

**Solution:**
Update `deployment/bootstrap.go` line dengan actual bcrypt:
```go
import "golang.org/x/crypto/bcrypt"

func hashPasswordBootstrap(password string) string {
    hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(hashed)
}
```

### Issue 3: Super Admin No Permission

**Error:** `403 Forbidden - Insufficient permissions`

**Cause:** Permissions tidak ter-assign ke role

**Solution:**
```sql
-- Manually grant wildcard permission
INSERT INTO role_permissions (role_id, tenant_id, app_id, permission_id, created_at)
VALUES ('super-admin-role', 'system', 'admin-console', 'perm_wildcard', NOW())
ON CONFLICT DO NOTHING;
```

### Issue 4: Bootstrap Tenant Login Fails

**Error:** `Tenant not found or inactive`

**Cause:** Bootstrap tenant di-suspend terlalu cepat

**Solution:**
```sql
-- Re-activate bootstrap tenant
UPDATE tenants SET status = 'active' WHERE id = 'system';
```

## 📚 Related Documentation

- [Bootstrap Guide](../../docs/BOOTSTRAP.md) - Complete bootstrap documentation
- [Multi-Tenant Architecture](../../docs/multi_tenant_architecture.md) - Architecture overview
- [Security Best Practices](../../docs/security.md) - Production security guide
- [Tenant Management API](../../docs/tenant_management.md) - API reference

## 🔐 Security Checklist

Before going to production:

- [ ] Change `SUPER_ADMIN_PASSWORD` to strong password (16+ chars)
- [ ] Store password in secrets manager (not .env file)
- [ ] Disable bootstrap tenant after first real tenant created
- [ ] Rotate super admin password every 90 days
- [ ] Enable audit logging for super admin actions
- [ ] Restrict super admin access to specific IPs
- [ ] Setup MFA for super admin login
- [ ] Monitor bootstrap tenant status (should stay suspended)
- [ ] Remove `SUPER_ADMIN_PASSWORD` from env after setup
- [ ] Document super admin recovery procedure

## 📞 Support

Jika ada issues:

1. Check logs: `examples/01_deployment/logs/`
2. Verify database: Run verification queries above
3. Check config: `examples/01_deployment/config/`
4. Review documentation: `docs/BOOTSTRAP.md`
