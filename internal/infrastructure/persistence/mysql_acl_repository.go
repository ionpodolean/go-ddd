package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainACL "go-ddd/internal/domain/acl"
	domainUser "go-ddd/internal/domain/user"
)

const authDomainSlug = "auth"

type MySQLACLRepository struct{ db *sql.DB }

func NewMySQLACLRepository(db *sql.DB) *MySQLACLRepository { return &MySQLACLRepository{db: db} }

func (r *MySQLACLRepository) SeedAuth(ctx context.Context, admin *domainUser.User) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM domains WHERE slug = ?`, authDomainSlug).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}

		result, err := tx.ExecContext(ctx, `INSERT INTO domains (name, slug) VALUES (?, ?)`, "Auth", authDomainSlug)
		if err != nil {
			return err
		}
		domainID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		return r.createAdministrator(ctx, tx, domainID, "Auth Admin", admin, true)
	})
}

func (r *MySQLACLRepository) CreateDomainWithAdmin(ctx context.Context, domain *domainACL.Domain, admin *domainUser.User) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `INSERT INTO domains (name, slug, created_at, updated_at) VALUES (?, ?, ?, ?)`, domain.Name, domain.Slug, domain.CreatedAt, domain.UpdatedAt)
		if err != nil {
			return err
		}
		domainID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		domain.ID = domainID
		return r.createAdministrator(ctx, tx, domainID, "Admin", admin, false)
	})
}

func (r *MySQLACLRepository) createAdministrator(ctx context.Context, tx *sql.Tx, domainID int64, roleName string, admin *domainUser.User, allowExistingUser bool) error {
	roleResult, err := tx.ExecContext(ctx, `INSERT INTO roles (domain_id, name, code, description) VALUES (?, ?, 'admin', 'Full access within this domain')`, domainID, roleName)
	if err != nil {
		return err
	}
	roleID, err := roleResult.LastInsertId()
	if err != nil {
		return err
	}

	permissionIDs, err := r.ensureAdminPermissions(ctx, tx)
	if err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, roleID, permissionID); err != nil {
			return err
		}
	}

	userID, err := r.createUser(ctx, tx, admin, allowExistingUser)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_domain (user_id, domain_id) VALUES (?, ?)`, userID, domainID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, roleID)
	return err
}

func (r *MySQLACLRepository) ensureAdminPermissions(ctx context.Context, tx *sql.Tx) ([]int64, error) {
	permissions := []struct{ name, code, description string }{
		{"Manage domains", "domains.manage", "Create, update, and delete domains"},
		{"Manage roles", "roles.manage", "Create, update, and delete domain roles"},
		{"Manage permissions", "permissions.manage", "Create, update, and delete permissions"},
	}
	ids := make([]int64, 0, len(permissions))
	for _, permission := range permissions {
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO permissions (name, code, description) VALUES (?, ?, ?)`, permission.name, permission.code, permission.description); err != nil {
			return nil, err
		}
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM permissions WHERE code = ?`, permission.code).Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *MySQLACLRepository) createUser(ctx context.Context, tx *sql.Tx, user *domainUser.User, allowExisting bool) (int64, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO users (email, password, first_name, last_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, user.Email, user.Password, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)
	if err == nil {
		id, idErr := result.LastInsertId()
		if idErr == nil {
			user.ID = id
		}
		return id, idErr
	}
	if !allowExisting {
		return 0, err
	}
	var id int64
	if queryErr := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, user.Email).Scan(&id); queryErr != nil {
		return 0, err
	}
	return id, nil
}

