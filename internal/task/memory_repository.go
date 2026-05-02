package task

import (
	"context"
	"strings"
	"sync"
	"time"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (r *MemoryRepository) GetAll(ctx context.Context, userID int) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		if t.UserID != userID {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id, userID int) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok || task.UserID != userID {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (r *MemoryRepository) Create(ctx context.Context, userID int, title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	id := r.nextID
	task := Task{
		ID:        id,
		UserID:    userID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
	}
	r.tasks[id] = task
	r.nextID++
	return task, nil
}

func (r *MemoryRepository) Update(ctx context.Context, id, userID int, title *string, done *bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok || task.UserID != userID {
		return Task{}, ErrNotFound
	}

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return Task{}, ErrInvalidTitle
		}
		task.Title = trimmed
	}
	if done != nil {
		task.Done = *done
	}
	r.tasks[id] = task

	return task, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id, userID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok || task.UserID != userID {
		return ErrNotFound
	}

	delete(r.tasks, id)
	return nil
}
