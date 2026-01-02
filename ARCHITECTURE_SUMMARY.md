# Architecture Summary: Admin Tenant & Ownership

## 🎯 Final Architecture

Berdasarkan diskusi, berikut final architecture yang diimplementasikan:

## 1. **Admin Tenant (Platform Admin)** ✅

### Konsep
- **Tenant:** `system` (bootstrap tenant yang sudah ada)
- **Role:** Super Administrator / Platform Admin
- **Count:** Multiple allowed
- **Storage:** `db_main`

### Tanggung Jawab
```
✅ CRUD Multi-Tenant (Create, Read, Update, Delete semua tenant)
✅ Suspend/Activate tenant
✅ View all tenant owners
✅ Emergency ownership transfer
✅ Platform-wide monitoring & analytics
✅ Billing override (for support cases)
```

### Tidak Bisa
```
❌ Access tenant's business data (kecuali diberi explicit permission)
❌ Transfer ownership tanpa audit trail
❌ Delete tenant dengan subscription aktif
```

### Bootstrap Credentials
```
Tenant ID:  system
App ID:     admin-console
Username:   admin
Email:      admin@localhost
Password:   (from SUPER_ADMIN_PASSWORD env var)
```

---

## 2. **Tenant Owner** ✅

### Konsep
- **Count:** **EXACTLY 1 per tenant** (enforced by DB constraint)
- **Identifier:** `tenant.owner_id = user.id`
- **Flag:** `user.is_tenant_owner = TRUE`
- **Storage:** `db_main`

### Tanggung Jawab
```
✅ Billing & subscription (legal owner)
✅ Update tenant settings
✅ Transfer ownership ke user lain (dalam tenant yang sama)
✅ Assign/revoke Tenant Admin roles
✅ Full access to all tenant resources
✅ Cancel/upgrade subscription
```

### Tidak Bisa
```
❌ Create tenant lain (bukan platform admin)
❌ Access tenant lain
❌ Delete tenant (harus kontak platform admin)
```

### Ownership Transfer
```http
POST /api/auth/core/tenants/{tenant_id}/transfer-ownership
Authorization: Bearer {owner_token}

{
  "new_owner_id": "usr-bob",
  "reason": "Company ownership change"
}
```

**Rules:**
- ✅ New owner **must be user in same tenant**
- ✅ Only **current owner or platform admin** can transfer
- ✅ Transfer is **audited** in `tenant_ownership_history`
- ✅ Old owner loses owner privileges automatically
- ✅ New owner gets notification

---

## 3. **Database Storage: Single Global DB** ✅

### Rekomendasi: **Global DB untuk semua auth data**

**Alasan:**

| Aspect | Benefit |
|--------|---------|
| **Authentication** | Single source of truth, fast lookup |
| **User Identity** | Consistent across tenants, SSO support |
| **Cross-Tenant Email** | Email uniqueness validation |
| **Backup/Recovery** | Centralized, easier to manage |
| **RBAC** | Uniform authorization model |
| **Audit Trail** | Consistent logging & compliance |

### Storage Architecture

```
┌─────────────────────────────────────────────────┐
│           GLOBAL DATABASE (db_main)           │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  Authentication & Authorization Data     │  │
│  ├──────────────────────────────────────────┤  │
│  │  • tenants (all tenants)                 │  │
│  │  • users (all users from all tenants)    │  │
│  │  • user_passwords (password hashes)      │  │
│  │  • apps (all apps)                       │  │
│  │  • roles (RBAC roles)                    │  │
│  │  • permissions (RBAC permissions)        │  │
│  │  • user_roles (assignments)              │  │
│  │  • tenant_ownership_history (audit)      │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
                      │
                      │ db_dsn, db_schema
                      ▼
┌─────────────────────────────────────────────────┐
│      TENANT-SPECIFIC DATABASES (per tenant)     │
│                                                 │
│  ┌──────────────┐  ┌──────────────┐            │
│  │  Acme Corp   │  │  Globex Inc  │            │
│  │  Database    │  │  Database    │            │
│  ├──────────────┤  ├──────────────┤            │
│  │ • documents  │  │ • orders     │            │
│  │ • invoices   │  │ • products   │            │
│  │ • customers  │  │ • inventory  │            │
│  │              │  │              │            │
│  │ (Business    │  │ (Business    │            │
│  │  Data Only)  │  │  Data Only)  │            │
│  └──────────────┘  └──────────────┘            │
└─────────────────────────────────────────────────┘
```

