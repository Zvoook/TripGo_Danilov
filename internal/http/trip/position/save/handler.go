package save

import (
	"encoding/json"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

type Handler struct {
	saveUsecase saver
}

func NewHandler(saveUC saver) *Handler {
	return &Handler{
		saveUsecase: saveUC,
	}
}

func (h *Handler) CreateTripPosition(
	w http.ResponseWriter,
	r *http.Request,
	tripID generated.TripId,
) {
	position, err := readPosition(w, r, tripID)
	if err != nil {
		domainerror.WriteProblem(w, r, domainerror.Problem(err))
		return
	}

	saved, err := h.saveUsecase.Save(r.Context(), position)
	if err != nil {
		problem := domainerror.Problem(err)
		if problem.Status >= http.StatusInternalServerError {
			requestlog.Error(r.Context(), "failed to save trip position", err)
		}
		domainerror.WriteProblem(w, r, problem)
		return
	}

	response := generated.TripPosition{
		Id:         saved.ID,
		TripId:     saved.TripID,
		RecordedAt: saved.RecordedAt,
		Coordinates: generated.Coordinates{
			Latitude:  saved.Latitude,
			Longitude: saved.Longitude,
		},
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		requestlog.Error(r.Context(), "json encoding failed", err)
	}
}
