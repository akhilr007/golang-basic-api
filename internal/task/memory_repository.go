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

func (r *MemoryRepository) GetAll(ctx context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id int) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (r *MemoryRepository) Create(ctx context.Context, title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	task := Task{
		ID:        id,
		Title:     strings.TrimSpace(title),
		Done:      false,
		CreatedAt: time.Now(),
	}
	r.tasks[id] = task
	r.nextID++
	return task, nil
}

func (r *MemoryRepository) Update(ctx context.Context, id int, title *string, done *bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if title != nil {
		task.Title = strings.TrimSpace(*title)
	}
	if done != nil {
		task.Done = *done
	}
	r.tasks[id] = task

	return task, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrNotFound
	}

	delete(r.tasks, id)
	return nil
}
