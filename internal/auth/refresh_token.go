package auth

import "time"

type RefreshToken struct {
	UserID    int
	JTI       string
	FamilyID  string
	ExpiresAt time.Time
	RevokedAt *time.Time
}
