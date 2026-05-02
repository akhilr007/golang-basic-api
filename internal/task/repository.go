package task

import (
	"context"
	"errors"
)

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("title is required")
)

type Repository interface {
	GetAll(ctx context.Context, userID int) ([]Task, error)
	GetByID(ctx context.Context, id, userID int) (Task, error)
	Create(ctx context.Context, userID int, title string) (Task, error)
	Update(ctx context.Context, id, userID int, title *string, done *bool) (Task, error)
	Delete(ctx context.Context, id, userID int) error
}
