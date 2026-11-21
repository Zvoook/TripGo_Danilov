package trip

import (
	"testing"
	"time"
)

// A row with database timestamps in a non-UTC timezone.
type timestampRow struct {
	at       time.Time
	finished bool
}

func (r timestampRow) Scan(dest ...any) error {
	*dest[9].(*time.Time) = r.at
	*dest[10].(*time.Time) = r.at
	*dest[11].(*time.Time) = r.at
	if r.finished {
		at := r.at
		*dest[12].(**time.Time) = &at
	}
	return nil
}

func TestScanTripUTC(t *testing.T) {
	at := time.Date(2026, 10, 6, 12, 30, 0, 123456000, time.FixedZone("UTC+3", 3*60*60))
	for _, finished := range []bool{false, true} {
		name := "active"
		if finished {
			name = "completed"
		}
		t.Run(name, func(t *testing.T) {
			got, err := scanTrip(timestampRow{at: at, finished: finished})
			if err != nil {
				t.Fatal(err)
			}
			times := []time.Time{got.StartedAt, got.CreatedAt, got.UpdatedAt}
			if finished {
				if got.FinishedAt == nil {
					t.Fatal("finished_at was lost")
				}
				times = append(times, *got.FinishedAt)
			} else if got.FinishedAt != nil {
				t.Fatal("active trip has finished_at")
			}
			for _, value := range times {
				if value.Location() != time.UTC || !value.Equal(at) {
					t.Fatalf("timestamp changed or not UTC: %v", value)
				}
			}
		})
	}
}
