# Bootstrap System - Initial Setup Guide

## 🎯 Problem: Chicken-and-Egg

Ketika sistem auth pertama kali dijalankan:
- Belum ada tenant
- Belum ada user
- Belum ada role/permission
- **Bagaimana cara login untuk membuat yang pertama kali?**

## ✅ Solusi: Super Admin Bootstrap

Sistem ini menyediakan **3 pendekatan** untuk bootstrap:

---

## 1️⃣ Auto-Bootstrap (Recommended)

Sistem otomatis membuat **bootstrap tenant + super admin** jika database kosong.

### Setup

```go
// main.go
func main() {
    // ... setup database ...
    
    // Bootstrap akan otomatis check database
    // Jika kosong, create super admin
    bootstrapped, err := deployment.BootstrapSystem(deployment.BootstrapConfig{
        EnableAutoBootstrap: true,  // ← Enable auto-bootstrap
        SuperAdminPassword:  "",    // Akan baca dari env var
        KeepBootstrapTenant: false, // Auto-disable setelah tenant pertama dibuat
    })
    if err != nil {
        log.Fatalf("Bootstrap failed: %v", err)
    }
    
    // ... start server ...
}
```

### Environment Variable

```bash
# Set password untuk super admin
export SUPER_ADMIN_PASSWORD="YourSecurePassword123!@#"

# Atau di .env file
SUPER_ADMIN_PASSWORD=YourSecurePassword123!@#
```

### Cara Kerja

1. **Database Kosong?**
   - ✅ Ya → Create bootstrap tenant + super admin
   - ❌ Tidak → Skip bootstrap (sudah ada tenant)

2. **Bootstrap Creates:**
   - Tenant: `system`
   - App: `admin-console`
   - User: `admin` (email: `admin@localhost`)
   - Role: `super-admin` dengan permission `*` (wildcard)

3. **Login Credentials:**
   ```
   Tenant ID:  system
   App ID:     admin-console
   Username:   admin
   Password:   (dari SUPER_ADMIN_PASSWORD env var)
   ```

4. **Workflow:**
   ```
   Start App → Check DB → Empty?
       ↓
   Create Bootstrap Tenant/User
       ↓
   Login sebagai Super Admin
       ↓
   Create Tenant Pertama (real tenant)
       ↓
   Bootstrap Tenant Auto-Disabled ✅
   ```

---

## 2️⃣ Manual Migration

Jika tidak mau auto-bootstrap, gunakan SQL migration.

### Run Migration

```bash
psql -U postgres -d lokstra_auth -f deployment/migrations/000_bootstrap.sql
```

### Update Password Hash

Edit migration file atau update langsung:

```sql
-- Generate bcrypt hash (cost=12) untuk password "Admin123!@#"
UPDATE user_passwords 
SET password_hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIiIkIiIkI'
WHERE tenant_id = 'system' AND user_id = 'super-admin';
```

**Generate Hash:**
```go
import "golang.org/x/crypto/bcrypt"

hash, _ := bcrypt.GenerateFromPassword([]byte("Admin123!@#"), 12)
fmt.Println(string(hash))
```

---

## 3️⃣ Hardcoded Bypass (Development Only)

Untuk development, bisa tambahkan bypass logic di middleware:

```go
// middleware/tenant.go
func TenantMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Params("tenant_id")
        
        // DEV ONLY: Bypass validation for system tenant
        if tenantID == "system" && os.Getenv("ENV") == "development" {
            c.Locals("tenant_id", tenantID)
            return c.Next()
        }
        
        // Normal validation
        tenant, err := tenantService.Get(tenantID)
        if err != nil {
            return c.Status(404).JSON(fiber.Map{"error": "tenant not found"})
        }
        
        c.Locals("tenant", tenant)
        return c.Next()
    }
}
```

⚠️ **NEVER use in production!**

---

## 📋 Complete Workflow

### Step 1: First Run (Database Empty)

```bash
# Set password
export SUPER_ADMIN_PASSWORD="SecurePass123!@#"

# Run application
go run main.go
```

**Output:**
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
╠════════════════════════════════════════════════════════╣
║ ⚠️  IMPORTANT: Change the super admin password ASAP!  ║
║ 💡 Create your first real tenant and disable bootstrap ║
╚════════════════════════════════════════════════════════╝
```

### Step 2: Login sebagai Super Admin

```http
POST /api/auth/cred/basic/authenticate
Content-Type: application/json

