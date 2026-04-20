package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Store struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewStore() *Store {
	return &Store{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (s *Store) GetAll() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *Store) GetByID(id int) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (s *Store) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID

	task := Task{
		ID:        id,
		Title:     strings.TrimSpace(title),
		Done:      false,
		CreatedAt: time.Now(),
	}
	s.tasks[id] = task
	s.nextID++
	return task
}

func (s *Store) Update(id int, title *string, done *bool) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
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

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}
