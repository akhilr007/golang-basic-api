package app

import (
	"log/slog"

	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/db"
	"github.com/akhilr007/tasks/internal/task"
)

type App struct {
	Logger *slog.Logger

	TaskService *task.Service
	AuthService *auth.Service
}

func New(logger *slog.Logger, database db.DBTX) *App {
	taskRepo := task.NewPGRepository(database)
	authRepo := auth.NewPGRepository(database)
	taskService := task.NewService(taskRepo, logger)
	authService := auth.NewService(authRepo, logger)

	return &App{
		Logger:      logger,
		TaskService: taskService,
		AuthService: authService,
	}
}
