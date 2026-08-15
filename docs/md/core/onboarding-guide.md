# Developer Onboarding Guide

<!-- metadata
title: Developer Onboarding Guide
category: Core
status: active
last_updated: 2026-08-15
-->

Welcome to the **go-ddd** project! This guide will walk you through setting up your local environment, running the application with Docker or Go directly, running database migrations, and running tests.

## Related Documentation
- [Architecture Reference](/docs?page=architecture)
- [User Management Module](/docs?page=user-module)
- [Error Builder Guide](/docs?page=error-builder)

---

## Prerequisites

Before starting, ensure you have the following installed on your machine:
- **Go**: v1.25+ (`go version`)
- **Docker & Docker Compose** (for MySQL database container)
- **Goose**: CLI for database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- **Task**: (Optional) Task runner (`brew install go-task/tap/go-task`)

---

## Quick Start Setup

<!-- covers: docker-compose.yml, Makefile, Taskfile.yml, example.env -->

### 1. Environment Configuration
Copy the sample environment file `.env`:
```bash
cp example.env .env
```

Ensure `.env` contains:
```env
PORT=8080
JWT_SECRET=super-secret-jwt-key
DB_DSN=user:password@tcp(localhost:3306)/data?charset=utf8&parseTime=True&loc=Local
```

### 2. Start MySQL Container
Run Docker Compose to spin up MySQL database:
```bash
docker compose up -d
```

### 3. Run Database Migrations
Run database migrations using Goose or Task commands:
```bash
task migrate:up
```
Or directly with Goose:
```bash
goose -dir ./migrations mysql "user:password@tcp(localhost:3306)/data?parseTime=true" up
```

### 4. Run the API Server
Start the HTTP server:
```bash
go run ./cmd/main.go
```
The server will start listening on port `:8080`.

---

## Running Tests

<!-- covers: internal/**/*_test.go -->

Run all unit and integration tests:
```bash
go test -v ./...
```

Run race detector tests:
```bash
task test
```

---

## Useful Development Tasks

<!-- covers: Taskfile.yml -->

| Command | Action |
| --- | --- |
| `task run` | Runs `cmd/main.go` directly |
| `task test` | Runs tests with race detector |
| `task lint` | Runs `golangci-lint` |
| `task swagger` | Regenerates Swagger docs (`docs/docs.go`) |
| `task migrate:up` | Applies database migrations |
| `task docgen` | Regenerates static HTML documentation |
