package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Create the HTTP server
	server := http.Server{Addr: ":3000", Handler: service()}

	// Create a context that listens for SIGINT and SIGTERM signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run the server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Wait until the context is canceled (signal received)
	<-ctx.Done()

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sup\n"))
	})

	r.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)

		w.Write([]byte("ok\n"))
	})

	return r
}
