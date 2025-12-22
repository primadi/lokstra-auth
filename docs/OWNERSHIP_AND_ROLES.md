# Ownership and Roles Architecture

## 🎯 Overview

Sistem ini menggunakan **3-tier role hierarchy** untuk multi-tenant management:

1. **Platform Admin** - Manages all tenants (CRUD tenants)
2. **Tenant Owner** - Owns a specific tenant (billing, legal owner)
3. **Tenant Admin** - Manages tenant resources (apps, users, roles)

## 📊 Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    PLATFORM LEVEL                           │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Platform Admin (system tenant)                    │    │
│  │  - CRUD all tenants                                │    │
│  │  - Suspend/activate tenants                        │    │
│  │  - View all tenant owners                          │    │
│  │  - Transfer ownership (emergency)                  │    │
│  │  - Global analytics & monitoring                   │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ manages
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    TENANT LEVEL                             │
│  ┌─────────────────────┐    ┌──────────────────────────┐   │
│  │   Tenant Owner      │    │   Tenant Admins          │   │
│  │   (1 per tenant)    │    │   (multiple)             │   │
│  │                     │    │                          │   │
│  │   - Billing         │    │   - Manage apps          │   │
│  │   - Subscription    │    │   - Manage users         │   │
│  │   - Legal owner     │    │   - Assign roles         │   │
│  │   - Transfer owner  │    │   - Create permissions   │   │
│  │   - Assign admins   │    │   - Cannot transfer      │   │
│  └─────────────────────┘    └──────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ contains
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    USER LEVEL                               │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Regular Users                                     │    │
│  │  - Access apps                                     │    │
│  │  - Use features                                    │    │
│  │  - Limited permissions                             │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## 🗄️ Database Storage Architecture

### Global Database (`db_main`)

**Stores ALL authentication & authorization data:**

```sql
-- Core Tables (in db_main)
tenants                 -- All tenant metadata
├── id
├── name
├── owner_id           -- FK to users.id (REQUIRED, UNIQUE per tenant)
├── db_dsn             -- Connection to tenant's business DB
└── db_schema

users                   -- ALL users from ALL tenants
├── id (global UUID)
├── tenant_id           -- FK to tenants.id
├── username            -- Unique within tenant
├── email               -- Unique within tenant
├── is_tenant_owner     -- TRUE for tenant owner (only 1 per tenant)
└── metadata

user_passwords          -- Password hashes
├── user_id
├── tenant_id
└── password_hash

apps                    -- All apps from all tenants
├── id
├── tenant_id
└── config

roles                   -- All roles from all tenants
├── id
├── tenant_id
├── app_id
└── name

permissions             -- All permissions from all tenants
├── id
├── tenant_id
├── app_id
├── resource
└── action

user_roles              -- User-role assignments
├── user_id
├── tenant_id
├── app_id
└── role_id

tenant_ownership_history  -- Ownership transfer audit
├── id
├── tenant_id
├── previous_owner
├── new_owner
├── transferred_by
├── reason
└── transferred_at
```

### Tenant-Specific Database (`tenant.db_dsn`)

**Stores ONLY business/application data:**

```sql
-- Example: Acme Corp's Business DB
documents
orders
invoices
products
customers
-- etc (business-specific tables)

-- NO auth tables here!
```

### Why This Design? ✅

| Aspect | Decision | Reason |
|--------|----------|--------|
| **User Storage** | Global DB | - Single source of truth<br>- Cross-tenant email uniqueness<br>- SSO support<br>- Easier backup |
| **Auth Data** | Global DB | - Centralized authentication<br>- Fast user lookup<br>- Consistent identity |
| **Business Data** | Tenant DB | - Data isolation<br>- Independent scaling<br>- Custom schema per tenant<br>- Tenant-specific backups |
| **Permissions** | Global DB | - Uniform RBAC model<br>- Easy cross-tenant reporting<br>- Consistent audit trail |

## 👥 Role Definitions

### 1. Platform Admin (Admin Tenant)

**Tenant:** `system` (bootstrap tenant)  
**Count:** Multiple allowed  
**Storage:** `db_main.users` where `tenant_id = 'system'`

**Permissions:**
```json
{
  "tenant": ["create", "read", "update", "delete", "suspend", "activate"],
  "tenant_ownership": ["view", "transfer_emergency"],
  "platform": ["analytics", "monitoring", "audit"],
  "billing": ["view_all", "override"]
}
```

