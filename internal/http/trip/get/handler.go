package get

import (
	"encoding/json"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	convert "github.com/Zvoook/TripGo_Danilov/internal/http/convert"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

type Handler struct {
	tripGetter tripGetter
}

func NewHandler(tripGetter tripGetter) *Handler {
	return &Handler{
		tripGetter: tripGetter,
	}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId generated.TripId) {
	var response generated.Trip

	result, err := h.tripGetter.Get(r.Context(), tripId)
	w.Header().Set("Content-type", "application/json")

	if err != nil {
		problem := domainerror.Problem(err)
		if problem.Status >= http.StatusInternalServerError {
			requestlog.Error(r.Context(), "trip operation failed", err)
		}
		domainerror.WriteProblem(w, r, problem)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		response = convert.ToAPITrip(result)
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		requestlog.Error(r.Context(), "json encoding failed", err)
	}
}
