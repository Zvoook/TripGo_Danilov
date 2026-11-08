package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/Zvoook/TripGo_Danilov/internal/config"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	httptransport "github.com/Zvoook/TripGo_Danilov/internal/http"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	"github.com/Zvoook/TripGo_Danilov/internal/http/middleware"
	httpcreate "github.com/Zvoook/TripGo_Danilov/internal/http/trip/create"
	httpfinish "github.com/Zvoook/TripGo_Danilov/internal/http/trip/finish"
	httpget "github.com/Zvoook/TripGo_Danilov/internal/http/trip/get"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/postgres"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/telemetry"
	"github.com/Zvoook/TripGo_Danilov/internal/repository/trip"
	create "github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/create"
	finish "github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/finish"
	get "github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/get"
)

type App struct {
	Logger            *slog.Logger
	Server            *http.Server
	pool              *pgxpool.Pool
	shutdownTelemetry func(context.Context) error
}

func (a *App) ShutdownTelemetry(ctx context.Context) error {
	return a.shutdownTelemetry(ctx)
}

func New(cfg config.Config, signalCtx context.Context) (*App, error) {
	logger := telemetry.NewLogger(cfg.Loglevel, os.Stdout)

	provider, err := telemetry.NewTracerProvider(signalCtx)
	if err != nil {
		return nil, err
	}

	meterProvider, err := telemetry.NewMeterProvider(signalCtx)
	if err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(signalCtx), cfg.ShutdownTimeout)
		defer cancel()
		return nil, errors.Join(err, provider.Shutdown(shutdownCtx))
	}
	shutdown := func(ctx context.Context) error {
		return errors.Join(provider.Shutdown(ctx), meterProvider.Shutdown(ctx))
	}
	app, err := newApp(cfg, signalCtx, logger, provider, meterProvider, shutdown)
	if err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(signalCtx), cfg.ShutdownTimeout)
		defer cancel()
		return nil, errors.Join(err, shutdown(shutdownCtx))
	}
	return app, nil
}

func newApp(
	cfg config.Config, ctx context.Context, logger *slog.Logger, provider trace.TracerProvider,
	meterProvider *sdkmetric.MeterProvider, shutdown func(context.Context) error,
) (*App, error) {
	meter := meterProvider.Meter("trip-service")
	router := chi.NewRouter()
	httpMetrics, err := middleware.Metrics(router, meter)
	if err != nil {
		return nil, err
	}
	created, err := newTripEvent(meter, logger, "tripgo.trips.created", "trip created")
	if err != nil {
		return nil, err
	}
	completed, err := newTripEvent(meter, logger, "tripgo.trips.completed", "trip completed")
	if err != nil {
		return nil, err
	}
	dbTracer, err := postgres.NewQueryTracer(provider, meter)
	if err != nil {
		return nil, err
	}
	pool, err := postgres.NewPool(ctx, cfg, dbTracer)
	if err != nil {
		return nil, fmt.Errorf("postgres connection: %w", err)
	}

	logger.InfoContext(ctx, "postgres connected")
	repository := trip.NewRepository(pool, cfg.DatabaseQueryTimeout)
	if err := observeActiveTrips(meter, repository.CountActive); err != nil {
		pool.Close()
		return nil, err
	}

	transactions := postgres.NewTxManager(pool, cfg.DatabaseQueryTimeout, logger)

	createUC := create.New(repository, transactions, cfg.OperationTimeout)
	getUC := get.New(repository, cfg.OperationTimeout)
	finishUC := finish.New(repository, transactions, cfg.OperationTimeout)

	getHandler := httpget.NewHandler(getUC)
	createHandler := httpcreate.NewHandler(observedCreate{run: createUC.Create, event: created})
	finishHandler := httpfinish.NewHandler(observedFinish{run: finishUC.Finish, event: completed})

	handler := httptransport.NewHandler(
		pool,
		cfg.DatabaseQueryTimeout,
		getHandler,
		createHandler,
		finishHandler,
	)

	router.Use(middleware.Recover(logger))
	router.Use(middleware.Tracing(provider))
	router.Use(middleware.Logging(logger))
	router.Use(httpMetrics)
	apiHandler := generated.HandlerWithOptions(
		handler,
		generated.ChiServerOptions{
			BaseRouter:       router,
			ErrorHandlerFunc: domainerror.HandleParameterError,
		},
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apiHandler,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	return &App{
		Logger:            logger,
		Server:            server,
		pool:              pool,
		shutdownTelemetry: shutdown,
	}, nil
}

func (a *App) CloseConnection() {
	a.pool.Close()
}

func (a *App) CloseServer() error {
	return a.Server.Close()
}
