# go-ddd

An API service written in **Go**, built on **Domain-Driven Design (DDD)** and Clean Architecture principles. It provides JWT authentication, user management, and an ACL module for domains, roles, and permissions.

## Technologies

- Go 1.25+
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

You need Docker and Docker Compose. From the project directory, run:

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

1. Install Go 1.25+ and start a MySQL 8 instance.
2. Create an optional `.env` file with the required configuration:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=user
DB_PASSWORD=password
DB_NAME=data
JWT_SECRET=replace-this-secret
GRPC_PORT=9090
```

3. Download dependencies and start the application:

```bash
go mod download
go run ./cmd/main.go
```

The application applies Goose migrations at startup. For common commands, you can also use [Task](https://taskfile.dev/): `task run`, `task test`, `task docker:up`.

## Useful endpoints

- `POST /api/v1/auth/register` — creates an account and returns a JWT token.
- `POST /api/v1/auth/login` — authenticates a user.
- `GET /api/v1/users/me` — returns the authenticated user's profile.
- `/api/v1/acl/*` — manages domains, roles, and permissions; requires an administrator account.

For the complete API contract, open Swagger UI after starting the application.

## Testing and development

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
