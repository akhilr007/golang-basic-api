package routes

import (
	"net/http"

	"github.com/akhilr007/tasks/internal/app"
	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/task"
	"github.com/akhilr007/tasks/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Mount(application *app.App) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	taskHandler := task.NewHandler(application.TaskService, application.Logger)
	authHandler := auth.NewHandler(application.AuthService, application.Logger)

	taskHandler.Routes(r)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	return r
}
