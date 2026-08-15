package user

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
