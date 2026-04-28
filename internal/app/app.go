package app

import (
	"log/slog"

	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/db"
	"github.com/akhilr007/tasks/internal/task"
)

type App struct {
	Logger      *slog.Logger
	TaskHandler *task.Handler
	AuthHandler *auth.Handler
}

func New(logger *slog.Logger, database db.DBTX) *App {
	taskRepo := task.NewPGRepository(database)
	authRepo := auth.NewPGRepository(database)
	authService := auth.NewService(authRepo, logger)

	return &App{
		Logger:      logger,
		TaskHandler: task.NewHandler(taskRepo, logger),
		AuthHandler: auth.NewHandler(authService, logger),
	}
}
