package get

import (
	"encoding/json"
	"log"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	convert "github.com/Zvoook/TripGo_Danilov/internal/http/convert"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
)

type Handler struct {
	getUsecase getter
}

func NewHandler(getUsecase getter) *Handler {
	return &Handler{
		getUsecase: getUsecase,
	}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId generated.TripId) {
	var response generated.Trip

	result, err := h.getUsecase.Get(r.Context(), tripId)
	w.Header().Set("Content-type", "application/json")

	if err != nil {
		log.Printf("get trip failed: %v", err)
		problem := domainerror.Problem(err)
		domainerror.WriteProblem(w, r, problem)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		response = convert.ToAPITrip(result)
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("json encoding failed: %v", err)
	}
}
