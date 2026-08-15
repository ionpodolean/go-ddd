package acl

import (
	"strings"
	"time"
)

type Domain struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID            int64   `json:"id"`
	DomainID      int64   `json:"domain_id"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Description   string  `json:"description,omitempty"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type Permission struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

func NewDomain(name, slug string) (*Domain, error) {
	name = strings.TrimSpace(name)
	slug = normalizeCode(slug)
	if name == "" || slug == "" {
		return nil, ErrInvalidDomain
	}
	now := time.Now()
	return &Domain{Name: name, Slug: slug, CreatedAt: now, UpdatedAt: now}, nil
}

func NewRole(domainID int64, name, code, description string, permissionIDs []int64) (*Role, error) {
	name = strings.TrimSpace(name)
	code = normalizeCode(code)
	if domainID < 1 || name == "" || code == "" {
		return nil, ErrInvalidRole
	}
	return &Role{DomainID: domainID, Name: name, Code: code, Description: strings.TrimSpace(description), PermissionIDs: permissionIDs}, nil
}

func NewPermission(name, code, description string) (*Permission, error) {
	name = strings.TrimSpace(name)
	code = normalizeCode(code)
	if name == "" || code == "" {
		return nil, ErrInvalidPermission
	}
	return &Permission{Name: name, Code: code, Description: strings.TrimSpace(description)}, nil
}

func normalizeCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	return value
}
