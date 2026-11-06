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
	"github.com/Zvoook/TripGo_Danilov/internal/container"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	app, err := container.New(cfg, signalCtx)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	defer app.CloseConnection()
	defer app.CloseServer()

	serverErr := make(chan error, 1)
	//TODO: понять надо ли пробрасывать канал с ошибкой внутрь container.New()

	go func() {
		serverErr <- app.Server.ListenAndServe()
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
		err := app.Server.Shutdown(shutdownCtx)
		if err != nil {
			app.CloseServer()
			fmt.Println("Server stop were interrupted due to shutdown error")
			return fmt.Errorf("server issue: %w", err)
		} else {
			fmt.Println("Server stop successfully")
		}
	}
	return nil
}
