package trip

import (
	"errors"
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestCreateInputValidate(t *testing.T) {
	const latitudeRange = "must be between -90 and 90"
	const longitudeRange = "must be between -180 and 180"
	const priceRange = "must be nonnegative"
	const idEmpty = "must not be empty"
	valid := CreateInput{
		UserID: uuid.New(), DriverID: uuid.New(),
		StartLatitude: -90, StartLongitude: 180,
		EndLatitude: 90, EndLongitude: -180,
		Price: 10,
	}
	cases := []struct {
		name    string
		change  func(*CreateInput)
		field   string
		message string
	}{
		{"boundary coordinates and zero price", func(*CreateInput) {}, "", ""},
		{"missing user", func(in *CreateInput) { in.UserID = uuid.Nil }, "user_id", idEmpty},
		{"missing driver", func(in *CreateInput) { in.DriverID = uuid.Nil }, "driver_id", idEmpty},
		{"start latitude", func(in *CreateInput) { in.StartLatitude = -91 }, "start_point.latitude", latitudeRange},
		{"start longitude", func(in *CreateInput) { in.StartLongitude = 181 }, "start_point.longitude", longitudeRange},
		{"end latitude", func(in *CreateInput) { in.EndLatitude = 91 }, "end_point.latitude", latitudeRange},
		{"end longitude", func(in *CreateInput) { in.EndLongitude = -181 }, "end_point.longitude", longitudeRange},
		{"NaN coordinate", func(in *CreateInput) { in.StartLatitude = math.NaN() }, "start_point.latitude", latitudeRange},
		{"infinite coordinate", func(in *CreateInput) { in.EndLongitude = math.Inf(1) }, "end_point.longitude", longitudeRange},
		{"negative price", func(in *CreateInput) { in.Price = -1 }, "price", priceRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			err := input.Validate()
			if tc.field == "" {
				if err != nil {
					t.Fatalf("expected valid input, got %v", err)
				}
				return
			}
			var validationErr ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if validationErr.Field != tc.field {
				t.Errorf("field = %q, want %q", validationErr.Field, tc.field)
			}
			if validationErr.Message != tc.message {
				t.Errorf("message = %q, want %q", validationErr.Message, tc.message)
			}
			if tc.field == "" {
				if err != nil {
					t.Fatalf("expected valid input, got %v", err)
				}
				return
			}
		})
	}
}
