// application/user_service.go
package application

import (
	"go-ddd/internal/domain/entity"
	"go-ddd/internal/domain/interfaces/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(user entity.User) error {
	// Add business logic for user registration
	return s.repo.CreateUser(user)
}

func (s *UserService) AuthenticateUser(username, password string) (entity.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return entity.User{}, err
	}
	// Add password verification logic
	return user, nil
}
