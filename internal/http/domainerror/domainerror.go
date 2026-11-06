package domainerror

import (
	"errors"
	"net/http"
	"strings"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	api "github.com/Zvoook/TripGo_Danilov/internal/generated"
)

func Problem(err error) api.Problem {
	problem := api.Problem{}
	detail := err.Error()
	switch {
	case errors.Is(err, tripdomain.ErrTripNotFound):
		problem = api.Problem{
			Title:  "Trip not found",
			Status: http.StatusNotFound,
			Code:   "trip_not_found",
		}
	case errors.Is(err, tripdomain.ErrTripCompleted):
		problem = api.Problem{
			Title:  "Trip completed",
			Status: http.StatusConflict,
			Code:   "trip_completed",
		}
	case errors.Is(err, tripdomain.ErrDriverBusy):
		problem = api.Problem{
			Title:  "Driver busy",
			Status: http.StatusConflict,
			Code:   "driver_busy",
		}
	case errors.Is(err, tripdomain.ErrInvalidRequest):
		problem = api.Problem{
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Code:   "invalid_request",
		}
	default:
		problem = api.Problem{
			Title:  "Internal server error",
			Status: http.StatusInternalServerError,
			Code:   "internal_error",
		}
		detail = "Internal server error"
	}
	problem.Type = "https://tripgo.example/problems/" + strings.ReplaceAll(problem.Code, "_", "-")
	problem.Detail = &detail
	return problem
}
