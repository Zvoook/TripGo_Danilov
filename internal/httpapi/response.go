package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/trip"
)

func writeProblem(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	detail string,
) {
	var problem generated.Problem
	problem.Status = int32(status)
	problem.Code = code
	problem.Detail = &detail
	problem.Instance = &(r.URL.Path)
	problem.Type = "about:blank"
	problem.Title = http.StatusText(status)

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(problem)
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}

func HandleParameterError(w http.ResponseWriter, r *http.Request, err error) {
	writeProblem(w, r, 400, "invalid_request", "Invalid request parameter")
}

func toAPITrip(value trip.Trip) generated.Trip {
	return generated.Trip{
		Id:         value.ID,
		UserId:     value.UserID,
		DriverId:   value.DriverID,
		Price:      value.Price,
		Status:     generated.TripStatus(value.Status),
		StartedAt:  value.StartedAt,
		FinishedAt: value.FinishedAt,
		StartPoint: generated.Coordinates{
			Latitude:  value.StartLatitude,
			Longitude: value.StartLongitude,
		},
		EndPoint: generated.Coordinates{
			Latitude:  value.EndLatitude,
			Longitude: value.EndLongitude,
		},
	}
}
