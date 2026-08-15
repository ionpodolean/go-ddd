# Feature Module Title

<!-- metadata
title: Feature Module Title
category: Modules
status: draft | active | deprecated
last_updated: YYYY-MM-DD
owner: Team Name
-->

Short 1-2 sentence description of the module and its primary responsibility in the application.

## Related Documentation
- [Architecture Reference](/docs?page=architecture)
- [Developer Onboarding Guide](/docs?page=onboarding)

---

## Section Checklist
- [ ] Module Overview & Key Domain Entities
- [ ] Primary Code Paths documented with `<!-- covers: ... -->` markers
- [ ] Cross-Cutting Concerns & Security Audit
- [ ] Data Persistence Schema & Migrations

---

## Module Overview

<!-- covers: internal/domain/{feature}/** -->

Detailed explanation of the domain model, aggregate roots, value objects, and business rules.

### Key Domain Structs
```go
type FeatureEntity struct {
    ID        int64
    Name      string
    CreatedAt time.Time
}
```

---

## Primary Code Paths

<!-- covers: internal/application/{feature}/**, internal/infrastructure/persistence/{feature}_repository.go -->

- **Domain**: `internal/domain/{feature}/`
- **Application**: `internal/application/{feature}/`
- **Infrastructure**: `internal/infrastructure/persistence/{feature}_repository.go`
- **Presentation**: `internal/presentation/http/{feature}_handler.go`

---

## Risk Areas & Security Audit

<!-- covers: internal/infrastructure/security/** -->

Document risk areas, performance bottlenecks, caching strategies, and security considerations.
