package trip

import "errors"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

var (
	ErrTripNotFound   = errors.New("trip not found")
	ErrTripCompleted  = errors.New("trip already completed")
	ErrDriverBusy     = errors.New("driver already has an active trip")
	ErrInvalidRequest = errors.New("invalid trip request")
)
