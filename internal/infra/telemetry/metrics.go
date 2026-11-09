package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	res, err := newResource(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("create metric reader: %w", err)
	}
	return sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(reader)), nil
}
