Global App :
- Port: 8080
- TenantService:
    - GET /auth/tenants -> ListTenants
    - POST /auth/tenants -> CreateTenant
    - GET /auth/tenants/{id} -> GetTenant
    - PUT /auth/tenants/{id} -> UpdateTenant
    - DELETE /auth/tenants/{id} -> DeleteTenant
    - POST /auth/tenants/{id}/activate -> ActivateTenant
    - POST /auth/tenants/{id}/suspend -> SuspendTenant

Tenant App:
- Port: 8081
- UserService: