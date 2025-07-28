package domain

import "context"

type UserRepository interface {
	GetAll(ctx context.Context) ([]*User, error)           // Adjust: return slice of pointer, tipe User dari domain, bukan model.User
	GetByID(ctx context.Context, id string) (*User, error) // Adjust: id jadi string, bukan ObjectID
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, id string, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}
