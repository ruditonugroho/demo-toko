package domain

type AuthService interface {
	Register(name, email, password string) (*User, error)
	Login(email, password string) (string, error) // JWT token
}
