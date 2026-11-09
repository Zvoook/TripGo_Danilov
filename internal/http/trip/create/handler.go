package create

import (
	"encoding/json"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	convert "github.com/Zvoook/TripGo_Danilov/internal/http/convert"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

type Handler struct {
	createUsecase creator
}

func NewHandler(createUsecase creator) *Handler {
	return &Handler{
		createUsecase: createUsecase,
	}
}

func (h *Handler) CreateTrip(
	w http.ResponseWriter,
	r *http.Request,
	_ generated.CreateTripParams,
) {
	input, err := readCreateTrip(w, r)
	if err != nil {
		problem := domainerror.Problem(err)
		if problem.Status >= http.StatusInternalServerError {
			requestlog.Error(r.Context(), "trip operation failed", err)
		}
		domainerror.WriteProblem(w, r, problem)
		return
	}

	result, err := h.createUsecase.Create(r.Context(), input)
	if err != nil {
		problem := domainerror.Problem(err)
		if problem.Status >= http.StatusInternalServerError {
			requestlog.Error(r.Context(), "trip operation failed", err)
		}
		domainerror.WriteProblem(w, r, problem)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/v1/trips/"+result.ID.String())
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(convert.ToAPITrip(result))
	if err != nil {
		requestlog.Error(r.Context(), "json encoding failed", err)
	}
}
