package user

import (
	"context"

	domainUser "go-ddd/internal/domain/user"
)

type TokenService interface {
	GenerateToken(userID int64, email string) (string, error)
}

type UserService struct {
	userRepo     domainUser.UserRepository
	tokenService TokenService
}

func NewUserService(userRepo domainUser.UserRepository, tokenService TokenService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainUser.ErrUserAlreadyExists
	}

	newUser, err := domainUser.NewUser(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	token, err := s.tokenService.GenerateToken(newUser.ID, newUser.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:        newUser.ID,
			Email:     newUser.Email,
			FirstName: newUser.FirstName,
			LastName:  newUser.LastName,
			CreatedAt: newUser.CreatedAt,
		},
	}, nil
}

func (s *UserService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == domainUser.ErrUserNotFound {
			return nil, domainUser.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.CheckPassword(req.Password) {
		return nil, domainUser.ErrInvalidCredentials
	}

	token, err := s.tokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*UserDTO, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}, nil
}
