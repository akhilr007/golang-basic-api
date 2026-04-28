package auth

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}
