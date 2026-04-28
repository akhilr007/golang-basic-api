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

func (s *Service) GetAll(ctx context.Context) ([]Task, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, title string) (Task, error) {
	title = strings.TrimSpace(title)
	return s.repo.Create(ctx, title)
}

func (s *Service) Update(ctx context.Context, id int, title *string, done *bool) (Task, error) {
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		title = &trimmed
	}

	return s.repo.Update(ctx, id, title, done)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
