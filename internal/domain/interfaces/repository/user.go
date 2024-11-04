package repository

import (
	"go-ddd/internal/domain/entity"
)

type UserRepository interface {
	CreateUser(user entity.User) error
	GetUserByID(id int) (entity.User, error)
	UpdateUser(user entity.User) error
	DeleteUser(id int) error
	GetUserByUsername(username string) (entity.User, error)
}
