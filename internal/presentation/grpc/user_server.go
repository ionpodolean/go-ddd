package grpc

import (
	"context"
	"errors"

	userv1 "go-ddd/api/gen/go/user/v1"
	appUser "go-ddd/internal/application/user"
	domainUser "go-ddd/internal/domain/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	userService *appUser.UserService
}

func NewUserServer(userService *appUser.UserService) *UserServer {
	return &UserServer{userService: userService}
}

func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.AuthResponse, error) {
	response, err := s.userService.Register(ctx, appUser.RegisterRequest{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		return nil, userServiceError(err)
	}
	return authResponse(response), nil
}

func (s *UserServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.AuthResponse, error) {
	response, err := s.userService.Login(ctx, appUser.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, userServiceError(err)
	}
	return authResponse(response), nil
}

func (s *UserServer) GetProfile(ctx context.Context, _ *userv1.GetProfileRequest) (*userv1.User, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, userServiceError(err)
	}
	return userMessage(profile), nil
}

func authResponse(response *appUser.AuthResponse) *userv1.AuthResponse {
	return &userv1.AuthResponse{Token: response.Token, User: userMessage(&response.User)}
}

func userMessage(user *appUser.UserDTO) *userv1.User {
	return &userv1.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}

func userServiceError(err error) error {
	switch {
	case errors.Is(err, domainUser.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainUser.ErrInvalidEmail), errors.Is(err, domainUser.ErrInvalidPassword), errors.Is(err, domainUser.ErrEmptyFields):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainUser.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domainUser.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
