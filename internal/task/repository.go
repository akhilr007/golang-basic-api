package task

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

type Repository interface {
	GetAll(context.Context) ([]Task, error)
	GetByID(context.Context, int) (Task, error)
	Create(context.Context, string) (Task, error)
	Update(context.Context, int, *string, *bool) (Task, error)
	Delete(context.Context, int) error
}
