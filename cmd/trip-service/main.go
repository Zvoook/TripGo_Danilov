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
	"github.com/Zvoook/TripGo_Danilov/internal/infra/telemetry"
)

func main() {
	if err := run(); err != nil {
		telemetry.NewLogger("info", os.Stdout).Error("service failed", "error", err)
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
	app.Logger.InfoContext(signalCtx, "service starting", "address", cfg.HTTPAddr)
	defer app.Logger.Info("service stopped")
	defer app.CloseConnection()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Server.ListenAndServe()
	}()

	var runErr error
	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("serve HTTP: %w", err)
		}
	case <-signalCtx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	shutdownErr := app.Server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, app.CloseServer())
	}
	if err := app.ShutdownTelemetry(shutdownCtx); err != nil {
		app.Logger.ErrorContext(shutdownCtx, "shutdown telemetry failed", "error", err)
	}
	return errors.Join(runErr, shutdownErr)
}
