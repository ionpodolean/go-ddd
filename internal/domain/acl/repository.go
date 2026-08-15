package acl

import (
	"context"

	domainUser "go-ddd/internal/domain/user"
)

type Repository interface {
	SeedAuth(ctx context.Context, admin *domainUser.User) error
	CreateDomainWithAdmin(ctx context.Context, domain *Domain, admin *domainUser.User) error
	ListDomains(ctx context.Context) ([]Domain, error)
	GetDomain(ctx context.Context, id int64) (*Domain, error)
	UpdateDomain(ctx context.Context, domain *Domain) error
	DeleteDomain(ctx context.Context, id int64) error
	CreateRole(ctx context.Context, role *Role) error
	ListRoles(ctx context.Context, domainID int64) ([]Role, error)
	GetRole(ctx context.Context, id int64) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id int64) error
	CreatePermission(ctx context.Context, permission *Permission) error
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermission(ctx context.Context, id int64) (*Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) error
	DeletePermission(ctx context.Context, id int64) error
	IsAuthAdmin(ctx context.Context, userID int64) (bool, error)
}
