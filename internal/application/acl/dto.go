package acl

type AdminUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type CreateDomainRequest struct {
	Name  string           `json:"name"`
	Slug  string           `json:"slug"`
	Admin AdminUserRequest `json:"admin"`
}

type UpdateDomainRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateRoleRequest struct {
	DomainID      int64   `json:"domain_id"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type UpdateRoleRequest = CreateRoleRequest

type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type UpdatePermissionRequest = CreatePermissionRequest
