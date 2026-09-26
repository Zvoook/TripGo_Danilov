package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	config "github.com/Zvoook/TripGo_Danilov/internal/config"
	"github.com/Zvoook/TripGo_Danilov/internal/httpapi"
	"github.com/Zvoook/TripGo_Danilov/internal/postgres"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("postgres connection: %w", err)
	}
	defer pool.Close()
	fmt.Println("Successful connection")

	handler := httpapi.NewHandler(pool, cfg.DatabaseQueryTimeout)

	router := chi.NewRouter()
	router.Get("/health", handler.Health)
	router.Get("/ready", handler.Ready)
	router.Get("/debug/slow", handler.Waiting)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server falls due to server error")
			return fmt.Errorf("server issue: %w", err)
		} else {
			return nil
		}
	case <-signalCtx.Done():
		fmt.Printf("\nShutdown the server...\n")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.ShutdownTimeout,
		)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			server.Close()
			fmt.Println("Server stop were interrupted due to shutdown error")
			return fmt.Errorf("server issue: %w", err)
		} else {
			fmt.Println("Server stop successfully")
		}
	}
	return nil
}
