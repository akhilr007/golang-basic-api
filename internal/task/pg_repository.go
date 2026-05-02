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

func (r *PGRepository) GetAll(ctx context.Context, userID int) ([]Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx,
		`SELECT id, title, user_id, done, created_at
		FROM tasks
		WHERE user_id=$1
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.UserID, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *PGRepository) GetByID(ctx context.Context, id, userID int) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var t Task
	err := r.db.QueryRow(ctx,
		`SELECT id, title, user_id, done, created_at FROM tasks WHERE id = $1 AND user_id=$2`,
		id, userID,
	).Scan(&t.ID, &t.Title, &t.UserID, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Create(ctx context.Context, userID int, title string) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	title = strings.TrimSpace(title)

	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	var t Task

	err := r.db.QueryRow(ctx,
		`INSERT INTO tasks (title, user_id)
		VALUES ($1, $2)
		RETURNING id, title, user_id, done, created_at`,
		title, userID,
	).Scan(&t.ID, &t.Title, &t.UserID, &t.Done, &t.CreatedAt)

	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Update(ctx context.Context, id int, userID int, title *string, done *bool) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return Task{}, ErrInvalidTitle
		}
		title = &trimmed
	}

	var t Task
	err := r.db.QueryRow(ctx,
		`UPDATE tasks SET
			title = COALESCE($1, title),
			done = COALESCE($2, done)
		WHERE id = $3 AND user_id = $4
		RETURNING id, title, user_id, done, created_at`,
		title, done, id, userID,
	).Scan(&t.ID, &t.Title, &t.UserID, &t.Done, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return t, nil
}

func (r *PGRepository) Delete(ctx context.Context, id int, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd, err := r.db.Exec(ctx,
		`DELETE FROM tasks WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
