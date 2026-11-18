package trip

import "github.com/google/uuid"

type StaleTrip struct {
	ID       uuid.UUID
	DriverID uuid.UUID
}
