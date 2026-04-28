package routes

import (
	"net/http"

	"github.com/akhilr007/tasks/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Mount(application *app.App) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	application.TaskHandler.Routes(r)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", application.AuthHandler.Register)
		r.Post("/login", application.AuthHandler.Login)
	})

	return r
}
