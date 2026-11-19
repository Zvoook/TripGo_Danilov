package save

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
)

func readPosition(
	w http.ResponseWriter,
	r *http.Request,
	tripID generated.TripId,
) (tripdomain.Position, error) {
	var input tripdomain.Position
	var request generated.CreateTripPositionJSONRequestBody

	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return input, tripdomain.ErrInvalidRequest
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return input, tripdomain.ErrInvalidRequest
	}

	fields, err := requiredFields(
		body,
		"coordinates", "recorded_at",
	)
	if err != nil {
		return input, tripdomain.ErrInvalidRequest
	}

	_, err = requiredFields(fields["coordinates"], "latitude", "longitude")
	if err != nil {
		return input, tripdomain.ErrInvalidRequest
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&request)
	if err != nil {
		return input, tripdomain.ErrInvalidRequest
	}

	input.TripID = tripID
	input.Latitude = request.Coordinates.Latitude
	input.Longitude = request.Coordinates.Longitude
	input.RecordedAt = request.RecordedAt

	return input, nil
}

func requiredFields(
	body []byte,
	names ...string,
) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage

	err := json.Unmarshal(body, &fields)
	if err != nil || fields == nil {
		return nil, fmt.Errorf("expected a JSON object")
	}

	for _, name := range names {
		value, exists := fields[name]

		if !exists {
			return nil, fmt.Errorf("%s is required", name)
		}

		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("%s must not be null", name)
		}
	}

	if len(fields) != len(names) {
		return nil, fmt.Errorf("unknown fields are not allowed")
	}

	return fields, nil
}
