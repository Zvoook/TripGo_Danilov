package trip

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Position struct {
	ID         int64
	TripID     uuid.UUID
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
}

func (p *Position) Validate(now time.Time) error {
	if p.TripID == uuid.Nil {
		return fmt.Errorf("%w: trip_id must not be empty", ErrInvalidRequest)
	}
	if !validCoordinate(p.Latitude, 90) {
		return fmt.Errorf("%w: latitude must be in range [-90, 90]", ErrInvalidRequest)
	}
	if !validCoordinate(p.Longitude, 180) {
		return fmt.Errorf("%w: longitude must be in range [-180, 180]", ErrInvalidRequest)
	}
	if p.RecordedAt.IsZero() || p.RecordedAt.After(now.Add(time.Minute)) {
		return fmt.Errorf("%w: recorded_at must not be zero or in the future", ErrInvalidRequest)
	}
	return nil
}
