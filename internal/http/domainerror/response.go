package domainerror

import (
	"encoding/json"
	"net/http"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

func WriteProblem(
	w http.ResponseWriter,
	r *http.Request,
	problem generated.Problem,
) {
	problem.Instance = &(r.URL.Path)

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(int(problem.Status))
	err := json.NewEncoder(w).Encode(problem)
	if err != nil {
		requestlog.Error(r.Context(), "json encoding failed", err)
	}
}

func HandleParameterError(w http.ResponseWriter, r *http.Request, _ error) {
	problem := Problem(tripdomain.ErrInvalidRequest)
	WriteProblem(w, r, problem)
}
