package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/akhilr007/tasks/internal/config"
	"github.com/akhilr007/tasks/internal/db"
	"github.com/akhilr007/tasks/internal/handler"
	"github.com/akhilr007/tasks/internal/store"
)

func main() {
	
	cfg := config.Load()
	
	pool, err := db.NewPool(cfg.DB.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	
	
	// get my store
	// store := store.NewStore()
	pgStore := store.NewPGStore(pool)
	handler := handler.NewHandler(pgStore)

	// add a router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handler.Routes(r)

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
		log.Printf("Tasks api running on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}