**Responsibilities:**
- ✅ Create new tenants
- ✅ Suspend/activate tenants
- ✅ Delete tenants (soft delete)
- ✅ View all tenant owners
- ✅ Transfer ownership (emergency only)
- ✅ Platform-wide analytics
- ✅ Global monitoring

**Cannot:**
- ❌ Access tenant's business data (unless explicitly granted)
- ❌ Transfer ownership without audit trail
- ❌ Delete tenant with active subscription

**Example:**
```http
POST /api/auth/core/tenants
Authorization: Bearer {platform_admin_token}

{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "owner_id": "usr-alice",  // Alice becomes tenant owner
  "db_dsn": "postgres://localhost/acme_db",
  "db_schema": "acme"
}
```

### 2. Tenant Owner

**Tenant:** Any tenant (except `system`)  
**Count:** **Exactly 1 per tenant** (enforced by DB constraint)  
**Storage:** `db_main.users` where `is_tenant_owner = true`  
**Identifier:** `tenant.owner_id = user.id`

**Permissions:**
```json
{
  "tenant": ["read", "update"],
  "tenant_settings": ["update"],
  "billing": ["read", "update", "cancel"],
  "ownership": ["transfer"],
  "admin": ["assign", "revoke"],
  "app": ["create", "read", "update", "delete"],
  "user": ["create", "read", "update", "delete"]
}
```

**Responsibilities:**
- ✅ **Billing & subscription** (legal owner)
- ✅ Update tenant settings
- ✅ Transfer ownership to another user
- ✅ Assign/revoke tenant admin roles
- ✅ Full access to tenant resources
- ✅ Cancel tenant subscription

**Cannot:**
- ❌ Create other tenants (not platform admin)
- ❌ Access other tenants' data
- ❌ Delete tenant (must contact platform admin)

**Transfer Ownership:**
```http
POST /api/auth/core/tenants/{tenant_id}/transfer-ownership
Authorization: Bearer {owner_token}

{
  "new_owner_id": "usr-bob",  // Bob must be user in this tenant
  "reason": "Company ownership change"
}
```

**Enforcement:**
- ✅ Tenant can have **only 1 owner** at a time
- ✅ Owner must be a user in the tenant
- ✅ Transfer is **audited** in `tenant_ownership_history`
- ✅ New owner gets notification
- ✅ Old owner loses owner privileges (becomes regular user/admin)

### 3. Tenant Admin

**Tenant:** Any tenant (except `system`)  
**Count:** Multiple allowed  
**Storage:** `db_main.users` with role `tenant-admin` in `user_roles`

**Permissions:**
```json
{
  "tenant": ["read"],
  "app": ["create", "read", "update", "delete"],
  "user": ["create", "read", "update", "delete"],
  "role": ["create", "read", "update", "delete", "assign"],
  "permission": ["create", "read", "update", "delete"],
  "policy": ["create", "read", "update", "delete"]
}
```

**Responsibilities:**
- ✅ Manage apps (CRUD)
- ✅ Manage users (CRUD)
- ✅ Create and assign roles
- ✅ Define permissions
- ✅ Configure app settings
- ✅ View tenant information

**Cannot:**
- ❌ Update billing information
- ❌ Transfer ownership
- ❌ Delete tenant
- ❌ Access billing/subscription
- ❌ Assign/revoke tenant admin role (only owner can)

**Assignment:**
```http
# Owner assigns admin role to user
POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles
Authorization: Bearer {owner_token}

{
  "role_id": "role-tenant-admin"
}
```

### 4. Regular User

**Tenant:** Any tenant  
**Count:** Unlimited (within tenant quota)  
**Storage:** `db_main.users` with standard roles

**Permissions:** Defined by assigned roles

**Responsibilities:**
- ✅ Access assigned apps
- ✅ Use app features
- ✅ Update own profile
- ✅ Limited to granted permissions

## 🔐 Permission Matrix

