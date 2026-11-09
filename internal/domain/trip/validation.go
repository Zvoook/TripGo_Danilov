package trip

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

func validCoordinate(value, limit float64) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return value >= -limit && value <= limit
}

func (input CreateInput) Validate() error {
	if input.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id must not be empty", ErrInvalidRequest)
	}
	if input.DriverID == uuid.Nil {
		return fmt.Errorf("%w: driver_id must not be empty", ErrInvalidRequest)
	}
	if !validCoordinate(input.StartLatitude, 90) {
		return fmt.Errorf("%w: invalid start latitude", ErrInvalidRequest)
	}
	if !validCoordinate(input.StartLongitude, 180) {
		return fmt.Errorf("%w: invalid start longitude", ErrInvalidRequest)
	}
	if !validCoordinate(input.EndLatitude, 90) {
		return fmt.Errorf("%w: invalid end latitude", ErrInvalidRequest)
	}
	if !validCoordinate(input.EndLongitude, 180) {
		return fmt.Errorf("%w: invalid end longitude", ErrInvalidRequest)
	}
	if input.Price < 0 {
		return fmt.Errorf("%w: price must be nonnegative", ErrInvalidRequest)
	}
	return nil
}
