package get_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
	transport "github.com/Zvoook/TripGo_Danilov/internal/http"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
	get "github.com/Zvoook/TripGo_Danilov/internal/http/trip/get"
)

func TestGetTrip_NotFound(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctrl := gomock.NewController(t)
	usecase := NewMockGetter(ctrl)
	usecase.EXPECT().
		Get(gomock.Any(), id).
		Return(tripdomain.Trip{}, tripdomain.ErrTripNotFound).
		Times(1)

	handler := get.NewHandler(usecase)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetTrip(rec, req, id)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}

	var problem generated.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if problem.Code != "trip_not_found" {
		t.Errorf("problem code = %q, want trip_not_found", problem.Code)
	}
}

func TestGetTrip_OK(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctrl := gomock.NewController(t)
	usecase := NewMockGetter(ctrl)

	foundTrip := tripdomain.Trip{
		ID:             id,
		Status:         tripdomain.StatusActive,
		Price:          1000,
		StartLatitude:  10.00,
		StartLongitude: 30.00,
	}

	usecase.EXPECT().
		Get(gomock.Any(), id).
		Return(foundTrip, nil).
		Times(1)

	handler := get.NewHandler(usecase)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetTrip(rec, req, id)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got generated.Trip
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Id != foundTrip.ID {
		t.Errorf("trip ID = %q, want %q", got.Id, id)
	}
	if string(got.Status) != "active" {
		t.Errorf("trip Status = %q, want %q", got.Status, "active")
	}
	if got.Price != foundTrip.Price {
		t.Errorf("trip Price = %d, want %d", got.Price, foundTrip.Price)
	}
	if got.StartPoint.Latitude != foundTrip.StartLatitude {
		t.Errorf("trip StartLatitude = %f, want %f", got.StartPoint.Latitude, foundTrip.StartLatitude)
	}
	if got.StartPoint.Longitude != foundTrip.StartLongitude {
		t.Errorf("trip StartLongitude = %f, want %f", got.StartPoint.Longitude, foundTrip.StartLongitude)
	}
}

func TestGetTrip_InternalError(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ctrl := gomock.NewController(t)
	usecase := NewMockGetter(ctrl)

	const privateMessage = "db-private-marker"

	usecase.EXPECT().
		Get(gomock.Any(), id).
		Return(tripdomain.Trip{}, errors.New(privateMessage)).
		Times(1)

	handler := get.NewHandler(usecase)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetTrip(rec, req, id)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	var problem generated.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if problem.Code != "internal_error" {
		t.Errorf("problem code = %q, want internal_error", problem.Code)
	}
	if strings.Contains(rec.Body.String(), privateMessage) {
		t.Errorf("response leaked internal error: %s", rec.Body.String())
	}
}

func TestGetTripThroughRouter(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	usecase := NewMockGetter(gomock.NewController(t))
	usecase.EXPECT().Get(gomock.Any(), id).
		Return(tripdomain.Trip{ID: id, Status: tripdomain.StatusActive}, nil)

	handler := transport.NewHandler(nil, time.Second, get.NewHandler(usecase), nil, nil)
	router := generated.HandlerWithOptions(handler, generated.ChiServerOptions{
		ErrorHandlerFunc: domainerror.HandleParameterError,
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body)
	}
}
