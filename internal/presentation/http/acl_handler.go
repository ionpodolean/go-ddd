package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	appACL "go-ddd/internal/application/acl"
	domainACL "go-ddd/internal/domain/acl"
)

type ACLHandler struct{ service *appACL.Service }

func NewACLHandler(service *appACL.Service) *ACLHandler { return &ACLHandler{service: service} }

func (h *ACLHandler) RequireAuthAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int64)
		if !ok {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		isAdmin, err := h.service.IsAuthAdmin(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !isAdmin {
			respondError(w, http.StatusForbidden, domainACL.ErrForbidden.Error())
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ListDomains godoc
// @Summary List domains
// @Tags ACL Domains
// @Produce json
// @Security BearerAuth
// @Success 200 {array} acl.Domain
// @Router /api/v1/acl/domains [get]
func (h *ACLHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.service.ListDomains(r.Context())
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, domains)
}

// CreateDomain godoc
// @Summary Create a domain and its administrator
// @Description Creates the domain, its Admin role, an administrator user, membership, and role assignment in one transaction.
// @Tags ACL Domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body acl.CreateDomainRequest true "Domain and initial administrator"
// @Success 201 {object} acl.Domain
// @Router /api/v1/acl/domains [post]
func (h *ACLHandler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req appACL.CreateDomainRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	domain, err := h.service.CreateDomain(r.Context(), req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, domain)
}

// GetDomain godoc
// @Summary Get a domain
// @Tags ACL Domains
// @Produce json
// @Security BearerAuth
// @Param id path int true "Domain ID"
// @Success 200 {object} acl.Domain
// @Router /api/v1/acl/domains/{id} [get]
func (h *ACLHandler) GetDomain(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	domain, err := h.service.GetDomain(r.Context(), id)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, domain)
}

// UpdateDomain godoc
// @Summary Update a domain
// @Tags ACL Domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Domain ID"
// @Param request body acl.UpdateDomainRequest true "Domain"
// @Success 200 {object} acl.Domain
// @Router /api/v1/acl/domains/{id} [put]
func (h *ACLHandler) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	var req appACL.UpdateDomainRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	domain, err := h.service.UpdateDomain(r.Context(), id, req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, domain)
}

// DeleteDomain godoc
// @Summary Delete a domain
// @Tags ACL Domains
// @Security BearerAuth
// @Param id path int true "Domain ID"
// @Success 204
// @Router /api/v1/acl/domains/{id} [delete]
func (h *ACLHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteDomain(r.Context(), id); err != nil {
		respondACLServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRoles godoc
// @Summary List roles in a domain
// @Tags ACL Roles
// @Produce json
// @Security BearerAuth
// @Param domain_id query int true "Domain ID"
// @Success 200 {array} acl.Role
// @Router /api/v1/acl/roles [get]
func (h *ACLHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	domainID, err := strconv.ParseInt(r.URL.Query().Get("domain_id"), 10, 64)
	if err != nil || domainID < 1 {
		respondError(w, http.StatusBadRequest, "valid domain_id is required")
		return
	}
	roles, err := h.service.ListRoles(r.Context(), domainID)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, roles)
}

// CreateRole godoc
// @Summary Create a domain-scoped role
// @Tags ACL Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body acl.CreateRoleRequest true "Role"
// @Success 201 {object} acl.Role
// @Router /api/v1/acl/roles [post]
func (h *ACLHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req appACL.CreateRoleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	role, err := h.service.CreateRole(r.Context(), req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, role)
}

// GetRole godoc
// @Summary Get a role
// @Tags ACL Roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Success 200 {object} acl.Role
// @Router /api/v1/acl/roles/{id} [get]
func (h *ACLHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	role, err := h.service.GetRole(r.Context(), id)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, role)
}

// UpdateRole godoc
// @Summary Update a role
// @Tags ACL Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param request body acl.UpdateRoleRequest true "Role"
// @Success 200 {object} acl.Role
// @Router /api/v1/acl/roles/{id} [put]
func (h *ACLHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	var req appACL.UpdateRoleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	role, err := h.service.UpdateRole(r.Context(), id, req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Delete a role
// @Tags ACL Roles
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Success 204
// @Router /api/v1/acl/roles/{id} [delete]
func (h *ACLHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteRole(r.Context(), id); err != nil {
		respondACLServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListPermissions godoc
// @Summary List permissions
// @Tags ACL Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {array} acl.Permission
// @Router /api/v1/acl/permissions [get]
func (h *ACLHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.ListPermissions(r.Context())
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, permissions)
}

// CreatePermission godoc
// @Summary Create a permission
// @Tags ACL Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body acl.CreatePermissionRequest true "Permission"
// @Success 201 {object} acl.Permission
// @Router /api/v1/acl/permissions [post]
func (h *ACLHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req appACL.CreatePermissionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	permission, err := h.service.CreatePermission(r.Context(), req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, permission)
}

// GetPermission godoc
// @Summary Get a permission
// @Tags ACL Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} acl.Permission
// @Router /api/v1/acl/permissions/{id} [get]
func (h *ACLHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	permission, err := h.service.GetPermission(r.Context(), id)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, permission)
}

// UpdatePermission godoc
// @Summary Update a permission
// @Tags ACL Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Param request body acl.UpdatePermissionRequest true "Permission"
// @Success 200 {object} acl.Permission
// @Router /api/v1/acl/permissions/{id} [put]
func (h *ACLHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	var req appACL.UpdatePermissionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	permission, err := h.service.UpdatePermission(r.Context(), id, req)
	if err != nil {
		respondACLServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, permission)
}

// DeletePermission godoc
// @Summary Delete a permission
// @Tags ACL Permissions
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 204
// @Router /api/v1/acl/permissions/{id} [delete]
func (h *ACLHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeletePermission(r.Context(), id); err != nil {
		respondACLServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func requestID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(w, http.StatusBadRequest, "valid id is required")
		return 0, false
	}
	return id, true
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
func respondACLServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domainACL.ErrDomainNotFound), errors.Is(err, domainACL.ErrRoleNotFound), errors.Is(err, domainACL.ErrPermissionNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domainACL.ErrInvalidDomain), errors.Is(err, domainACL.ErrInvalidRole), errors.Is(err, domainACL.ErrInvalidPermission):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
