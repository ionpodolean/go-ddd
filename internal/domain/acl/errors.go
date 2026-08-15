package acl

import "errors"

var (
	ErrDomainNotFound     = errors.New("domain not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrInvalidDomain      = errors.New("domain name and slug are required")
	ErrInvalidRole        = errors.New("role domain, name, and code are required")
	ErrInvalidPermission  = errors.New("permission name and code are required")
	ErrForbidden          = errors.New("forbidden")
)
