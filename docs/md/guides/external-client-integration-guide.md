# External Client Integration Guide

<!-- metadata
title: External Client Integration Guide
category: Guides
status: active
last_updated: 2026-08-15
-->

This guide provides integration instructions for frontend applications (React, Next.js, Vue), mobile apps (iOS, Android, Flutter), and external API clients consuming the **go-ddd** API.

## Related Documentation
- [User Management Module](/docs?page=user-module)
- [Error Builder & Handling Guide](/docs?page=error-builder)
- [Architecture Reference](/docs?page=architecture)

---

## Authentication Flow Overview

<!-- covers: internal/presentation/http/user_handler.go, internal/infrastructure/security/jwt.go -->

The API uses **JWT (JSON Web Tokens)** for stateless authentication.

1. **Register**: Send user credentials (`name`, `email`, `password`) to POST `/api/v1/auth/register`.
2. **Login**: Send email and password to POST `/api/v1/auth/login` to receive a Bearer JWT token.
3. **Authenticated Requests**: Include the header `Authorization: Bearer <token>` on protected routes such as GET `/api/v1/users/me`.

---

## API Endpoint Reference

### 1. Register User
**Endpoint**: `POST /api/v1/auth/register`  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "SecretPassword123!"
}
```

**Response (201 Created)**:
```json
{
  "id": 1,
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2026-08-15T12:00:00Z"
}
```

---

### 2. User Login
**Endpoint**: `POST /api/v1/auth/login`  
**Content-Type**: `application/json`

**Request Body**:
```json
{
  "email": "jane@example.com",
  "password": "SecretPassword123!"
}
```

**Response (200 OK)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

---

### 3. Get User Profile (Protected)
**Endpoint**: `GET /api/v1/users/me`  
**Headers**: `Authorization: Bearer <JWT_TOKEN>`

**Response (200 OK)**:
```json
{
  "id": 1,
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2026-08-15T12:00:00Z"
}
```

**Response (401 Unauthorized)**:
```json
{
  "error": "Unauthorized - Invalid or missing token"
}
```
