package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewHandler(pool *pgxpool.Pool, timeout time.Duration) *Handler {
	return &Handler{
		pool:         pool,
		queryTimeout: timeout,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	var response generated.HealthResponse
	response.Status = generated.Ok
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		//Log.Write()
	}
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, close := context.WithTimeout(r.Context(), h.queryTimeout)
	defer close()
	var response generated.HealthResponse

	w.Header().Set("Content-type", "application/json")
	if err := h.pool.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		response.Status = generated.Unavailable
	} else {
		w.WriteHeader(http.StatusOK)
		response.Status = generated.Ok
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		//Log.Write()
	}
}

func (h *Handler) Waiting(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 15)
	var response generated.HealthResponse

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	response.Status = generated.Ok

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		//Log.Write()
	}
}
