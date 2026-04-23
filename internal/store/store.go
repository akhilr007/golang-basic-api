package store

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/akhilr007/tasks/internal/model"
)

var ErrNotFound = errors.New("task not found")

type TaskStore interface {
	GetAll(context.Context) ([]model.Task, error)
	GetByID(context.Context, int) (model.Task, error)
	Create(context.Context, string) (model.Task, error)
	Update(context.Context, int, *string, *bool) (model.Task, error)
	Delete(context.Context, int) error
}

type Store struct {
	mu     sync.RWMutex
	tasks  map[int]model.Task
	nextID int
}

func NewStore() *Store {
	return &Store{
		tasks:  make(map[int]model.Task),
		nextID: 1,
	}
}

func (s *Store) GetAll(ctx context.Context) ([]model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]model.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Store) GetByID(ctx context.Context, id int) (model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return model.Task{}, ErrNotFound
	}

	return task, nil
}

func (s *Store) Create(ctx context.Context, title string) (model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID

	task := model.Task{
		ID:        id,
		Title:     strings.TrimSpace(title),
		Done:      false,
		CreatedAt: time.Now(),
	}
	s.tasks[id] = task
	s.nextID++
	return task, nil
}

func (s *Store) Update(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return model.Task{}, ErrNotFound
	}

	if title != nil {
		task.Title = strings.TrimSpace(*title)
	}
	if done != nil {
		task.Done = *done
	}
	s.tasks[id] = task

	return task, nil
}

func (s *Store) Delete(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}
