# Bootstrap & Authorization Architecture

## 📋 **Table of Contents**
1. [Overview](#overview)
2. [Platform vs Tenant Hierarchy](#platform-vs-tenant-hierarchy)
3. [Bootstrap Process](#bootstrap-process)
4. [Authorization Model](#authorization-model)
5. [Tenant Owner Privileges](#tenant-owner-privileges)
6. [Implementation Guide](#implementation-guide)
7. [Security Considerations](#security-considerations)

---

## 🎯 **Overview**

Lokstra-auth uses a **two-tier authorization model**:

1. **Platform Level**: Super admins manage all tenants
2. **Tenant Level**: Tenant admins manage their own tenant

This solves the "chicken-and-egg" bootstrap problem where you need auth to create tenants, but need tenants to have users.

---

## 🏗️ **Platform vs Tenant Hierarchy**

```
┌──────────────────────────────────────────────────────────────┐
│ PLATFORM LEVEL (Global Scope)                                │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Tenant: "platform"                                          │
│  ├─ App: "platform-admin-app"                               │
│  └─ Users:                                                    │
│      └─ platform-admin (role: platform-admin)               │
│                                                               │
│  Permissions:                                                 │
│  ✅ Create/Read/Update/Delete ANY tenant                     │
│  ✅ View cross-tenant statistics                             │
│  ✅ Suspend/activate tenants                                 │
│  ❌ Access tenant-specific resources directly                │
│                                                               │
└──────────────────────────────────────────────────────────────┘
           │
           │ creates tenants
           ▼
┌──────────────────────────────────────────────────────────────┐
│ TENANT LEVEL (Isolated Scope)                                │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Tenant: "acme-corp"                                         │
│  ├─ App: "acme-corp-admin" (default)                        │
│  │   └─ Users:                                               │
│  │       ├─ admin (role: tenant-admin, is_tenant_owner: true)│
│  │       └─ alice (role: manager)                            │
│  │                                                            │
│  └─ App: "web-portal"                                        │
│      └─ Users:                                                │
│          ├─ admin (role: tenant-admin)                       │
│          ├─ alice (role: manager)                            │
│          └─ bob (role: viewer)                               │
│                                                               │
│  Tenant Admin Permissions:                                    │
│  ✅ Manage users/apps/roles within "acme-corp"              │
│  ✅ Update tenant settings                                   │
│  ❌ See or modify other tenants                              │
│                                                               │
│  Tenant Owner (is_tenant_owner=true):                        │
│  ✅ Implicit tenant-admin role                               │
│  ✅ Cannot be deleted (unless ownership transferred)          │
│  ✅ Can transfer ownership to another admin                  │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

---

## 🚀 **Bootstrap Process**

### **Step 1: Initial Database Migration**

Run the bootstrap migration **once** during initial deployment:

```bash
psql -U postgres -d lokstra_db -f examples/01_deployment/migrations/01_lokstra_auth/001_bootstrap_platform_admin.sql
```

This creates:
- ✅ Platform tenant (`platform`)
- ✅ Platform admin app (`platform-admin-app`)
- ✅ Platform admin user (`platform-admin`)
- ✅ Platform permissions & roles

**Default Credentials:**
```
Tenant:   platform
Username: platform-admin
Password: PlatformAdmin@2025!
```

⚠️ **CRITICAL**: Change this password immediately after first login!

---

### **Step 2: Login as Platform Admin**

```http
POST /api/auth/login
Content-Type: application/json

{
  "tenant_id": "platform",
  "username": "platform-admin",
  "password": "PlatformAdmin@2025!",
  "app_id": "platform-admin-app"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": "platform-admin-user",
    "username": "platform-admin",
    "tenant_id": "platform",
    "metadata": {
      "is_platform_admin": true
    }
  }
}
```

---

### **Step 3: Create First Tenant with Auto-Admin**

```http
POST /api/auth/core/tenants
Authorization: Bearer <platform-admin-token>
Content-Type: application/json

{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "domain": "acme-corp.com",
  "db_dsn": "postgres://postgres:pwd@localhost:5432/lokstra_db?sslmode=disable",
  "db_schema": "acme_corp",
  "settings": {
    "max_users": 100,
    "max_apps": 10
  },
  "admin": {
    "username": "admin",
    "email": "admin@acme-corp.com",
    "password": "AcmeAdmin@2025!",
    "full_name": "Acme Administrator"
  }
}
```

**Response:**
```json
{
  "tenant": {
    "id": "acme-corp",
    "name": "Acme Corporation",
    "status": "active"
  },
  "default_app": {
    "id": "acme-corp-admin",
    "name": "Acme Corporation Admin",
    "type": "admin"
  },
  "admin_user": {
    "id": "acme-corp-admin-user",
    "username": "admin",
    "email": "admin@acme-corp.com",
    "metadata": {
      "is_tenant_owner": true
    }
  }
}
```

---

### **Step 4: Tenant Admin Logs In**

```http
POST /api/auth/login
Content-Type: application/json

{
  "tenant_id": "acme-corp",
  "username": "admin",
  "password": "AcmeAdmin@2025!",
  "app_id": "acme-corp-admin"
}
```

Now tenant admin can manage their tenant!

---

## 🔐 **Authorization Model**

### **Platform Admin Authorization**

**Middleware Stack:**
```go
// @RouterService middlewares=["recovery", "request_logger", "auth", "platform_admin"]
```

**Middleware Chain:**
1. `auth` - Validates JWT, extracts subject
2. `platform_admin` - Checks:
   - `subject.TenantID == "platform"`
   - `subject.Roles` contains `"platform-admin"`

**Allowed Operations:**
- ✅ `POST /api/auth/core/tenants` - Create tenant
- ✅ `GET /api/auth/core/tenants` - List all tenants
- ✅ `PUT /api/auth/core/tenants/{id}` - Update any tenant
- ✅ `DELETE /api/auth/core/tenants/{id}` - Delete any tenant
- ❌ `POST /api/auth/core/tenants/{id}/apps` - Cannot access tenant resources directly

---

### **Tenant Admin Authorization**

**Middleware Stack:**
```go
// @RouterService middlewares=["recovery", "request_logger", "auth", "tenant_scope"]
```

**Middleware Chain:**
1. `auth` - Validates JWT, extracts subject
2. `tenant_scope` - Checks:
   - `subject.TenantID == {tenant_id}` from path
   - `subject.Roles` contains `"tenant-admin"` OR `"tenant-owner"`

**Allowed Operations:**
- ✅ `GET /api/auth/core/tenants/{tenant_id}` - View own tenant
- ✅ `POST /api/auth/core/tenants/{tenant_id}/users` - Create users in tenant
- ✅ `POST /api/auth/core/tenants/{tenant_id}/apps` - Create apps in tenant
- ❌ `GET /api/auth/core/tenants` - Cannot list all tenants
- ❌ `DELETE /api/auth/core/tenants/{other_tenant_id}` - Cannot touch other tenants

---

## 👑 **Tenant Owner Privileges**

### **What is Tenant Owner?**

The **first admin user** created when a tenant is bootstrapped. Identified by:

```json
{
  "user": {
    "metadata": {
      "is_tenant_owner": true
    }
  }
}
```

---

### **Implicit Privileges**

**TenantOwnerMiddleware** automatically grants:

1. **Implicit `tenant-owner` role** (even if not in DB)
2. **Implicit `tenant-admin` role** (full tenant access)
3. **Cannot be deleted** (must transfer ownership first)

```go
// @RouterService middlewares=["recovery", "request_logger", "auth", "tenant_owner", "tenant_scope"]
```

**Middleware Flow:**
```
Request → auth → tenant_owner → tenant_scope → handler
          ↓           ↓              ↓
          Extract   Check         Validate
          subject   is_tenant_    tenant
                    owner=true    matches
                       ↓
                    Add implicit
                    tenant-owner
                    tenant-admin
                    roles
```

---

### **Implementation**

```go
// middleware/tenant_owner.go
func TenantOwnerMiddleware(checker TenantOwnerChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, _ := r.Context().Value("subject").(token.Subject)
			
			// Check metadata.is_tenant_owner
			isOwner, _ := checker.IsTenantOwner(r.Context(), subject.TenantID, subject.UserID)
			
			if isOwner {
				// Add implicit roles
				enhancedSubject := subject
				enhancedSubject.Roles = append(enhancedSubject.Roles, "tenant-owner", "tenant-admin")
				
				ctx := context.WithValue(r.Context(), "subject", enhancedSubject)
				r = r.WithContext(ctx)
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
```

---

### **Ownership Transfer**

Only tenant owner can transfer ownership to another admin:

```http
POST /api/auth/core/tenants/{tenant_id}/transfer-ownership
Authorization: Bearer <tenant-owner-token>
Content-Type: application/json

{
  "new_owner_user_id": "another-admin-user-id"
}
```

**Business Logic:**
1. ✅ Verify current user has `is_tenant_owner: true`
2. ✅ Verify new owner has `tenant-admin` role
3. ✅ Set old owner `is_tenant_owner: false`
4. ✅ Set new owner `is_tenant_owner: true`
5. ✅ Audit log the transfer

---

## 📝 **Implementation Guide**

### **1. Run Bootstrap Migration**

```bash
# First time setup only
psql -U postgres -d lokstra_db -f examples/01_deployment/migrations/01_lokstra_auth/001_bootstrap_platform_admin.sql
```

---

### **2. Register Middlewares**

```go
// infrastructure/register.go
func RegisterMiddlewares(router *mux.Router) {
	// Platform admin middleware
	router.Use(middleware.PlatformAdminMiddleware)
	
	// Tenant owner middleware
	ownerChecker := middleware.NewInMemoryTenantOwnerChecker()
	router.Use(middleware.TenantOwnerMiddleware(ownerChecker))
}
```

---

### **3. Add Bootstrap Service**

```go
// @RouterService name="bootstrap-service", prefix="/api/auth/bootstrap"
type BootstrapService struct {
	// ... repositories
}

// @Route "POST /"
// @Middleware ["recovery", "request_logger", "auth", "platform_admin"]
func (s *BootstrapService) CreateTenantWithAdmin(
	ctx context.Context,
	req CreateTenantWithAdminRequest,
) (*CreateTenantWithAdminResponse, error) {
	// Atomic transaction:
	// 1. Create tenant
	// 2. Create admin app
	// 3. Create admin user with is_tenant_owner=true
	// 4. Grant user access to app
	// 5. Assign tenant-admin role
}
```

---

### **4. Update TenantService**

```go
// @RouterService middlewares=["recovery", "request_logger", "auth", "platform_admin"]
type TenantService struct {
	// Now requires platform admin
}

// Regular users cannot create tenants
// Only platform-admin can
```

---

## 🔒 **Security Considerations**

### **1. Password Security**

❌ **NEVER** store plain passwords:

```go
// BAD
user.Password = req.Password

// GOOD
hashedPassword := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
user.PasswordHash = string(hashedPassword)
```

---

### **2. Account Lockout Protection**

Protect against brute-force attacks with automatic account lockout:

**Configuration (in Tenant Settings):**
```go
AccountLockout{
	Enabled:            true,
	MaxAttempts:        5,                // Lock after 5 failed attempts
	LockoutDuration:    15 * time.Minute, // Lock for 15 minutes
	ResetAttemptsAfter: 1 * time.Hour,    // Reset counter after 1 hour
	NotifyOnLockout:    true,             // Send email notification
}
```

**How it works:**
1. User enters wrong password → `FailedLoginAttempts` increments
2. After `MaxAttempts` failures → Account locked for `LockoutDuration`
3. During lockout → All login attempts fail with **generic error** (no hint that account is locked)
4. After `LockoutDuration` → Auto-unlock on next login attempt
5. On successful login → Reset `FailedLoginAttempts` to 0

**Security Benefits:**
- ✅ Prevents brute-force password guessing
- ✅ Limits automated attack attempts
- ✅ Generic errors prevent account enumeration
- ✅ Automatic unlock reduces admin overhead
- ✅ Email notifications alert legitimate users

**Database Schema:**
```sql
ALTER TABLE users ADD COLUMN failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN last_failed_login_at TIMESTAMP;
ALTER TABLE users ADD COLUMN locked_at TIMESTAMP;
ALTER TABLE users ADD COLUMN locked_until TIMESTAMP;
ALTER TABLE users ADD COLUMN lockout_count INT DEFAULT 0;
```

---

### **3. Generic Error Messages**

**Always return generic errors** for authentication failures:

```go
// ✅ CORRECT - Generic errors
if !passwordValid {
    return &TokenResponse{
        Success: false,
        Error: "invalid credentials", // Generic
    }
}

if user.IsLocked() {
    return &TokenResponse{
        Success: false,
        Error: "invalid credentials", // Don't reveal lock status
    }
}

// ❌ WRONG - Specific errors leak information
if !passwordValid {
    return errors.New("incorrect password") // Reveals username exists!
}

if user.IsLocked() {
    return errors.New("account is locked") // Confirms account exists!
}
```

**Why Generic Errors:**
- Prevents **user enumeration** (attacker can't tell if user exists)
- Prevents **account state disclosure** (attacker can't tell if locked/suspended)
- Prevents **password validation hints** (attacker can't tell if password is close)

---

### **4. Password Reset Security**

**Forgot Password Flow:**
```http
POST /api/auth/cred/basic/forgot-password
{
  "email": "user@example.com"
}

Response (ALWAYS):
{
  "success": true,
  "message": "If the email exists, a password reset link has been sent"
}
```

**Security Measures:**
- ✅ Same response for valid/invalid emails (prevents enumeration)
- ✅ Short-lived reset tokens (15 minutes expiry)
- ✅ Single-use tokens (revoked after password reset)
- ✅ Token includes type claim: `"type": "password_reset"`
- ✅ Reset only via email (no SMS/phone fallback to prevent SIM swapping)

**Reset Password Flow:**
```http
POST /api/auth/cred/basic/reset-password
{
  "reset_token": "eyJhbGciOiJIUzI1NiI...",
  "new_password": "NewSecurePassword123!"
}
```

**Token Validation:**
1. Verify token signature
2. Check token type == "password_reset"
3. Check expiry (max 15 minutes)
4. Check not already used
5. Hash new password
6. Update database
7. Revoke token to prevent reuse

---

### **5. Default Credentials**

⚠️ **Force password change** on first login:

```json
{
  "metadata": {
    "force_password_change": true
  }
}
```

```go
// In auth/login handler
if user.Metadata["force_password_change"] == true {
	return &LoginResponse{
		Message: "Password change required",
		RequirePasswordChange: true,
	}
}
```

---

### **3. Platform Admin Scope**

Platform admins should **NOT** have direct access to tenant resources:

```
✅ GET /api/auth/core/tenants/{id}           - View tenant metadata
❌ GET /api/auth/core/tenants/{id}/users     - Cannot list tenant users
❌ POST /api/auth/core/tenants/{id}/apps     - Cannot create apps
```

If platform admin needs to manage tenant, they must:
1. Create a regular user in that tenant
2. Assign tenant-admin role
3. Login as that user

---

### **4. Tenant Isolation**

PostgreSQL schema-per-tenant ensures:

```sql
-- acme_corp schema
CREATE SCHEMA acme_corp;
SET search_path TO acme_corp;

-- Widget Corp cannot access acme_corp data
CREATE SCHEMA widget_corp;
SET search_path TO widget_corp;
```

---

### **5. Audit Logging**

Log all platform admin actions:

```go
// middleware/audit.go
type AuditLog struct {
	TenantID  string
	UserID    string
	Action    string    // "tenant:create", "tenant:delete"
	Resource  string    // "tenant/acme-corp"
	IPAddress string
	Timestamp time.Time
}
```

---

## 🧪 **Testing**

See complete API tests in:
```
examples/01_deployment/http-tests/platform/01-platform-admin.http
```

**Test Coverage:**
1. ✅ Platform admin login
2. ✅ Create tenant with auto-admin
3. ✅ Tenant admin login
4. ✅ Verify tenant isolation (admin cannot see other tenants)
5. ✅ Verify tenant owner implicit privileges
6. ✅ Change default passwords
7. ✅ Transfer ownership

---

## 📚 **Related Documentation**

- [Multi-Tenant Architecture](./multi_tenant_architecture.md)
- [RBAC Implementation](./OWNERSHIP_AND_ROLES.md)
- [Credential Configuration](./credential_configuration.md)
- [Migration Management](./MIGRATION_MANAGEMENT.md)

---

## 🎓 **FAQ**

### **Q: Can I have multiple platform admins?**

✅ Yes! Create additional users in "platform" tenant and assign "platform-admin" role.

---

### **Q: Can tenant owner role be revoked?**

❌ No. Ownership must be **transferred** to another tenant-admin, not revoked.

---

### **Q: What happens if platform admin is locked out?**

🆘 Recovery process:
1. Access database directly
2. Reset platform-admin password hash
3. Or create new platform admin via SQL

```sql
-- Emergency: Create new platform admin
INSERT INTO users (id, tenant_id, username, password_hash, status)
VALUES ('emergency-admin', 'platform', 'emergency', 'hashed-pwd', 'active');
```

---

### **Q: Can tenant admins create apps in their tenant?**

✅ Yes! Tenant admins have full CRUD on:
- Users
- Apps
- Branches
- Roles
- Permissions (within their tenant)

---

### **Q: Is platform tenant isolated from regular tenants?**

✅ Yes! Platform tenant uses schema `"platform"`, completely isolated from `"acme_corp"`, `"widget_corp"`, etc.

---

## ✅ **Summary**

| Aspect | Platform Admin | Tenant Owner | Tenant Admin | Regular User |
|--------|---------------|--------------|--------------|--------------|
| **Tenant** | `platform` | Any tenant | Any tenant | Any tenant |
| **Role** | `platform-admin` | `tenant-owner` (implicit) | `tenant-admin` | Custom roles |
| **Can CRUD tenants** | ✅ All | ❌ | ❌ | ❌ |
| **Can CRUD users** | ❌ | ✅ Own tenant | ✅ Own tenant | ❌ |
| **Can CRUD apps** | ❌ | ✅ Own tenant | ✅ Own tenant | Depends on role |
| **Can be deleted** | ✅ | ❌ (must transfer) | ✅ | ✅ |
| **Bootstrap method** | SQL migration | Auto-created | Manual | Manual |

---

**Created:** 2025-01-XX  
**Last Updated:** 2025-01-XX  
**Version:** 1.0
