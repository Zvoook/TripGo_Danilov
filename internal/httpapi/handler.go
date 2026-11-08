package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/trip"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	trips        *trip.Repository
	service      *trip.Service
}

func NewHandler(
	pool *pgxpool.Pool,
	timeout time.Duration,
	repos *trip.Repository,
	service *trip.Service,
) *Handler {
	return &Handler{
		pool:         pool,
		queryTimeout: timeout,
		trips:        repos,
		service:      service,
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
	var response generated.Trip

	result, err := h.trips.GetByID(r.Context(), tripId)
	w.Header().Set("Content-type", "application/json")

	if errors.Is(err, trip.ErrNotFound) {
		writeProblem(w, r, http.StatusNotFound, "trip_not_found", "No trips was found")
		return
	} else if err != nil {
		log.Printf("get trip failed: %v", err)
		writeProblem(
			w, r,
			http.StatusInternalServerError,
			"internal_error",
			"Internal server error",
		)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		response = toAPITrip(result)
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}

func (h *Handler) CreateTrip(
	w http.ResponseWriter,
	r *http.Request,
	params generated.CreateTripParams,
) {
	input, err := readCreateTrip(w, r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result, err := h.service.Create(r.Context(), input)

	if errors.Is(err, trip.ErrInvalidInput) {
		writeProblem(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if errors.Is(err, trip.ErrDriverBusy) {
		writeProblem(
			w, r,
			http.StatusConflict,
			"driver_busy",
			"Driver already has an active trip",
		)
		return
	}

	if err != nil {
		log.Printf("create trip failed: %v", err)
		writeProblem(
			w, r,
			http.StatusInternalServerError,
			"internal_error",
			"Internal server error",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/v1/trips/"+result.ID.String())
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(toAPITrip(result))
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}

func (h *Handler) FinishTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripId generated.TripId,
) {
	result, err := h.service.Finish(r.Context(), tripId)

	if errors.Is(err, trip.ErrNotFound) {
		writeProblem(
			w, r,
			http.StatusNotFound,
			"trip_not_found",
			"Trip not found",
		)
		return
	}

	if errors.Is(err, trip.ErrTripCompleted) {
		writeProblem(
			w, r,
			http.StatusConflict,
			"trip_completed",
			"Trip is already completed",
		)
		return
	}

	if err != nil {
		log.Printf("finish trip failed: %v", err)
		writeProblem(
			w, r,
			http.StatusInternalServerError,
			"internal_error",
			"Internal server error",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(toAPITrip(result))
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}