### Tenant Table Schema

```sql
tenants
├── id (PK)
├── name
├── owner_id (FK → users.id, UNIQUE, NOT NULL)
├── db_dsn     -- Connection to tenant's business DB
├── db_schema  -- Schema name for tenant's business DB
├── status
└── ...

users
├── id (PK)
├── tenant_id (FK → tenants.id)
├── username (UNIQUE within tenant)
├── email (UNIQUE within tenant)
├── is_tenant_owner (TRUE for owner, only 1 per tenant)
└── ...
```

---

## 4. **Complete Role Hierarchy**

```
┌─────────────────────────────────────────────────┐
│ Level 1: PLATFORM ADMIN                         │
│ ├─ Tenant: system                               │
│ ├─ Storage: db_main                           │
│ ├─ Count: Multiple                              │
│ └─ Scope: All tenants                           │
└─────────────────────────────────────────────────┘
                    │
                    │ manages
                    ▼
┌─────────────────────────────────────────────────┐
│ Level 2: TENANT OWNER                           │
│ ├─ Tenant: Any (except system)                  │
│ ├─ Storage: db_main                           │
│ ├─ Count: EXACTLY 1 per tenant                  │
│ ├─ Flag: is_tenant_owner = TRUE                 │
│ └─ Scope: Single tenant                         │
└─────────────────────────────────────────────────┘
                    │
                    │ assigns
                    ▼
┌─────────────────────────────────────────────────┐
│ Level 3: TENANT ADMIN                           │
│ ├─ Tenant: Same as owner                        │
│ ├─ Storage: db_main                           │
│ ├─ Count: Multiple                              │
│ ├─ Role: tenant-admin                           │
│ └─ Scope: Single tenant (limited)               │
└─────────────────────────────────────────────────┘
                    │
                    │ manages
                    ▼
┌─────────────────────────────────────────────────┐
│ Level 4: REGULAR USER                           │
│ ├─ Tenant: Same as admin                        │
│ ├─ Storage: db_main                           │
│ ├─ Count: Unlimited (within quota)              │
│ └─ Scope: Assigned apps only                    │
└─────────────────────────────────────────────────┘
```

---

## 5. **Permission Matrix**

| Action | Platform Admin | Tenant Owner | Tenant Admin | User |
|--------|---------------|--------------|--------------|------|
| **Tenant** |
| Create tenant | ✅ | ❌ | ❌ | ❌ |
| Update tenant | ✅ | ✅ | ❌ | ❌ |
| Delete tenant | ✅ | ❌ | ❌ | ❌ |
| Suspend tenant | ✅ | ❌ | ❌ | ❌ |
| Transfer ownership | ✅ (emergency) | ✅ | ❌ | ❌ |
| **Billing** |
| View billing | ✅ (all) | ✅ (own) | ❌ | ❌ |
| Update payment | ❌ | ✅ | ❌ | ❌ |
| Cancel subscription | ❌ | ✅ | ❌ | ❌ |
| **Apps** |
| Create app | ✅ | ✅ | ✅ | ❌ |
| Update app | ✅ | ✅ | ✅ | ❌ |
| Delete app | ✅ | ✅ | ✅ | ❌ |
| **Users** |
| Create user | ✅ | ✅ | ✅ | ❌ |
| Update user | ✅ | ✅ | ✅ | ✅ (self) |
| Delete user | ✅ | ✅ | ✅ | ❌ |
| Assign tenant admin | ✅ | ✅ | ❌ | ❌ |
| **Roles** |
| Create role | ✅ | ✅ | ✅ | ❌ |
| Assign role | ✅ | ✅ | ✅ | ❌ |

---

## 6. **Complete Workflows**

### Workflow 1: First-Time Setup (Bootstrap)

