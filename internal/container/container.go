package container

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Zvoook/TripGo_Danilov/internal/config"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	httptransport "github.com/Zvoook/TripGo_Danilov/internal/http"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	httpcreate "github.com/Zvoook/TripGo_Danilov/internal/http/trip/create"
	httpfinish "github.com/Zvoook/TripGo_Danilov/internal/http/trip/finish"
	httpget "github.com/Zvoook/TripGo_Danilov/internal/http/trip/get"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/postgres"
	"github.com/Zvoook/TripGo_Danilov/internal/repository/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/create"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/finish"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/get"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server *http.Server
	pool   *pgxpool.Pool
}

func New(cfg config.Config, signalCtx context.Context) (*App, error) {
	pool, err := postgres.NewPool(signalCtx, cfg)
	if err != nil {
		if err := signalCtx.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("postgres connection: %w", err)
	}
	fmt.Println("Successful connection")

	repository := trip.NewRepository(pool, cfg.DatabaseQueryTimeout)

	transactions := postgres.NewTxManager(pool, cfg.DatabaseQueryTimeout)

	createUC := create.New(repository, transactions, cfg.OperationTimeout)
	getUC := get.New(repository, cfg.OperationTimeout)
	finishUC := finish.New(repository, transactions, cfg.OperationTimeout)

	getHandler := httpget.NewHandler(getUC)
	createHandler := httpcreate.NewHandler(createUC)
	finishHandler := httpfinish.NewHandler(finishUC)

	handler := httptransport.NewHandler(
		pool,
		cfg.DatabaseQueryTimeout,
		getHandler,
		createHandler,
		finishHandler,
	)

	router := chi.NewRouter()
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
	}
	return &App{
		Server: server,
		pool:   pool,
	}, nil
}

func (a *App) CloseConnection() {
	a.pool.Close()
}

func (a *App) CloseServer() {
	a.Server.Close()
}
