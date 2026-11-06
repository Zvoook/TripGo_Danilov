package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	create "github.com/Zvoook/TripGo_Danilov/internal/http/trip/create"
	finish "github.com/Zvoook/TripGo_Danilov/internal/http/trip/finish"
	get "github.com/Zvoook/TripGo_Danilov/internal/http/trip/get"
)

type Handler struct {
	pool          readinessChecker
	queryTimeout  time.Duration
	getUsecase    *get.Handler
	createUsecase *create.Handler
	finishUsecase *finish.Handler
}

func NewHandler(
	pool readinessChecker,
	timeout time.Duration,
	getUC *get.Handler,
	createUC *create.Handler,
	finishUC *finish.Handler,
) *Handler {
	return &Handler{
		pool:          pool,
		queryTimeout:  timeout,
		getUsecase:    getUC,
		createUsecase: createUC,
		finishUsecase: finishUC,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	var response generated.HealthResponse
	response.Status = generated.Ok
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("json encoding failed: %v", err)
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
		log.Printf("json encoding failed: %v", err)
	}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId generated.TripId) {
	h.getUsecase.GetTrip(w, r, tripId)
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params generated.CreateTripParams) {
	h.createUsecase.CreateTrip(w, r, params)
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId generated.TripId) {
	h.finishUsecase.FinishTrip(w, r, tripId)
}
