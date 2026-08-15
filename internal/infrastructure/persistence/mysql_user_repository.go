package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainUser "go-ddd/internal/domain/user"
	"go-ddd/internal/infrastructure/persistence/sqlc"
)

type MySQLUserRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *MySQLUserRepository) Create(ctx context.Context, u *domainUser.User) error {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:     u.Email,
		Password:  sql.NullString{String: u.Password, Valid: u.Password != ""},
		FirstName: sql.NullString{String: u.FirstName, Valid: u.FirstName != ""},
		LastName:  sql.NullString{String: u.LastName, Valid: u.LastName != ""},
		CreatedAt: sql.NullTime{Time: u.CreatedAt, Valid: !u.CreatedAt.IsZero()},
		UpdatedAt: sql.NullTime{Time: u.UpdatedAt, Valid: !u.UpdatedAt.IsZero()},
	})
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		u.ID = id
	}

	return nil
}

func (r *MySQLUserRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, err
	}

	return mapSqlcUserToDomain(&row), nil
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id int64) (*domainUser.User, error) {
	row, err := r.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, err
	}

	return mapSqlcUserToDomain(&row), nil
}

func (r *MySQLUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.queries.ExistsUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func mapSqlcUserToDomain(u *sqlc.User) *domainUser.User {
	var createdAt, updatedAt time.Time
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		updatedAt = u.UpdatedAt.Time
	}

	return &domainUser.User{
		ID:        int64(u.ID),
		Email:     u.Email,
		Password:  u.Password.String,
		FirstName: u.FirstName.String,
		LastName:  u.LastName.String,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
