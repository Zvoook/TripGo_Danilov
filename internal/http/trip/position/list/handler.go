package list

import (
	"encoding/json"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

type Handler struct {
	usecase lister
}

func NewHandler(usecase lister) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) ListTripPositions(
	w http.ResponseWriter,
	r *http.Request,
	tripID generated.TripId,
) {
	positions, err := h.usecase.List(r.Context(), tripID)
	if err != nil {
		problem := domainerror.Problem(err)
		if problem.Status >= http.StatusInternalServerError {
			requestlog.Error(r.Context(), "list trip positions failed", err)
		}
		domainerror.WriteProblem(w, r, problem)
		return
	}

	response := generated.TripPositionsResponse{
		Positions: make([]generated.TripPosition, 0, len(positions)),
	}
	for _, position := range positions {
		response.Positions = append(response.Positions, generated.TripPosition{
			Id:         position.ID,
			TripId:     position.TripID,
			RecordedAt: position.RecordedAt,
			Coordinates: generated.Coordinates{
				Latitude:  position.Latitude,
				Longitude: position.Longitude,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		requestlog.Error(r.Context(), "json encoding failed", err)
	}
}
