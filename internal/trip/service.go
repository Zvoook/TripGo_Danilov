package trip

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type TxManager interface {
	Do(context.Context, func(context.Context) error) error
}

type Service struct {
	repository       *Repository
	transactions     TxManager
	operationTimeout time.Duration
}

func NewService(repository *Repository, transactions TxManager, timeout time.Duration) *Service {
	return &Service{
		repository:       repository,
		transactions:     transactions,
		operationTimeout: timeout,
	}
}

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
		return fmt.Errorf("%w: user_id must not be empty", ErrInvalidInput)
	}
	if input.DriverID == uuid.Nil {
		return fmt.Errorf("%w: driver_id must not be empty", ErrInvalidInput)
	}
	if !validCoordinate(input.StartLatitude, 90) {
		return fmt.Errorf("%w: invalid start latitude", ErrInvalidInput)
	}
	if !validCoordinate(input.StartLongitude, 180) {
		return fmt.Errorf("%w: invalid start longitude", ErrInvalidInput)
	}
	if !validCoordinate(input.EndLatitude, 90) {
		return fmt.Errorf("%w: invalid end latitude", ErrInvalidInput)
	}
	if !validCoordinate(input.EndLongitude, 180) {
		return fmt.Errorf("%w: invalid end longitude", ErrInvalidInput)
	}
	if input.Price < 0 {
		return fmt.Errorf("%w: price must be nonnegative", ErrInvalidInput)
	}
	return nil
}

func validCoordinate(value float64, limit float64) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return value >= -limit && value <= limit
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Trip, error) {
	err := input.Validate()
	if err != nil {
		return Trip{}, err
	}

	operationCtx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	// PostgreSQL stores timestamps with microsecond precision.
	now := time.Now().UTC().Truncate(time.Microsecond)
	value := Trip{
		ID:             uuid.New(),
		UserID:         input.UserID,
		DriverID:       input.DriverID,
		StartLongitude: input.StartLongitude,
		StartLatitude:  input.StartLatitude,
		EndLongitude:   input.EndLongitude,
		EndLatitude:    input.EndLatitude,
		Price:          input.Price,
		Status:         StatusActive,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.transactions.Do(operationCtx, func(txCtx context.Context) error {
		err := s.repository.Insert(txCtx, value)
		if err != nil {
			return err
		}

		err = s.repository.AddHistory(txCtx, value.ID, nil, StatusActive, now)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return Trip{}, err
	}
	return value, nil
}

func (s *Service) Finish(ctx context.Context, id uuid.UUID) (Trip, error) {
	operationCtx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	var result Trip
	err := s.transactions.Do(operationCtx, func(txCtx context.Context) error {
		value, err := s.repository.GetForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		if value.Status == StatusCompleted {
			return ErrTripCompleted
		}

		now := time.Now().UTC().Truncate(time.Microsecond)
		err = s.repository.Complete(txCtx, id, now)
		if err != nil {
			return err
		}

		previous := value.Status
		err = s.repository.AddHistory(txCtx, id, &previous, StatusCompleted, now)
		if err != nil {
			return err
		}

		value.Status = StatusCompleted
		value.FinishedAt = &now
		value.UpdatedAt = now
		result = value
		return nil
	})
	if err != nil {
		return Trip{}, err
	}
	return result, nil
}
