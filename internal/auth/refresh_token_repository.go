package auth

import (
	"context"
	"time"
)

type RefreshTokenRepository interface {
	SaveToken(ctx context.Context, userID int, hash, jti, familyID string, time time.Time) error
	GetByHash(ctx context.Context, hash string) (RefreshToken, error)
	RotateToken(ctx context.Context, oldHash, newHash, newJTI string, expiry time.Time) (int, error)
	RevokeByJTI(ctx context.Context, jti string) error
	RevokeFamily(ctx context.Context, familyID string) error
}
