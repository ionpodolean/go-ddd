package acl

import (
	"context"

	domainACL "go-ddd/internal/domain/acl"
	domainUser "go-ddd/internal/domain/user"
)

type Service struct {
	repo domainACL.Repository
}

func NewService(repo domainACL.Repository) *Service { return &Service{repo: repo} }

func (s *Service) SeedAuth(ctx context.Context, admin AdminUserRequest) error {
	user, err := domainUser.NewUser(admin.Email, admin.Password, admin.FirstName, admin.LastName)
	if err != nil {
		return err
	}
	return s.repo.SeedAuth(ctx, user)
}

func (s *Service) CreateDomain(ctx context.Context, req CreateDomainRequest) (*domainACL.Domain, error) {
	domain, err := domainACL.NewDomain(req.Name, req.Slug)
	if err != nil {
		return nil, err
	}
	admin, err := domainUser.NewUser(req.Admin.Email, req.Admin.Password, req.Admin.FirstName, req.Admin.LastName)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateDomainWithAdmin(ctx, domain, admin); err != nil {
		return nil, err
	}
	return domain, nil
}

func (s *Service) ListDomains(ctx context.Context) ([]domainACL.Domain, error) {
	return s.repo.ListDomains(ctx)
}
func (s *Service) GetDomain(ctx context.Context, id int64) (*domainACL.Domain, error) {
	return s.repo.GetDomain(ctx, id)
}

func (s *Service) UpdateDomain(ctx context.Context, id int64, req UpdateDomainRequest) (*domainACL.Domain, error) {
	domain, err := domainACL.NewDomain(req.Name, req.Slug)
	if err != nil {
		return nil, err
	}
	domain.ID = id
	if err := s.repo.UpdateDomain(ctx, domain); err != nil {
		return nil, err
	}
	return s.repo.GetDomain(ctx, id)
}

func (s *Service) DeleteDomain(ctx context.Context, id int64) error {
	return s.repo.DeleteDomain(ctx, id)
}

func (s *Service) CreateRole(ctx context.Context, req CreateRoleRequest) (*domainACL.Role, error) {
	role, err := domainACL.NewRole(req.DomainID, req.Name, req.Code, req.Description, req.PermissionIDs)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) ListRoles(ctx context.Context, domainID int64) ([]domainACL.Role, error) {
	return s.repo.ListRoles(ctx, domainID)
}
func (s *Service) GetRole(ctx context.Context, id int64) (*domainACL.Role, error) {
	return s.repo.GetRole(ctx, id)
}

func (s *Service) UpdateRole(ctx context.Context, id int64, req UpdateRoleRequest) (*domainACL.Role, error) {
	current, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.DomainID == 0 {
		req.DomainID = current.DomainID
	}
	role, err := domainACL.NewRole(req.DomainID, req.Name, req.Code, req.Description, req.PermissionIDs)
	if err != nil {
		return nil, err
	}
	role.ID = id
	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}
	return s.repo.GetRole(ctx, id)
}

func (s *Service) DeleteRole(ctx context.Context, id int64) error { return s.repo.DeleteRole(ctx, id) }

func (s *Service) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*domainACL.Permission, error) {
	permission, err := domainACL.NewPermission(req.Name, req.Code, req.Description)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}

func (s *Service) ListPermissions(ctx context.Context) ([]domainACL.Permission, error) {
	return s.repo.ListPermissions(ctx)
}
func (s *Service) GetPermission(ctx context.Context, id int64) (*domainACL.Permission, error) {
	return s.repo.GetPermission(ctx, id)
}

func (s *Service) UpdatePermission(ctx context.Context, id int64, req UpdatePermissionRequest) (*domainACL.Permission, error) {
	permission, err := domainACL.NewPermission(req.Name, req.Code, req.Description)
	if err != nil {
		return nil, err
	}
	permission.ID = id
	if err := s.repo.UpdatePermission(ctx, permission); err != nil {
		return nil, err
	}
	return s.repo.GetPermission(ctx, id)
}

func (s *Service) DeletePermission(ctx context.Context, id int64) error {
	return s.repo.DeletePermission(ctx, id)
}
func (s *Service) IsAuthAdmin(ctx context.Context, userID int64) (bool, error) {
	return s.repo.IsAuthAdmin(ctx, userID)
}
