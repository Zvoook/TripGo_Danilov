package httpapi

import (
	"encoding/json"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
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
		//Log.Write()
	}
}

func HandleParameterError(w http.ResponseWriter, r *http.Request, err error) {
	writeProblem(w, r, 400, "invalid_request", "Invalid request parameter")
}
