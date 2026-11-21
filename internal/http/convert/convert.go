package convert

import (
	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	generated "github.com/Zvoook/TripGo_Danilov/internal/generated"
)

func ToAPITrip(value tripdomain.Trip) generated.Trip {
	return generated.Trip{
		Id:         value.ID,
		UserId:     value.UserID,
		DriverId:   value.DriverID,
		Price:      value.Price,
		Status:     generated.TripStatus(value.Status),
		StartedAt:  value.StartedAt,
		FinishedAt: value.FinishedAt,
		StartPoint: generated.Coordinates{
			Latitude:  value.StartLatitude,
			Longitude: value.StartLongitude,
		},
		EndPoint: generated.Coordinates{
			Latitude:  value.EndLatitude,
			Longitude: value.EndLongitude,
		},
	}
}
