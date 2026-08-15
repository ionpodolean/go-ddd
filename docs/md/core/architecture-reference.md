# Architecture Reference

<!-- metadata
title: Architecture Reference
category: Core
status: active
last_updated: 2026-08-15
-->

This document outlines the Domain-Driven Design (DDD) architectural layers, code structure, and design principles implemented in **go-ddd**.

## Related Documentation
- [Developer Onboarding Guide](/docs?page=onboarding)
- [User Management Module](/docs?page=user-module)

---

## High-Level Architectural Layers

<!-- covers: internal/domain/**, internal/application/**, internal/infrastructure/**, internal/presentation/** -->

The application adheres strictly to the **Dependency Rule**: inner layers (Domain) do not depend on outer layers (Infrastructure, Presentation). Dependencies point inward.

```
+-------------------------------------------------------------------+
|                       PRESENTATION LAYER                          |
|  (HTTP Handlers, Router, Middleware, JSON Serialization)          |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
|                       APPLICATION LAYER                           |
|  (Use Cases, Application Services, DTOs, Orchestration)           |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
|                         DOMAIN LAYER                              |
|  (Entities, Aggregates, Value Objects, Domain Repository Interfaces)|
+-------------------------------------------------------------------+
                                  ^
                                  | (implements)
+-------------------------------------------------------------------+
|                     INFRASTRUCTURE LAYER                          |
|  (MySQL Persistence, Goose Migrations, JWT Security, SQLC)        |
+-------------------------------------------------------------------+
```

---

## Layer Breakdown

### 1. Domain Layer (`internal/domain/`)
- Contains domain entities (e.g. `User`), domain logic, validation, and repository interface definitions.
- Zero external dependencies. Pure Go logic.

### 2. Application Layer (`internal/application/`)
- Encapsulates use cases (e.g. `UserService`).
- Coordinates domain entities and infrastructure abstractions (repositories, token services).
- Defines request/response DTOs.

### 3. Infrastructure Layer (`internal/infrastructure/`)
- Provides concrete implementations for domain interfaces.
- `infrastructure/persistence`: MySQL user repository (`MySQLUserRepository`).
- `infrastructure/security`: JWT token generation & verification (`JWTService`).

### 4. Presentation Layer (`internal/presentation/http/`)
- HTTP endpoints, router configuration, middleware (logging, recovery, auth).
- Deserializes JSON requests, invokes Application Services, and formats HTTP responses.

---

## Directory Layout Summary

```
go-ddd/
├── cmd/
│   ├── main.go               # Server entry point
│   └── docgen/               # Static documentation compiler tool
├── docs/                     # Documentation assets, MD sources, HTML outputs
│   ├── assets/               # docs.css, manifest.json
│   ├── md/                   # Source Markdown files
│   └── html/                 # Generated HTML artifacts
├── internal/
│   ├── domain/               # Domain Entities & Interfaces
│   ├── application/          # Use Cases & Application Services
│   ├── infrastructure/       # MySQL Repositories & JWT Security
│   └── presentation/         # HTTP Handlers, Routers, Middleware
├── migrations/               # Goose SQL migrations
└── pkg/                      # Shared DB initialization utilities
```
