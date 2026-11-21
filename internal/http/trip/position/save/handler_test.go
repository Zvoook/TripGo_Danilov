package save

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	api "github.com/Zvoook/TripGo_Danilov/internal/generated"
)

type saveFunc func(context.Context, tripdomain.Position) (tripdomain.Position, error)

func (f saveFunc) Save(ctx context.Context, p tripdomain.Position) (tripdomain.Position, error) {
	return f(ctx, p)
}

func TestCreateTripPosition(t *testing.T) {
	tripID := uuid.New()
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{name: "saved", want: http.StatusCreated},
		{name: "missing", err: tripdomain.ErrTripNotFound, want: http.StatusNotFound},
		{name: "completed", err: tripdomain.ErrTripCompleted, want: http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(saveFunc(func(_ context.Context, position tripdomain.Position) (tripdomain.Position, error) {
				if position.TripID != tripID || position.Latitude != 59.93 || position.Longitude != 30.33 || position.RecordedAt.IsZero() {
					t.Fatalf("unexpected position: %+v", position)
				}
				position.ID = 42
				return position, tc.err
			}))
			r := httptest.NewRequest(http.MethodPost, "/positions", strings.NewReader(`{"coordinates":{"latitude":59.93,"longitude":30.33},"recorded_at":"2025-01-01T00:00:00Z"}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.CreateTripPosition(w, r, tripID)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.want, w.Body)
			}
			if tc.err == nil {
				var response api.TripPosition
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Id != 42 || response.TripId != tripID {
					t.Fatalf("unexpected response: %s, decode error: %v", w.Body, err)
				}
			}
		})
	}
}