func (r *MySQLACLRepository) ListDomains(ctx context.Context) ([]domainACL.Domain, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, created_at, updated_at FROM domains ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var domains []domainACL.Domain
	for rows.Next() {
		var domain domainACL.Domain
		if err := rows.Scan(&domain.ID, &domain.Name, &domain.Slug, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, rows.Err()
}

func (r *MySQLACLRepository) GetDomain(ctx context.Context, id int64) (*domainACL.Domain, error) {
	domain := &domainACL.Domain{}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, slug, created_at, updated_at FROM domains WHERE id = ?`, id).Scan(&domain.ID, &domain.Name, &domain.Slug, &domain.CreatedAt, &domain.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainACL.ErrDomainNotFound
	}
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (r *MySQLACLRepository) UpdateDomain(ctx context.Context, domain *domainACL.Domain) error {
	result, err := r.db.ExecContext(ctx, `UPDATE domains SET name = ?, slug = ? WHERE id = ?`, domain.Name, domain.Slug, domain.ID)
	if err != nil {
		return err
	}
	return notFoundIfNoRows(result, domainACL.ErrDomainNotFound)
}

func (r *MySQLACLRepository) DeleteDomain(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM domains WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return notFoundIfNoRows(result, domainACL.ErrDomainNotFound)
}

func (r *MySQLACLRepository) CreateRole(ctx context.Context, role *domainACL.Role) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `INSERT INTO roles (domain_id, name, code, description) VALUES (?, ?, ?, ?)`, role.DomainID, role.Name, role.Code, nullableString(role.Description))
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		role.ID = id
		return replaceRolePermissions(ctx, tx, id, role.PermissionIDs)
	})
}

func (r *MySQLACLRepository) ListRoles(ctx context.Context, domainID int64) ([]domainACL.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, domain_id, name, code, description FROM roles WHERE domain_id = ? ORDER BY id`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []domainACL.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		role.PermissionIDs, err = r.rolePermissionIDs(ctx, r.db, role.ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, *role)
	}
	return roles, rows.Err()
}

func (r *MySQLACLRepository) GetRole(ctx context.Context, id int64) (*domainACL.Role, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, domain_id, name, code, description FROM roles WHERE id = ?`, id)
	role, err := scanRole(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainACL.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	role.PermissionIDs, err = r.rolePermissionIDs(ctx, r.db, id)
	return role, err
}

func (r *MySQLACLRepository) UpdateRole(ctx context.Context, role *domainACL.Role) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE roles SET domain_id = ?, name = ?, code = ?, description = ? WHERE id = ?`, role.DomainID, role.Name, role.Code, nullableString(role.Description), role.ID)
		if err != nil {
			return err
		}
		if err := notFoundIfNoRows(result, domainACL.ErrRoleNotFound); err != nil {
			return err
		}
		return replaceRolePermissions(ctx, tx, role.ID, role.PermissionIDs)
	})
}

func (r *MySQLACLRepository) DeleteRole(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return notFoundIfNoRows(result, domainACL.ErrRoleNotFound)
}

func (r *MySQLACLRepository) CreatePermission(ctx context.Context, permission *domainACL.Permission) error {
	result, err := r.db.ExecContext(ctx, `INSERT INTO permissions (name, code, description) VALUES (?, ?, ?)`, permission.Name, permission.Code, nullableString(permission.Description))
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	permission.ID = id
	return err
}

func (r *MySQLACLRepository) ListPermissions(ctx context.Context) ([]domainACL.Permission, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, code, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var permissions []domainACL.Permission
	for rows.Next() {
		permission, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, *permission)
	}
	return permissions, rows.Err()
}

func (r *MySQLACLRepository) GetPermission(ctx context.Context, id int64) (*domainACL.Permission, error) {
	permission, err := scanPermission(r.db.QueryRowContext(ctx, `SELECT id, name, code, description FROM permissions WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainACL.ErrPermissionNotFound
	}
	return permission, err
}

func (r *MySQLACLRepository) UpdatePermission(ctx context.Context, permission *domainACL.Permission) error {
	result, err := r.db.ExecContext(ctx, `UPDATE permissions SET name = ?, code = ?, description = ? WHERE id = ?`, permission.Name, permission.Code, nullableString(permission.Description), permission.ID)
	if err != nil {
		return err
	}
	return notFoundIfNoRows(result, domainACL.ErrPermissionNotFound)
}

func (r *MySQLACLRepository) DeletePermission(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return notFoundIfNoRows(result, domainACL.ErrPermissionNotFound)
}

func (r *MySQLACLRepository) IsAuthAdmin(ctx context.Context, userID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(
        SELECT 1 FROM user_roles ur
        JOIN roles r ON r.id = ur.role_id
        JOIN domains d ON d.id = r.domain_id
        WHERE ur.user_id = ? AND d.slug = ? AND r.code = 'admin'
    )`, userID, authDomainSlug).Scan(&exists)
	return exists == 1, err
}

type scanner interface{ Scan(...any) error }

func scanRole(row scanner) (*domainACL.Role, error) {
	var role domainACL.Role
	var description sql.NullString
	err := row.Scan(&role.ID, &role.DomainID, &role.Name, &role.Code, &description)
	role.Description = description.String
	return &role, err
}
func scanPermission(row scanner) (*domainACL.Permission, error) {
	var permission domainACL.Permission
	var description sql.NullString
	err := row.Scan(&permission.ID, &permission.Name, &permission.Code, &description)
	permission.Description = description.String
	return &permission, err
}
func (r *MySQLACLRepository) rolePermissionIDs(ctx context.Context, query interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, roleID int64) ([]int64, error) {
	rows, err := query.QueryContext(ctx, `SELECT permission_id FROM role_permissions WHERE role_id = ? ORDER BY permission_id`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
func replaceRolePermissions(ctx context.Context, tx *sql.Tx, roleID int64, permissionIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, roleID, permissionID); err != nil {
			return err
		}
	}
	return nil
}
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func notFoundIfNoRows(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound
	}
	return nil
}
func (r *MySQLACLRepository) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("acl transaction: %w", err)
	}
	return tx.Commit()
}
