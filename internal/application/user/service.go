package user

import (
	"context"
	"fmt"
	"log"

	domainMail "go-ddd/internal/domain/mail"
	domainUser "go-ddd/internal/domain/user"
)

type TokenService interface {
	GenerateToken(userID int64, email string) (string, error)
}

type UserService struct {
	userRepo     domainUser.UserRepository
	tokenService TokenService
	mailer       domainMail.Mailer
}

func NewUserService(userRepo domainUser.UserRepository, tokenService TokenService, mailer domainMail.Mailer) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenService: tokenService,
		mailer:       mailer,
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

	// Send welcome email asynchronously
	if s.mailer != nil {
		go func(toEmail, firstName string) {
			msg := domainMail.Message{
				To:       []domainMail.Address{{Name: firstName, Email: toEmail}},
				Subject:  "Welcome to Go-DDD App!",
				TextBody: fmt.Sprintf("Hello %s,\n\nWelcome to Go-DDD API Application!", firstName),
				HTMLBody: fmt.Sprintf("<h1>Welcome %s!</h1><p>Thank you for registering on Go-DDD API Application.</p>", firstName),
			}
			if err := s.mailer.Send(context.Background(), msg); err != nil {
				log.Printf("Failed to send welcome email to %s: %v", toEmail, err)
			}
		}(newUser.Email, newUser.FirstName)
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
