# Go DDD Technical Documentation

Welcome to the technical documentation system for **go-ddd** — a clean Domain-Driven Design REST API implemented in Go.

## Documentation Index

### Core Documentation
- **[Developer Onboarding Guide](/docs?page=onboarding)**
  Environment setup, local running with Docker, Goose migrations, unit testing, and developer workflow.
- **[Architecture Reference](/docs?page=architecture)**
  Domain-Driven Design layer boundaries (Domain, Application, Infrastructure, Presentation) and Go implementation patterns.

### Integration & Feature Guides
- **[External Client Integration Guide](/docs?page=external-client-integration)**
  Complete guide for frontend, mobile, and third-party API clients interacting with registration, login, and JWT protected routes.
- **[Error Builder & Handling Guide](/docs?page=error-builder)**
  Domain error definitions, error mapping to HTTP status codes, and unified JSON error payload structure.

### Domain Modules
- **[User Management Module](/docs?page=user-module)**
  User aggregate root, password hashing, MySQL user repository, JWT service, and user HTTP handlers.

### Documentation Templates
- **[Feature Module Template](/docs?page=feature-template)**
- **[Deep-Dive Guide Template](/docs?page=guide-template)**
- **[Swagger Description Template](/docs?page=swagger-description-template)**

---

## Machine-Readable Manifest
This documentation suite includes an AI-routable machine-readable manifest at `docs/assets/manifest.json` which maps modules to code paths (`primary_paths`, `cross_cutting_paths`, `risk_areas`).
