# Error Builder & Handling Guide

<!-- metadata
title: Error Builder & Handling Guide
category: Guides
status: active
last_updated: 2026-08-15
-->

This document describes error handling conventions across domain, application, infrastructure, and presentation layers in **go-ddd**.

## Related Documentation
- [Architecture Reference](/docs?page=architecture)
- [External Client Integration Guide](/docs?page=external-client-integration)

---

## Domain Error Definitions

<!-- covers: internal/domain/user/repository.go -->

Domain errors are sentinel errors defined within domain packages:

```go
package user

import "errors"

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid email or password")
)
```

---

## HTTP Error Response Schema

<!-- covers: internal/presentation/http/user_handler.go, internal/presentation/http/middleware.go -->

HTTP handlers map domain and infrastructure errors to standard JSON response objects:

```json
{
  "error": "Human-readable error summary message",
  "code": "OPTIONAL_ERROR_CODE",
  "details": null
}
```

### HTTP Status Code Mapping

| Domain / Scenario | HTTP Status Code | Response Body Message |
| --- | --- | --- |
| Duplicate email registration | `409 Conflict` | `Email already exists` |
| Invalid credentials on login | `401 Unauthorized` | `Invalid email or password` |
| Missing or invalid JWT header | `401 Unauthorized` | `Unauthorized` |
| User resource not found | `404 Not Found` | `User not found` |
| Internal DB query failure | `500 Internal Server Error` | `Internal server error` |

---

## Middleware Error Recovery

The `RecoveryMiddleware` catches any uncaught runtime panics in HTTP request handlers, logs the stack trace, and returns a `500 Internal Server Error` response.
