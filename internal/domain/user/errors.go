package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidEmail       = errors.New("invalid email address format")
	ErrInvalidPassword    = errors.New("password must be at least 6 characters long")
	ErrEmptyFields        = errors.New("first name and last name cannot be empty")
)
