# go-ddd

An API service written in **Go**, built on **Domain-Driven Design (DDD)** and Clean Architecture principles. It provides JWT authentication, user management, and an ACL module for domains, roles, and permissions.

## Technologies

- Go 1.27.3+
- MySQL 8
- REST HTTP and gRPC
- JWT authentication
- Goose for database migrations and SQLC for type-safe data access
- OpenTelemetry, Prometheus, Grafana, Loki, and Tempo for observability

## Architecture

The code is organized in layers, with dependencies pointing toward the domain:

```text
presentation  →  application  →  domain
      ↓
infrastructure (implements domain interfaces)
```

- `internal/domain` — entities, business rules, and repository interfaces.
- `internal/application` — use cases, application services, and DTOs.
- `internal/infrastructure` — MySQL, JWT, email, and concrete interface implementations.
- `internal/presentation` — HTTP/gRPC endpoints, middleware, and JSON serialization.
- `migrations` — database migrations.
- `cmd/main.go` — application entry point.

## Quick start with Docker

You need Docker and Docker Compose. Copy `.env.example` to `.env`, then set
`JWT_SECRET` to a random secret of at least 32 bytes and set
`AUTH_ADMIN_EMAIL` and `AUTH_ADMIN_PASSWORD` (at least 12 bytes). From the
project directory, run:

```bash
docker compose up -d
```

On startup, the application builds the server, automatically runs migrations, and starts all required services.

| Service | Address |
| --- | --- |
| API HTTP | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Health check | http://localhost:8080/health |
| Prometheus metrics | http://localhost:8080/metrics |
| gRPC | `localhost:9090` |
| Prometheus | http://localhost:9091 |
| Adminer | http://localhost:8081 |
| Grafana | http://localhost:3000 (`admin` / `admin`) |

To follow the logs:

```bash
docker compose logs -f app
```

To stop the services:

```bash
docker compose down
```

## Local setup

1. Install Go 1.27.3+ and start a MySQL 8 instance.
2. Copy `.env.example` to `.env` and set the required security configuration. For example:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=user
DB_PASSWORD=password
DB_NAME=data
JWT_SECRET=<random-secret-at-least-32-bytes>
AUTH_ADMIN_EMAIL=<administrator-email>
AUTH_ADMIN_PASSWORD=<administrator-password-at-least-12-bytes>
GRPC_PORT=9090
```

3. Download dependencies and start the application:

```bash
go mod download
go run ./cmd/main.go
```

The application applies Goose migrations at startup and stops if they fail. The
bootstrap credentials are required at startup; they create the initial Auth
administrator only when the Auth domain does not exist. If that email already
belongs to a user, bootstrap fails rather than granting that account access.
For common commands, you can also use [Task](https://taskfile.dev/): `task run`, `task test`, `task docker:up`.

The Compose setup is for local development and publishes ports only on loopback.
When run directly, the application listens on `127.0.0.1` by default; set
`BIND_ADDRESS` explicitly if it must listen on another interface.
The application itself serves HTTP and gRPC without TLS; use TLS termination and
access controls before exposing it to a network. Login and registration share
an in-process limit of 20 requests per minute per direct peer and 300 per minute
overall. Forwarded IP headers are ignored. Behind a proxy, all clients may share
one limit; configure a trusted edge rate limiter and TLS, especially when using
multiple replicas, since this in-process limit is not shared across instances.

The administrator of the `auth` domain is the super admin: they can manage all
domains, roles, and the shared permission catalogue. Each new domain is created
with its own administrator. A domain administrator can view and update their
domain, manage its ordinary roles, and read the shared permission catalogue;
they cannot access another domain or manage global settings. The built-in
administrator roles and the `auth` domain cannot be changed or deleted through
the API. Stored permissions can be assigned to roles, but are not yet evaluated
as access policies for other application operations.

## Useful endpoints

- `POST /api/v1/auth/register` — creates an account and returns a JWT token.
- `POST /api/v1/auth/login` — authenticates a user.
- `GET /api/v1/users/me` — returns the authenticated user's profile.
- `/api/v1/acl/*` — manages domains, roles, and permissions; requires an administrator account.

For the complete API contract, open Swagger UI after starting the application.

## Testing and development

Pull requests targeting `test`, `stage`, or `main` run the `Go tests` check.
The repository ruleset requires a pull request and a successful `Go tests`
status check for these branches. The check also runs for merge queue entries.

```bash
go test -v -race ./...
```

Other available commands:

```bash
task lint        # checks style and static-analysis issues
task sqlc        # regenerates SQLC code
task proto:generate  # regenerates Go code from Protobuf contracts
task swagger     # regenerates OpenAPI documentation
task docgen      # regenerates internal documentation
```

Internal documentation is available at `http://localhost:8080/docs` after the application starts.
