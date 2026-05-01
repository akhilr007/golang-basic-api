package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrTokenNotFound = errors.New("refresh token not found")

func (r *PGRepository) SaveToken(ctx context.Context, userID int, hash, jti, familyID string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, jti, family_id, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, hash, jti, familyID, expiry,
	)

	if err != nil {
		return fmt.Errorf("save token (user_id=%d, jti=%s): %w", userID, jti, err)
	}

	return nil
}

func (r *PGRepository) GetByHash(ctx context.Context, hash string) (RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var rt RefreshToken

	err := r.db.QueryRow(ctx,
		`SELECT user_id, jti, family_id, expires_at, revoked_at
		FROM refresh_tokens WHERE token_hash=$1`,
		hash,
	).Scan(&rt.UserID, &rt.JTI, &rt.FamilyID, &rt.ExpiresAt, &rt.RevokedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshToken{}, ErrTokenNotFound
		}
		return RefreshToken{}, fmt.Errorf("get token by hash: %w", err)
	}

	return rt, nil
}

func (r *PGRepository) RevokeByJTI(ctx context.Context, jti string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens
			 SET revoked_at = NOW()
			 WHERE jti = $1 AND revoked_at IS NULL`,
		jti,
	)
	if err != nil {
		return fmt.Errorf("revoke by jti: %w", err)
	}

	return nil
}

func (r *PGRepository) RevokeFamily(ctx context.Context, familyID string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=NOW() WHERE family_id=$1 AND revoked_at IS NULL`,
		familyID,
	)

	if err != nil {
		return fmt.Errorf("revoke family: %w", err)
	}

	return nil
}
