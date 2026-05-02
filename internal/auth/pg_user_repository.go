package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/akhilr007/tasks/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGRepository struct {
	db db.DBTX
}

func NewPGRepository(db db.DBTX) *PGRepository {
	return &PGRepository{
		db: db,
	}
}

func (r *PGRepository) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var u User

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, is_verified, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsVerified, &u.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (r *PGRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var u User

	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, is_verified, created_at
		FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsVerified, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}
