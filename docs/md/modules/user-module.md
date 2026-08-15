# User Management Module

<!-- metadata
title: User Management Module
category: Modules
status: active
last_updated: 2026-08-15
-->

The User Management module governs user registration, password security, authentication credentials, profile retrieval, and JWT issuance.

## Related Documentation
- [Architecture Reference](/docs?page=architecture)
- [External Client Integration Guide](/docs?page=external-client-integration)
- [Error Builder Guide](/docs?page=error-builder)

---

## Module Overview

<!-- covers: internal/domain/user/**, internal/application/user/**, internal/infrastructure/persistence/mysql_user_repository.go -->

| Aspect | Details |
| --- | --- |
| **Domain Package** | `internal/domain/user` |
| **Application Package** | `internal/application/user` |
| **Persistence** | `internal/infrastructure/persistence/mysql_user_repository.go` |
| **Security** | `internal/infrastructure/security/jwt.go` |
| **HTTP Handlers** | `internal/presentation/http/user_handler.go` |

---

## Primary Code Paths

- **Domain Entity & Repository Interface**:
  - `internal/domain/user/user.go`
  - `internal/domain/user/repository.go`
- **Application Service**:
  - `internal/application/user/service.go`
- **Infrastructure Implementations**:
  - `internal/infrastructure/persistence/mysql_user_repository.go`
  - `internal/infrastructure/security/jwt.go`
- **HTTP Presentation**:
  - `internal/presentation/http/user_handler.go`

---

## Security & Risk Areas

- **Password Hashing**: Uses `golang.org/x/crypto/bcrypt` for secure salt & hash generation.
- **JWT Key**: `JWT_SECRET` must be set via environment variables in production environments.
- **Database Indexing**: Unique index on `users.email` column in MySQL to enforce email uniqueness at database level.