| Action | Platform Admin | Tenant Owner | Tenant Admin | Regular User |
|--------|---------------|--------------|--------------|--------------|
| **Tenant Management** |
| Create tenant | ✅ | ❌ | ❌ | ❌ |
| Update tenant settings | ✅ | ✅ | ❌ | ❌ |
| Delete tenant | ✅ | ❌ | ❌ | ❌ |
| Suspend tenant | ✅ | ❌ | ❌ | ❌ |
| View tenant info | ✅ | ✅ | ✅ | ✅ |
| Transfer ownership | ✅ (emergency) | ✅ | ❌ | ❌ |
| **Billing** |
| View billing | ✅ (all) | ✅ (own) | ❌ | ❌ |
| Update payment method | ❌ | ✅ | ❌ | ❌ |
| Cancel subscription | ❌ | ✅ | ❌ | ❌ |
| **App Management** |
| Create app | ✅ | ✅ | ✅ | ❌ |
| Update app | ✅ | ✅ | ✅ | ❌ |
| Delete app | ✅ | ✅ | ✅ | ❌ |
| **User Management** |
| Create user | ✅ | ✅ | ✅ | ❌ |
| Update user | ✅ | ✅ | ✅ | ✅ (self) |
| Delete user | ✅ | ✅ | ✅ | ❌ |
| Assign tenant admin | ✅ | ✅ | ❌ | ❌ |
| **Authorization** |
| Create role | ✅ | ✅ | ✅ | ❌ |
| Assign role | ✅ | ✅ | ✅ | ❌ |
| Create permission | ✅ | ✅ | ✅ | ❌ |

## 📋 Complete Workflows

### Workflow 1: Platform Admin Creates Tenant

```http
# 1. Platform admin creates tenant
POST /api/auth/core/tenants
Authorization: Bearer {platform_admin_token}

{
  "id": "acme-corp",
  "name": "Acme Corporation",
  "owner_id": "usr-alice",  // Alice must exist first
  "db_dsn": "postgres://localhost/acme_db",
  "db_schema": "acme",
  "settings": {
    "max_users": 100,
    "max_apps": 10
  }
}

# 2. System automatically:
# - Sets tenants.owner_id = usr-alice
# - Sets users.is_tenant_owner = TRUE for Alice
# - Creates default "tenant-admin" role
# - Assigns tenant-owner role to Alice
```

### Workflow 2: Tenant Owner Assigns Admin

```http
# 1. Owner creates user for tenant
POST /api/auth/core/tenants/acme-corp/users
Authorization: Bearer {owner_token}

{
  "username": "bob",
  "email": "bob@acme.com",
  "full_name": "Bob Manager"
}

# 2. Owner assigns tenant-admin role to Bob
POST /api/auth/rbac/tenants/acme-corp/apps/admin-console/users/{bob_id}/roles
Authorization: Bearer {owner_token}

{
  "role_id": "role-tenant-admin"
}
```

### Workflow 3: Transfer Ownership

```http
# 1. Current owner initiates transfer
POST /api/auth/core/tenants/acme-corp/transfer-ownership
Authorization: Bearer {alice_token}

{
  "new_owner_id": "usr-bob",
  "reason": "Company sold to Bob"
}

# 2. System automatically:
# - Validates Bob exists in tenant
# - Validates Alice is current owner
# - Updates tenants.owner_id = usr-bob
# - Sets users.is_tenant_owner = FALSE for Alice
# - Sets users.is_tenant_owner = TRUE for Bob
# - Creates audit record in tenant_ownership_history
# - Sends notification to both Alice and Bob
```

### Workflow 4: Platform Admin Emergency Transfer

```http
# When owner account is compromised/lost
POST /api/auth/core/tenants/acme-corp/transfer-ownership
Authorization: Bearer {platform_admin_token}

{
  "new_owner_id": "usr-charlie",
  "reason": "Emergency: original owner account compromised",
  "force": true
}

# Audit trail preserved in tenant_ownership_history
```

## 🛡️ Security & Constraints

### Database Constraints

```sql
-- Only 1 owner per tenant
ALTER TABLE tenants 
ADD CONSTRAINT unique_owner_per_tenant 
UNIQUE (owner_id);

-- Owner must be user in the same tenant
ALTER TABLE tenants
ADD CONSTRAINT fk_owner_same_tenant
FOREIGN KEY (owner_id, id) 
REFERENCES users(id, tenant_id);

-- Owner flag constraint
ALTER TABLE users
ADD CONSTRAINT check_owner_per_tenant
CHECK (
  -- If is_tenant_owner = TRUE, must match tenant.owner_id
  NOT is_tenant_owner OR 
  EXISTS (SELECT 1 FROM tenants WHERE owner_id = users.id AND id = users.tenant_id)
);
```

### Application-Level Checks

