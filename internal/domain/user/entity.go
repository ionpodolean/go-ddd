package user

import (
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(email, rawPassword, firstName, lastName string) (*User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if err := validateEmail(cleanEmail); err != nil {
		return nil, err
	}

	if len(rawPassword) < 6 {
		return nil, ErrInvalidPassword
	}

	cleanFirstName := strings.TrimSpace(firstName)
	cleanLastName := strings.TrimSpace(lastName)
	if cleanFirstName == "" || cleanLastName == "" {
		return nil, ErrEmptyFields
	}

	hashedPassword, err := hashPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:     cleanEmail,
		Password:  hashedPassword,
		FirstName: cleanFirstName,
		LastName:  cleanLastName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) CheckPassword(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPassword))
	return err == nil
}

func (u *User) SetPassword(rawPassword string) error {
	if len(rawPassword) < 6 {
		return ErrInvalidPassword
	}
	hashedPassword, err := hashPassword(rawPassword)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	u.UpdatedAt = time.Now()
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}
