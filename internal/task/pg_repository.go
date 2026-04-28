package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/akhilr007/tasks/internal/db"
	"github.com/jackc/pgx/v5"
)

type PGRepository struct {
	db db.DBTX
}

func NewPGRepository(db db.DBTX) *PGRepository {
	return &PGRepository{
		db: db,
	}
}

func (r *PGRepository) GetAll(ctx context.Context) ([]Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `SELECT id, title, done, created_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
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

func (r *PGRepository) GetByID(ctx context.Context, id int) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var t Task
	err := r.db.QueryRow(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`,
		id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Create(ctx context.Context, title string) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	title = strings.TrimSpace(title)

	if title == "" {
		return Task{}, errors.New("title cannot be empty")
	}

	var t Task

	err := r.db.QueryRow(ctx,
		`INSERT INTO tasks (title)
		VALUES ($1)
		RETURNING id, title, done, created_at`, title,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Update(ctx context.Context, id int, title *string, done *bool) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		title = &trimmed
	}

	var t Task
	err := r.db.QueryRow(ctx,
		`UPDATE tasks SET
			title = COALESCE($1, title),
			done = COALESCE($2, done)
		WHERE id = $3
		RETURNING id, title, done, created_at`,
		title, done, id,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd, err := r.db.Exec(ctx,
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
