package trip

import "github.com/google/uuid"

type CreateInput struct {
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLongitude float64
	StartLatitude  float64
	EndLongitude   float64
	EndLatitude    float64
	Price          int64
}