```go
// When creating tenant
func (s *TenantService) CreateTenant(req *CreateTenantRequest) error {
    // 1. Verify owner user exists
    owner, err := s.userStore.GetUser(req.OwnerID)
    if err != nil {
        return fmt.Errorf("owner user not found: %w", err)
    }
    
    // 2. Create tenant with owner
    tenant := &Tenant{
        ID:      req.ID,
        Name:    req.Name,
        OwnerID: req.OwnerID,
        // ...
    }
    
    // 3. Set user as tenant owner
    owner.IsTenantOwner = true
    s.userStore.UpdateUser(owner)
    
    // 4. Auto-assign tenant-owner role
    s.roleService.AssignRole(owner.ID, "tenant-owner")
    
    return nil
}

// When transferring ownership
func (s *TenantService) TransferOwnership(req *TransferOwnershipRequest) error {
    // 1. Authorization check
    if !s.isOwnerOrPlatformAdmin(req.CurrentUser, req.TenantID) {
        return ErrUnauthorized
    }
    
    // 2. Verify new owner exists in tenant
    newOwner, err := s.userStore.GetUser(req.NewOwnerID)
    if err != nil || newOwner.TenantID != req.TenantID {
        return ErrInvalidNewOwner
    }
    
    // 3. Get current owner
    tenant, _ := s.tenantStore.GetTenant(req.TenantID)
    oldOwner, _ := s.userStore.GetUser(tenant.OwnerID)
    
    // 4. Transfer ownership
    tenant.OwnerID = req.NewOwnerID
    oldOwner.IsTenantOwner = false
    newOwner.IsTenantOwner = true
    
    // 5. Create audit record
    s.auditStore.CreateOwnershipHistory(&TenantOwnershipHistory{
        TenantID:      req.TenantID,
        PreviousOwner: oldOwner.ID,
        NewOwner:      newOwner.ID,
        TransferredBy: req.CurrentUser,
        Reason:        req.Reason,
        TransferredAt: time.Now(),
    })
    
    return nil
}
```

## 📊 Benefits of This Architecture

### ✅ Advantages

1. **Clear Ownership**
   - Every tenant has exactly 1 legal owner
   - Billing responsibility is clear
   - Ownership transfer is audited

2. **Centralized Auth**
   - Single source of truth for users
   - Easier cross-tenant features (SSO)
   - Consistent user identity

3. **Data Isolation**
   - Business data separated per tenant
   - Independent scaling
   - Custom schemas possible

4. **Flexible Administration**
   - Multiple admins per tenant
   - Role-based access control
   - Owner can delegate management

5. **Audit Trail**
   - All ownership transfers logged
   - Platform admin actions tracked
   - Compliance-ready

### ⚠️ Considerations

1. **Global DB Performance**
   - Must handle all tenants' users
   - Proper indexing required
   - Consider partitioning at scale

2. **Backup Strategy**
   - Global DB = critical (auth data)
   - Tenant DBs = per-tenant backup
   - Must backup both

3. **Migration Complexity**
   - Existing tenants need owner assignment
   - Historical data migration
   - Ownership history backfill

## 🚀 Implementation Checklist

### Database Schema Updates
- [ ] Add `owner_id` to `tenants` table (NOT NULL, UNIQUE)
- [ ] Add `is_tenant_owner` to `users` table
- [ ] Create `tenant_ownership_history` table
- [ ] Add constraints for owner uniqueness
- [ ] Add indexes on `owner_id`, `is_tenant_owner`

### Application Code
- [ ] Update `Tenant` domain model with `OwnerID`
- [ ] Update `User` domain model with `IsTenantOwner`
- [ ] Implement `TransferOwnership` service method
- [ ] Add ownership validation in tenant creation
- [ ] Create ownership history audit service
- [ ] Add middleware to check owner permissions

### Bootstrap Updates
- [ ] Set platform admin in `system` tenant
- [ ] Ensure bootstrap creates owner for system tenant
- [ ] Update migration to set initial owner

### API Endpoints
- [ ] `POST /tenants/{id}/transfer-ownership`
- [ ] `GET /tenants/{id}/ownership-history`
- [ ] `GET /tenants/{id}/owner`
- [ ] Update tenant creation to require `owner_id`

### Documentation
- [ ] Update API docs with ownership endpoints
- [ ] Create ownership transfer guide
- [ ] Document platform admin vs tenant owner
- [ ] Create migration guide for existing tenants

## 📚 Related Documentation

- [Multi-Tenant Architecture](./multi_tenant_architecture.md)
- [Bootstrap Guide](./BOOTSTRAP.md)
- [RBAC Implementation](./rbac.md)
- [Security Best Practices](./security.md)
