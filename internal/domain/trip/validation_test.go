package trip

import (
	"errors"
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestCreateInputValidate(t *testing.T) {
	valid := CreateInput{
		UserID: uuid.New(), DriverID: uuid.New(),
		StartLatitude: -90, StartLongitude: 180,
		EndLatitude: 90, EndLongitude: -180,
	}
	cases := []struct {
		name   string
		change func(*CreateInput)
		want   error
	}{
		{"boundary coordinates and zero price", func(*CreateInput) {}, nil},
		{"missing user", func(in *CreateInput) { in.UserID = uuid.Nil }, ErrInvalidRequest},
		{"missing driver", func(in *CreateInput) { in.DriverID = uuid.Nil }, ErrInvalidRequest},
		{"start latitude", func(in *CreateInput) { in.StartLatitude = -91 }, ErrInvalidRequest},
		{"start longitude", func(in *CreateInput) { in.StartLongitude = 181 }, ErrInvalidRequest},
		{"end latitude", func(in *CreateInput) { in.EndLatitude = 91 }, ErrInvalidRequest},
		{"end longitude", func(in *CreateInput) { in.EndLongitude = -181 }, ErrInvalidRequest},
		{"NaN coordinate", func(in *CreateInput) { in.StartLatitude = math.NaN() }, ErrInvalidRequest},
		{"infinite coordinate", func(in *CreateInput) { in.EndLongitude = math.Inf(1) }, ErrInvalidRequest},
		{"negative price", func(in *CreateInput) { in.Price = -1 }, ErrInvalidRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			if err := input.Validate(); !errors.Is(err, tc.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tc.want)
			}
		})
	}
}
