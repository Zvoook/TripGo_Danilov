package create_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	api "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/http/trip/create"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

const validTripJSON = `{
	"user_id": "123e4567-e89b-12d3-a456-426614174001",
	"driver_id": "123e4567-e89b-12d3-a456-426614174002",
	"start_point": {"latitude": 30, "longitude": 10},
	"end_point": {"latitude": 60, "longitude": 45},
	"price": 1000
}`

func expectedCreateInput() tripdomain.CreateInput {
	return tripdomain.CreateInput{
		UserID:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		DriverID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		StartLatitude:  30,
		StartLongitude: 10,
		EndLatitude:    60,
		EndLongitude:   45,
		Price:          1000,
	}
}

func newCreateRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestCreateTrip_ValidJSON(t *testing.T) {
	input := expectedCreateInput()
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	usecase := NewMockCreator(gomock.NewController(t))
	usecase.EXPECT().Create(gomock.Any(), input).Return(tripdomain.Trip{
		ID:             id,
		UserID:         input.UserID,
		DriverID:       input.DriverID,
		Status:         tripdomain.StatusActive,
		Price:          input.Price,
		StartLatitude:  input.StartLatitude,
		StartLongitude: input.StartLongitude,
		EndLatitude:    input.EndLatitude,
		EndLongitude:   input.EndLongitude,
	}, nil).Times(1)

	rec := httptest.NewRecorder()
	create.NewHandler(usecase).CreateTrip(rec, newCreateRequest(validTripJSON), api.CreateTripParams{})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if got, want := rec.Header().Get("Location"), "/api/v1/trips/"+id.String(); got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}

	var got api.Trip
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Id != id || got.UserId != input.UserID || got.DriverId != input.DriverID {
		t.Errorf("unexpected response IDs: trip=%v user=%v driver=%v", got.Id, got.UserId, got.DriverId)
	}
	if string(got.Status) != "active" || got.Price != input.Price {
		t.Errorf("status=%q price=%d, want active and %d", got.Status, got.Price, input.Price)
	}
	if got.StartPoint.Latitude != input.StartLatitude || got.StartPoint.Longitude != input.StartLongitude {
		t.Errorf("unexpected start point: %+v", got.StartPoint)
	}
	if got.EndPoint.Latitude != input.EndLatitude || got.EndPoint.Longitude != input.EndLongitude {
		t.Errorf("unexpected end point: %+v", got.EndPoint)
	}
}

func TestCreateTrip_InvalidJSON(t *testing.T) {
	usecase := NewMockCreator(gomock.NewController(t))

	rec := httptest.NewRecorder()
	create.NewHandler(usecase).CreateTrip(rec, newCreateRequest("{"), api.CreateTripParams{})

	checkCreateProblem(t, rec, http.StatusBadRequest, "invalid_request")
}

func TestCreateTrip_Conflict(t *testing.T) {
	usecase := NewMockCreator(gomock.NewController(t))
	usecase.EXPECT().Create(gomock.Any(), expectedCreateInput()).
		Return(tripdomain.Trip{}, tripdomain.ErrDriverBusy).Times(1)

	rec := httptest.NewRecorder()
	create.NewHandler(usecase).CreateTrip(rec, newCreateRequest(validTripJSON), api.CreateTripParams{})

	checkCreateProblem(t, rec, http.StatusConflict, "driver_busy")
}

func checkCreateProblem(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, wantStatus, rec.Body.String())
	}
	var problem api.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if problem.Code != wantCode {
		t.Errorf("problem code = %q, want %q", problem.Code, wantCode)
	}
}
