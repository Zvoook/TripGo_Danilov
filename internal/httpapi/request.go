package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/trip"
)

func readCreateTrip(
	w http.ResponseWriter,
	r *http.Request,
) (trip.CreateInput, error) {
	var input trip.CreateInput
	var request generated.CreateTripJSONRequestBody

	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return input, fmt.Errorf("Content-Type must be application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return input, fmt.Errorf("cannot read request body")
	}

	fields, err := requiredFields(
		body,
		"user_id", "driver_id", "start_point", "end_point", "price",
	)
	if err != nil {
		return input, err
	}

	_, err = requiredFields(fields["start_point"], "latitude", "longitude")
	if err != nil {
		return input, fmt.Errorf("start_point: %w", err)
	}

	_, err = requiredFields(fields["end_point"], "latitude", "longitude")
	if err != nil {
		return input, fmt.Errorf("end_point: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&request)
	if err != nil {
		return input, fmt.Errorf("invalid request fields")
	}

	input.UserID = request.UserId
	input.DriverID = request.DriverId
	input.Price = request.Price
	input.StartLatitude = request.StartPoint.Latitude
	input.StartLongitude = request.StartPoint.Longitude
	input.EndLatitude = request.EndPoint.Latitude
	input.EndLongitude = request.EndPoint.Longitude

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
