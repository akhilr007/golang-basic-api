package task

import (
	"context"
	"log/slog"
	"strings"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetAll(ctx context.Context, userID int) ([]Task, error) {
	return s.repo.GetAll(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id, userID int) (Task, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, userID int, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	return s.repo.Create(ctx, userID, title)
}

func (s *Service) Update(ctx context.Context, id int, userID int, title *string, done *bool) (Task, error) {
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return Task{}, ErrInvalidTitle
		}
		title = &trimmed
	}

	return s.repo.Update(ctx, id, userID, title, done)
}

func (s *Service) Delete(ctx context.Context, id, userID int) error {
	return s.repo.Delete(ctx, id, userID)
}
