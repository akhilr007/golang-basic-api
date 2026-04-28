package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akhilr007/tasks/internal/app"
	"github.com/akhilr007/tasks/internal/config"
	"github.com/akhilr007/tasks/internal/db"
	"github.com/akhilr007/tasks/internal/logger"
	"github.com/akhilr007/tasks/internal/routes"
)

func main() {

	cfg := config.Load()

	log := logger.New(cfg.Logger)

	pool, err := db.NewPool(cfg.DB.URL)
	if err != nil {
		log.Error("failed to initialize db pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	application := app.New(log, pool)
	r := routes.Mount(application)

	// how to create a http server in go
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Info("Tasks api running", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen error", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
	} else {
		log.Info("server stopped gracefully")
	}
}
