package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	logger           *slog.Logger
}

var (
	ErrAuthInput        = errors.New("email and password required")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
)

func NewService(userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, log *slog.Logger) *Service {
	return &Service{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		logger:           log,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	log := s.logger.With(
		"service", "auth",
		"action", "register",
		"email", email,
	)

	if email == "" || password == "" {
		log.Error("invalid input: email or password missing")
		return User{}, ErrAuthInput
	}

	if len(password) < 6 {
		log.Error("password too short", "length", len(password))
		return User{}, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("password hashing failed", "error", err)
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.CreateUser(ctx, email, string(hash))
	if err != nil {
		log.Error("user creation failed", "error", err)
		return User{}, fmt.Errorf("create user: %w", err)
	}

	log.Info("user successfully created", "user_id", user.ID)

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, string, time.Time, error) {

	email = strings.TrimSpace(strings.ToLower(email))

	log := s.logger.With(
		"service", "auth",
		"action", "login",
		"email", email,
	)

	if email == "" || password == "" {
		log.Error("invalid input: email or password missing")
		return User{}, "", "", time.Time{}, errors.New("email and password required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Error("invalid login attempt")
			return User{}, "", "", time.Time{}, errors.New("invalid credentials")
		}

		log.Error("database error", "error", err)
		return User{}, "", "", time.Time{}, fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error("invalid login attempt")
		return User{}, "", "", time.Time{}, errors.New("invalid credentials")
	}

	log.Info("login successful", "user_id", user.ID)

	familyID := uuid.NewString()

	raw, hash, jti, err := GenerateRefreshToken()
	if err != nil {
		log.Error("refresh token generation failed", "error", err)
		return User{}, "", "", time.Time{}, fmt.Errorf("internal server error")
	}

	expiry := time.Now().Add(7 * 24 * time.Hour)

	err = s.refreshTokenRepo.SaveToken(ctx, user.ID, hash, jti, familyID, expiry)
	if err != nil {
		log.Error("refresh token save", "error", err)
		return User{}, "", "", time.Time{}, fmt.Errorf("internal server error")
	}

	accessToken, err := GenerateAccessToken(user.ID)
	if err != nil {
		log.Error("access token", "error", err)
		return User{}, "", "", time.Time{}, fmt.Errorf("internal server error")
	}

	return user, accessToken, raw, expiry, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (string, string, error) {
	if len(rawToken) < 6 {
		s.logger.Warn("invalid refresh token length")
		return "", "", errors.New("invalid token")
	}

	log := s.logger.With(
		"service", "auth",
		"action", "refresh",
		"token_prefix", rawToken[:6],
	)

	hashBytes := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(hashBytes[:])

	refreshToken, err := s.refreshTokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}

	// reuse detection
	if refreshToken.RevokedAt != nil {
		// token already used, possible theft
		_ = s.refreshTokenRepo.RevokeFamily(ctx, refreshToken.FamilyID)
		return "", "", errors.New("session compromised")
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return "", "", errors.New("expired token")
	}

	newRaw, newHash, newJTI, err := GenerateRefreshToken()
	if err != nil {
		log.Error("refresh token generation failed", "error", err)
		return "", "", errors.New("internal server error")
	}

	_, err = s.refreshTokenRepo.RotateToken(ctx, hash, newHash, newJTI, time.Now().Add(7*24*time.Hour))
	if err != nil {
		if errors.Is(err, ErrTokenAlreadyRevoked) {
			_ = s.refreshTokenRepo.RevokeFamily(ctx, refreshToken.FamilyID)
			return "", "", errors.New("session compromised")
		}

		log.Error("refresh token save", "error", err)
		return "", "", errors.New("something went wrong")
	}

	accessToken, err := GenerateAccessToken(refreshToken.UserID)
	if err != nil {
		log.Error("access token", "error", err)
		return "", "", errors.New("internal server error")
	}

	return accessToken, newRaw, nil
}
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	hashBytes := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(hashBytes[:])

	rt, err := s.refreshTokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return errors.New("invalid token")
	}

	if err := s.refreshTokenRepo.RevokeByJTI(ctx, rt.JTI); err != nil {
		s.logger.Error("failed to revoke token", "error", err)
		return errors.New("internal server error")
	}

	return nil
}
