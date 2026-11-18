package trip

import "errors"

var (
	ErrTripNotFound   = errors.New("trip not found")
	ErrTripCompleted  = errors.New("trip already completed")
	ErrDriverBusy     = errors.New("driver already has an active trip")
	ErrInvalidRequest = errors.New("invalid trip request")
)

// ValidationError указывает поле с неверным значением и причину ошибки.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
