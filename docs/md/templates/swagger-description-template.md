# Swagger API Annotation Template

<!-- metadata
title: Swagger API Annotation Template
category: Templates
status: active
last_updated: YYYY-MM-DD
-->

Standardized template for annotating Go HTTP handlers for Swaggo / OpenAPI documentation generation.

## Related Documentation
- [External Client Integration Guide](/docs?page=external-client-integration)
- [Error Builder Guide](/docs?page=error-builder)

---

## Section Checklist
- [ ] Swaggo Annotation Block
- [ ] Request & Response Payload DTOs
- [ ] HTTP Status Codes Defined

---

## Example Handler Annotations

<!-- covers: internal/presentation/http/*_handler.go -->

```go
// CreateResource handles new resource creation.
// @Summary      Create a new resource
// @Description  Creates a new resource domain entity with given attributes
// @Tags         Resource
// @Accept       json
// @Produce      json
// @Param        request body CreateResourceRequest true "Resource payload"
// @Success      201  {object}  ResourceResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/resources [post]
func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
    // Handler implementation...
}
```
