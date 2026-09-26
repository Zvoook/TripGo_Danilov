package httpapi

import (
	"context"
	"encoding/json"
	"errors"
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
	generated.Unimplemented
}

func NewHandler(pool *pgxpool.Pool, timeout time.Duration, repos *trip.Repository) *Handler {
	return &Handler{
		pool:         pool,
		queryTimeout: timeout,
		trips:        repos,
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

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId generated.TripId) {
	var response generated.Trip

	result, err := h.trips.GetByID(r.Context(), tripId)
	w.Header().Set("Content-type", "application/json")

	if errors.Is(err, trip.ErrNotFound) {
		writeProblem(w, r, http.StatusNotFound, "trip_not_found", "No trips was found")
		return
	} else if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Server arise the issue")
		return
	} else {
		w.WriteHeader(http.StatusOK)

		response = generated.Trip{
			Id:         result.ID,
			UserId:     result.UserID,
			DriverId:   result.DriverID,
			Price:      result.Price,
			StartedAt:  result.StartedAt,
			FinishedAt: result.FinishedAt,
			Status:     generated.TripStatus(result.Status),
			StartPoint: generated.Coordinates{
				Latitude:  result.StartLatitude,
				Longitude: result.StartLongitude,
			},
			EndPoint: generated.Coordinates{
				Latitude:  result.EndLatitude,
				Longitude: result.EndLongitude,
			},
			LastPositionAt: nil,
		}
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		//Log.Write()
	}
}
