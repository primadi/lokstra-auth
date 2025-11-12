# Layer 4: Authorization / Policy Evaluation

## Overview

The Authorization layer is responsible for determining whether a subject has access to specific resources or operations.

## Concepts

### Authorization Models
- **RBAC** (Role-Based Access Control): Access based on roles
- **ABAC** (Attribute-Based Access Control): Access based on attributes
- **PBAC** (Policy-Based Access Control): Access based on complex policies
- **ReBAC** (Relationship-Based Access Control): Access based on relationships

### Policy
Rules that define access conditions.

### Permission
Specific access rights to resources or operations.

### Resource
Protected entities (documents, API endpoints, features, etc.).

## Components

### Policy Evaluators
Components for evaluating policies against context.

### Permission Checkers
Components for checking specific permissions.

### Access Control Managers
Components for managing access control logic.

## Implementations

### RBAC (`/rbac`)
Role-Based Access Control with hierarchical role support.

### ABAC (`/abac`)
Attribute-Based Access Control for fine-grained authorization based on user, resource, and environmental attributes.

### Policy (`/policy`)
Policy-based authorization using declarative policy languages (e.g., Rego, Cedar).

## Contract

All implementations must adhere to the contracts defined in `contract.go`:

```go
// See /04_authz/contract.go for interface definitions
```

## Usage Examples

See `/examples/04_authz` folder for implementation examples.

## API Reference

(To be documented)
