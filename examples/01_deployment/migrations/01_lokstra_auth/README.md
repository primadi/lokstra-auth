# Lokstra Auth Database Migrations

## Migration Files Overview

### 000_auth_core.sql
**Main database schema containing all core tables and authorization structures**

Tables included:
- **Core Tables:**
  - `tenants` - Multi-tenant organizations (with owner_id)
  - `apps` - Applications within tenants
  - `branches` - Branch/location within apps
  - `users` - User accounts (with is_tenant_owner and lockout fields)
  - `user_identities` - Linked auth providers (OAuth2, SAML, etc.)
  - `user_apps` - User access control to apps
  - `credential_providers` - OAuth2/SAML/Email provider configs
  - `app_keys` - API keys for application authentication

- **RBAC Tables:**
  - `roles` - Role definitions (tenant+app scoped)
  - `user_roles` - User-to-role assignments
  - `permissions` - Permission definitions
  - `permission_compositions` - Compound permission definitions (UI permissions)
  - `role_permissions` - Role-to-permission assignments
  - `user_permissions` - Direct user-to-permission assignments

- **Policy-Based Access Control:**
  - `policies` - Authorization policies (PBAC/ABAC)

- **Tenant Ownership:**
  - `tenant_ownership_history` - Audit trail of ownership transfers

Functions:
- `transfer_tenant_ownership()` - Transfer tenant ownership
- `is_tenant_owner()` - Check if user is tenant owner
- `update_updated_at_column()` - Auto-update timestamps

### 001_bootstrap_platform_admin.sql
**Creates initial platform admin tenant, app, user, and permissions**

Creates:
- Platform tenant (`platform`)
- Platform admin app (`platform-admin-app`)
- Platform admin user (`platform-admin-user`)
- Platform admin role and permissions
- Default credentials for first login

**⚠️ IMPORTANT:** Change password immediately after first deployment!

### 001_create_sync_config.sql
**Configuration sync table with real-time NOTIFY support**

Creates:
- `sync_config` table - Key-value config store (JSONB)
- Auto-notify trigger on config changes
- PostgreSQL LISTEN/NOTIFY support

## Migration Order

Files are executed in alphabetical order:
1. `000_auth_core.sql` - Core schema
2. `001_bootstrap_platform_admin.sql` - Platform admin setup
3. `001_create_sync_config.sql` - Config sync

## Key Features

### Compound Permissions
The `permission_compositions` table supports UI-level permissions that automatically grant multiple underlying permissions:

```sql
-- Example: "ui:user_form" grants both "users:read" and "users:create"
INSERT INTO permissions (id, tenant_id, app_id, name)
VALUES ('ui_user_form', 'demo', 'app1', 'ui:user_form');

INSERT INTO permission_compositions (parent_permission_id, child_permission_id, tenant_id, app_id)
VALUES 
  ('ui_user_form', 'users_read', 'demo', 'app1'),
  ('ui_user_form', 'users_create', 'demo', 'app1');
```

### Tenant Ownership
Each tenant has exactly one owner (user with billing/legal responsibility):
- `tenants.owner_id` - References the owner user
- `users.is_tenant_owner` - Flag for quick owner identification
- `tenant_ownership_history` - Audit trail of ownership transfers

### Account Lockout
Users table includes built-in lockout protection:
- `failed_login_attempts` - Counter for failed logins
- `locked_until` - Auto-unlock timestamp
- `lockout_count` - Total times locked

## Notes

- All tables use composite keys where appropriate (tenant_id + id)
- Foreign keys ensure referential integrity
- Indexes optimized for common queries
- Soft delete support via `deleted_at` fields
- Automatic timestamp management via triggers
