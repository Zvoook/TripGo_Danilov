package trip

import (
	"math"
)

func validCoordinate(value, limit float64) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return value >= -limit && value <= limit
}
