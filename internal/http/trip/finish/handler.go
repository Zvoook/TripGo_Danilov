package finish

import (
	"encoding/json"
	"log"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	convert "github.com/Zvoook/TripGo_Danilov/internal/http/convert"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
)

type Handler struct {
	finishUsecase finisher
}

func NewHandler(finishUsecase finisher) *Handler {
	return &Handler{
		finishUsecase: finishUsecase,
	}
}

func (h *Handler) FinishTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripId generated.TripId,
) {
	result, err := h.finishUsecase.Finish(r.Context(), tripId)

	if err != nil {
		log.Printf("finish trip failed: %v", err)
		problem := domainerror.Problem(err)
		domainerror.WriteProblem(w, r, problem)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(convert.ToAPITrip(result))
	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}
