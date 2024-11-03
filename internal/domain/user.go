// domain/user.go
package domain

type User struct {
	ID        int
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type UserRepository interface {
	CreateUser(user User) error
	GetUserByID(id int) (User, error)
	UpdateUser(user User) error
	DeleteUser(id int) error
	GetUserByUsername(username string) (User, error)
}