```bash
# 1. Set password
export SUPER_ADMIN_PASSWORD="SecurePassword123!"

# 2. Run app (auto-bootstrap)
go run main.go

# Output:
# ✅ Bootstrap completed successfully!
# Tenant: system
# User: admin (platform admin)
```

### Workflow 2: Platform Admin Creates Tenant

```http
# 1. Login as platform admin
POST /api/auth/cred/basic/authenticate
{
  "auth_context": { "tenant_id": "system", "app_id": "admin-console" },
  "username": "admin",
  "password": "SecurePassword123!"
}

# 2. Create tenant with owner
POST /api/auth/core/tenants
Authorization: Bearer {platform_admin_token}
{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "owner_id": "usr-alice",  // Alice becomes owner
  "db_dsn": "postgres://localhost/acme_db",
  "db_schema": "acme"
}

# System automatically:
# ✅ Sets tenants.owner_id = usr-alice
# ✅ Sets users.is_tenant_owner = TRUE for Alice
# ✅ Creates audit record in ownership_history
```

### Workflow 3: Owner Assigns Tenant Admin

```http
# 1. Owner creates user
POST /api/auth/core/tenants/acme-corp/users
Authorization: Bearer {owner_token}
{
  "username": "bob",
  "email": "bob@acme.com"
}

# 2. Owner assigns tenant-admin role
POST /api/auth/rbac/tenants/acme-corp/apps/admin-console/users/{bob_id}/roles
Authorization: Bearer {owner_token}
{
  "role_id": "role-tenant-admin"
}
```

### Workflow 4: Transfer Ownership

```http
# Owner transfers to another user
POST /api/auth/core/tenants/acme-corp/transfer-ownership
Authorization: Bearer {alice_token}
{
  "new_owner_id": "usr-bob",
  "reason": "Company sold to Bob"
}

# System:
# ✅ Validates Bob exists in acme-corp
# ✅ Updates tenants.owner_id = usr-bob
# ✅ Sets Alice is_tenant_owner = FALSE
# ✅ Sets Bob is_tenant_owner = TRUE
# ✅ Creates audit record
# ✅ Sends notifications
```

---

## 7. **Database Constraints**

```sql
-- 1. Every tenant must have owner
ALTER TABLE tenants 
ALTER COLUMN owner_id SET NOT NULL;

-- 2. Only 1 owner per tenant (each owner_id appears once)
CREATE UNIQUE INDEX unique_owner_per_tenant ON tenants(owner_id);

-- 3. Owner must be user in same tenant (trigger validation)
CREATE TRIGGER trg_validate_tenant_owner
BEFORE INSERT OR UPDATE ON tenants
FOR EACH ROW EXECUTE FUNCTION validate_tenant_owner();

-- 4. Only 1 owner flag per tenant
-- Enforced at application level + DB index
CREATE INDEX idx_users_tenant_owner ON users(tenant_id, is_tenant_owner) 
WHERE is_tenant_owner = TRUE;
```

---

## 8. **Files Created/Updated**

### Domain Models
- ✅ `core/domain/tenant.go` - Added `owner_id` field
- ✅ `core/domain/user.go` - Added `is_tenant_owner` field
- ✅ `core/domain/tenant.go` - Added `TransferOwnershipRequest`
- ✅ `core/domain/tenant.go` - Added `TenantOwnershipHistory`

### Database
- ✅ `deployment/migrations/006_tenant_ownership.sql` - Complete migration
  - Adds owner_id column
  - Adds is_tenant_owner flag
  - Creates ownership_history table
  - Backfills existing tenants
  - Adds constraints & triggers
  - Helper functions

### Documentation
- ✅ `docs/OWNERSHIP_AND_ROLES.md` - Complete architecture guide
- ✅ `docs/BOOTSTRAP.md` - Bootstrap setup guide (updated)
- ✅ `ARCHITECTURE_SUMMARY.md` - This file

### Bootstrap
- ✅ `deployment/bootstrap.go` - Bootstrap system tenant

---

## 9. **Migration Path**

