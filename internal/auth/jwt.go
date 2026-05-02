package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func ConfigureJWTSecret(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}

	jwtSecret = []byte(secret)
	return nil
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int) (string, error) {

	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("jwt secret is not configured")
	}

	now := time.Now()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-tasks-app",
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

func ValidateToken(tokenStr string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, fmt.Errorf("jwt secret is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
