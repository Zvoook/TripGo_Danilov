package http

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestResponseSchemas(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile("../../contracts/openapi/trip-service.openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		body map[string]any
	}{
		{"Trip", map[string]any{
			"id":          "00000000-0000-0000-0000-000000000001",
			"user_id":     "00000000-0000-0000-0000-000000000002",
			"driver_id":   "00000000-0000-0000-0000-000000000003",
			"start_point": map[string]any{"latitude": 10.0, "longitude": 20.0},
			"end_point":   map[string]any{"latitude": 30.0, "longitude": 40.0},
			"price":       100.0, "status": "active", "started_at": "2026-10-06T09:00:00Z",
		}},
		{"TripPosition", map[string]any{
			"id": 1.0, "trip_id": "00000000-0000-0000-0000-000000000001",
			"coordinates": map[string]any{"latitude": 10.0, "longitude": 20.0},
			"recorded_at": "2026-10-06T09:00:00Z",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			schema := doc.Components.Schemas[tt.name].Value
			if err := schema.VisitJSON(tt.body); err != nil {
				t.Fatalf("valid response rejected: %v", err)
			}
			tt.body["unexpected"] = true
			if err := schema.VisitJSON(tt.body); err == nil {
				t.Fatal("unknown response property accepted")
			}
		})
	}
}
