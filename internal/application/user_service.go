// application/user_service.go
package application

import "go-ddd/internal/domain"

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(user domain.User) error {
	// Add business logic for user registration
	return s.repo.CreateUser(user)
}

func (s *UserService) AuthenticateUser(username, password string) (domain.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return domain.User{}, err
	}
	// Add password verification logic
	return user, nil
}