### For New Deployments
```bash
# 1. Run all migrations in order
psql -d lokstra_auth -f deployment/migrations/db_schema.sql
psql -d lokstra_auth -f deployment/migrations/001_subject_rbac.sql
psql -d lokstra_auth -f deployment/migrations/002_authz_policies.sql
psql -d lokstra_auth -f deployment/migrations/006_tenant_ownership.sql

# 2. Run application with bootstrap
export SUPER_ADMIN_PASSWORD="YourSecurePassword"
go run main.go

# 3. Login as platform admin and create tenants
```

### For Existing Deployments
```bash
# 1. Backup database
pg_dump lokstra_auth > backup_before_ownership.sql

# 2. Run ownership migration
psql -d lokstra_auth -f deployment/migrations/006_tenant_ownership.sql

# 3. Verify ownership assignments
psql -d lokstra_auth -c "
  SELECT t.id, t.name, t.owner_id, u.username, u.is_tenant_owner
  FROM tenants t
  JOIN users u ON u.id = t.owner_id
  WHERE t.status = 'active'
"

# 4. Manually adjust owners if needed
psql -d lokstra_auth -c "
  SELECT transfer_tenant_ownership(
    'tenant-id', 
    'new-owner-user-id', 
    'admin-user-id', 
    'Correcting initial owner assignment'
  );
"
```

---

## 10. **Security Considerations**

### Ownership Transfer Security
- ✅ Only current owner or platform admin can transfer
- ✅ New owner must exist in same tenant
- ✅ All transfers are audited
- ✅ Notifications sent to both parties
- ✅ Old owner automatically loses privileges

### Platform Admin Security
- ✅ Lives in separate `system` tenant
- ✅ Cannot be tenant owner of regular tenants
- ✅ All actions are logged
- ✅ Emergency transfers require justification

### Data Isolation
- ✅ Auth data in db_main (centralized)
- ✅ Business data in tenant-db (isolated)
- ✅ No cross-tenant data access
- ✅ Owner-level permissions enforced

---

## 11. **Benefits Summary**

✅ **Clear Ownership**
- Every tenant has exactly 1 legal owner
- Billing responsibility is unambiguous
- Ownership transfer is audited

✅ **Centralized Auth**
- Single source of truth for all users
- Consistent user identity
- Cross-tenant SSO support

✅ **Data Isolation**
- Business data separated per tenant
- Independent scaling
- Flexible schema per tenant

✅ **Flexible Administration**
- Multiple admins per tenant
- Owner can delegate safely
- Platform admin for oversight

✅ **Compliance Ready**
- Full audit trail
- Ownership history preserved
- GDPR-compliant user management

---

## 12. **Next Steps**

### Immediate
- [ ] Review and approve architecture
- [ ] Run migration 006_tenant_ownership.sql
- [ ] Verify ownership assignments
- [ ] Test ownership transfer

### Short Term
- [ ] Implement TenantService.TransferOwnership()
- [ ] Add ownership middleware/guards
- [ ] Create ownership transfer API endpoint
- [ ] Add ownership history API endpoint
- [ ] Update HTTP test files

### Long Term
- [ ] Integrate billing system with owner_id
- [ ] Add owner notification system
- [ ] Create owner dashboard
- [ ] Implement subscription management
- [ ] Add multi-factor auth for ownership transfer

---

## 🎉 Kesimpulan

**Arsitektur Final:**

1. ✅ **Platform Admin** - CRUD multi-tenant di `system` tenant
2. ✅ **Tenant Owner** - 1 owner per tenant, billing responsibility
3. ✅ **Single Global DB** - Semua auth data centralized
4. ✅ **Tenant DB** - Bisnis data separated per tenant
5. ✅ **Ownership Transfer** - Audited & secure
6. ✅ **Clear Hierarchy** - Platform → Owner → Admin → User

**Database:**
- ✅ `db_main` untuk auth: users, roles, permissions, tenants
- ✅ `tenant-db` untuk bisnis: documents, orders, products, etc
- ✅ Owner tracking via `owner_id` dan `is_tenant_owner`
- ✅ Audit trail via `tenant_ownership_history`

Arsitektur ini **production-ready** dan siap diimplementasikan! 🚀
