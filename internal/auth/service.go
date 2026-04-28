package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, log *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: log,
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
		return User{}, errors.New("email and password required")
	}

	if len(password) < 6 {
		log.Error("password too short", "length", len(password))
		return User{}, errors.New("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("password hashing failed", "error", err)
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, string(hash))
	if err != nil {
		log.Error("user creation failed", "error", err)
		return User{}, fmt.Errorf("create user: %w", err)
	}

	log.Info("user successfully created", "user_id", user.ID)

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, error) {

	email = strings.TrimSpace(strings.ToLower(email))

	log := s.logger.With(
		"service", "auth",
		"action", "login",
		"email", email,
	)

	if email == "" || password == "" {
		log.Error("invalid input: email or password missing")
		return User{}, errors.New("email and password required")
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Error("invalid login attempt")
			return User{}, errors.New("invalid credentials")
		}

		log.Error("database error", "error", err)
		return User{}, fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Error("invalid login attempt")
		return User{}, errors.New("invalid credentials")
	}

	log.Info("login successful", "user_id", user.ID)

	return user, nil
}
