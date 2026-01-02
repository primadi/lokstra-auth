# Layer 3: Subject Resolution / Identity Context

## Overview

The Subject Resolution layer is responsible for transforming claims into complete and usable identity context.

## Concepts

### Subject
The authenticated entity (usually a user, but can also be a service, device, etc.).

### Identity Context
Complete information about the subject including:
- User profile
- Roles and groups
- Permissions
- Metadata
- Session information

### Enrichment
Process of adding additional information from various sources (database, cache, external services).

## Components

### Subject Resolvers
Components for transforming claims into subject entities.

### Context Builders
Components for building complete identity context.

### Data Enrichers
Components for adding supplementary information to identity context.

## Implementations

### Simple (`/simple`)
Basic subject resolver with minimal data fetching.

### Enriched (`/enriched`)
Subject resolution with data enrichment from external sources (user profiles, preferences, etc.).

### Cached (`/cached`)
Performance-optimized resolver with caching layer to reduce database queries.

## Contract

All implementations must adhere to the contracts defined in `contract.go`:

```go
// See /rbac/contract.go for interface definitions
```

## Usage Examples

See `/examples/subject` folder for implementation examples.

## API Reference

(To be documented)
