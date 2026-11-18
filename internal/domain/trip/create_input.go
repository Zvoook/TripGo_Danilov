package trip

import "github.com/google/uuid"

// Объяснение для проверяющего зачем необходим отдельный CreateInput
// Внутри Trip есть поля, которые не должны быть установлены при создании новой поездки.
// Это, например, ID поездки, статус, время создания и обновления.
// Поэтому создается отдельную структура CreateInput, которая содержит только те поля,
// которые необходимы для создания новой поездки.
// Именно их присылает клиент в запросе на создание новой поездки.
// Данные поля мы как раз и проверяем в методе Validate.

type CreateInput struct {
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLongitude float64
	StartLatitude  float64
	EndLongitude   float64
	EndLatitude    float64
	Price          int64
}

func (input CreateInput) Validate() error {
	if input.UserID == uuid.Nil {
		return ValidationError{
			Field:   "user_id",
			Message: "must not be empty",
		}
	}

	if input.DriverID == uuid.Nil {
		return ValidationError{
			Field:   "driver_id",
			Message: "must not be empty",
		}
	}

	if !validCoordinate(input.StartLatitude, 90) {
		return ValidationError{
			Field:   "start_point.latitude",
			Message: "must be between -90 and 90",
		}
	}

	if !validCoordinate(input.StartLongitude, 180) {
		return ValidationError{
			Field:   "start_point.longitude",
			Message: "must be between -180 and 180",
		}
	}

	if !validCoordinate(input.EndLatitude, 90) {
		return ValidationError{
			Field:   "end_point.latitude",
			Message: "must be between -90 and 90",
		}
	}

	if !validCoordinate(input.EndLongitude, 180) {
		return ValidationError{
			Field:   "end_point.longitude",
			Message: "must be between -180 and 180",
		}
	}

	if input.Price < 0 {
		return ValidationError{
			Field:   "price",
			Message: "must be nonnegative",
		}
	}
	return nil
}
