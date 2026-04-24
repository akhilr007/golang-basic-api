package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/akhilr007/tasks/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGStore struct {
	db DBTX
}

type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func NewPGStore(db DBTX) *PGStore {
	return &PGStore{
		db: db,
	}
}

func (s *PGStore) GetAll(ctx context.Context) ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, title, done, created_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *PGStore) GetByID(ctx context.Context, id int) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var t model.Task
	err := s.db.QueryRow(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`,
		id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Task{}, ErrNotFound
		}
		return model.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return t, nil
}

func (s *PGStore) Create(ctx context.Context, title string) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	title = strings.TrimSpace(title)

	if title == "" {
		return model.Task{}, errors.New("title cannot be empty")
	}

	var t model.Task

	err := s.db.QueryRow(ctx,
		`INSERT INTO tasks (title)
		VALUES ($1)
		RETURNING id, title, done, created_at`, title,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}

func (s *PGStore) Update(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		title = &trimmed
	}

	var t model.Task
	err := s.db.QueryRow(ctx,
		`UPDATE tasks SET
			title = COALESCE($1, title),
			done = COALESCE($2, done)
		WHERE id = $3
		RETURNING id, title, done, created_at`,
		title, done, id,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Task{}, ErrNotFound
		}
		return model.Task{}, fmt.Errorf("update task: %w", err)
	}

	return t, nil
}

func (s *PGStore) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd, err := s.db.Exec(ctx,
		`DELETE FROM tasks WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