{
  "auth_context": {
    "tenant_id": "system",
    "app_id": "admin-console"
  },
  "username": "admin",
  "password": "SecurePass123!@#"
}
```

**Response:**
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

### Step 3: Create First Real Tenant

```http
POST /api/auth/core/tenants
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "Acme Corporation",
  "settings": {
    "max_users": 100,
    "max_apps": 10
  }
}
```

### Step 4: Create App untuk Tenant

```http
POST /api/auth/core/tenants/{tenant_id}/apps
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "Web Portal",
  "type": "web",
  "config": {
    "enable_basic": true,
    "enable_oauth2": false
  }
}
```

### Step 5: Create Admin User untuk Tenant Baru

```http
POST /api/auth/core/tenants/{tenant_id}/users
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "username": "alice",
  "email": "alice@acme.com",
  "full_name": "Alice Administrator",
  "password": "AliceSecure123!"
}
```

### Step 6: Assign Admin Role

```http
POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "Admin",
  "description": "Tenant administrator"
}
```

```http
POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "role_id": "{role_id}"
}
```

### Step 7: (Optional) Disable Bootstrap Tenant

```go
// Manually call or via API endpoint
err := deployment.DisableBootstrapTenant()
if err != nil {
    log.Printf("Failed to disable bootstrap: %v", err)
}
```

Or set `KeepBootstrapTenant: false` in config untuk auto-disable.

---

## 🔒 Security Best Practices

### 1. Change Default Password IMMEDIATELY

```http
PUT /api/auth/core/tenants/system/users/super-admin/password
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "old_password": "SecurePass123!@#",
  "new_password": "NewSuperSecurePassword456!@#$"
}
```

### 2. Disable Bootstrap Tenant After Setup

```bash
# In production, once you have real tenants:
curl -X POST http://localhost:8080/api/auth/admin/bootstrap/disable \
  -H "Authorization: Bearer {super_admin_token}"
```

### 3. Use Strong Password for Super Admin

```
Minimum Requirements:
- Length: 16+ characters
- Uppercase + Lowercase
- Numbers + Special chars
- No dictionary words
- Rotate every 90 days
```

### 4. Environment Variable Protection

```bash
# Never commit .env to git
echo ".env" >> .gitignore

# Use secrets manager in production
# AWS Secrets Manager, HashiCorp Vault, etc.
```

### 5. Audit Logging

```go
// Log all super admin actions
log.Printf("AUDIT: Super admin %s performed %s on %s", 
    userID, action, resource)
```

---

## ❓ FAQ

### Q: Apakah bootstrap tenant bisa dihapus?

**A:** Bisa, tapi **tidak recommended**. Lebih baik di-suspend agar history tetap ada.

```sql
UPDATE tenants SET status = 'suspended' WHERE id = 'system';
```

### Q: Bagaimana jika lupa password super admin?

**A:** Reset via database langsung:

```sql
-- Generate new hash dengan bcrypt
UPDATE user_passwords 
SET password_hash = '$2a$12$NewHashHere...', updated_at = NOW()
WHERE tenant_id = 'system' AND user_id = 'super-admin';
```

### Q: Apakah super admin bisa akses semua tenant?

**A:** Secara default **TIDAK**. Super admin hanya untuk:
1. Create tenant pertama
2. Setup initial configuration
3. Emergency recovery

Untuk cross-tenant access, perlu explicit grant.

### Q: Bisakah ada multiple super admin?

**A:** Ya, bisa create user baru di tenant `system`:

```http
POST /api/auth/core/tenants/system/users
Authorization: Bearer {super_admin_token}

{
  "username": "bob_admin",
  "email": "bob@localhost",
  "password": "BobSecure123!"
}
```

Then assign `super-admin` role.

### Q: Apakah bootstrap berjalan setiap kali app restart?

**A:** **TIDAK**. Bootstrap hanya jalan jika:
- `EnableAutoBootstrap = true` **DAN**
- Database **benar-benar kosong** (no tenants)

Jika sudah ada tenant, bootstrap di-skip.

---

## 🚀 Production Deployment Checklist

- [ ] Set `SUPER_ADMIN_PASSWORD` di environment (secrets manager)
- [ ] Enable `EnableAutoBootstrap` untuk deployment pertama
- [ ] Run migration atau biarkan auto-bootstrap
- [ ] Login dan ganti password super admin
- [ ] Create tenant pertama untuk production
- [ ] Create admin user untuk tenant tersebut
- [ ] Disable/suspend bootstrap tenant
- [ ] Set `EnableAutoBootstrap = false` untuk prevent re-run
- [ ] Remove `SUPER_ADMIN_PASSWORD` dari env (atau rotate)
- [ ] Setup monitoring untuk bootstrap tenant status
- [ ] Backup database setelah setup

---

## 📚 Related Files

- `deployment/bootstrap.go` - Bootstrap logic
- `deployment/migrations/000_bootstrap.sql` - Manual bootstrap migration
- `examples/01_deployment/main.go` - Example usage
- `docs/BOOTSTRAP.md` - This file

## 🔗 See Also

- [Multi-Tenant Architecture](./multi_tenant_architecture.md)
- [Tenant Management API](./tenant_management.md)
- [Security Best Practices](./security.md)
