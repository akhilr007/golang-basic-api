package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

func GenerateRefreshToken() (raw string, hash string, jti string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	raw = hex.EncodeToString(b) // convert byte to string for user

	h := sha256.Sum256([]byte(raw)) // hash the byte
	hash = hex.EncodeToString(h[:]) // convert it to string

	jti = uuid.NewString()
	return
}
