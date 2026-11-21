// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package translator

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Malformed Datadog v1 points must not panic or prevent valid points from being translated.
func TestTranslateSeriesV1SkipsNullPoints(t *testing.T) {
	for _, metricType := range []string{TypeGauge, TypeCount, TypeRate} {
		t.Run(metricType, func(t *testing.T) {
			firstTS, secondTS := float64(1636629071), float64(1636629081)
			firstValue, secondValue := 1.5, 2.5
			metricTypeCopy := metricType

			input := SeriesList{Series: []datadogV1.Series{{
				Metric: "probe.null",
				Type:   &metricTypeCopy,
				Points: [][]*float64{
					nil,                          // Missing both fields.
					{},                           // Empty point.
					{&firstTS},                   // Missing value.
					{nil, &firstValue},           // Null timestamp.
					{&firstTS, nil},              // Null value.
					{nil, nil},                   // Both null.
					{&firstTS, &firstValue},      // Valid point after malformed points.
					{&firstTS, &firstValue, nil}, // Extra field.
					{&secondTS, &secondValue},    // Another valid point.
				},
			}}}

			got := createMetricsTranslator().TranslateSeriesV1(input)
			require.Equal(t, 2, got.DataPointCount())
			metricSlice := got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
			require.Equal(t, 1, metricSlice.Len())

			var points pmetric.NumberDataPointSlice
			if metricType == TypeGauge {
				points = metricSlice.At(0).Gauge().DataPoints()
			} else {
				points = metricSlice.At(0).Sum().DataPoints()
			}
			require.Equal(t, 2, points.Len())
			require.Equal(t, pcommon.Timestamp(int64(firstTS)*int64(time.Second)), points.At(0).Timestamp())
			require.Equal(t, firstValue, points.At(0).DoubleValue())
			require.Equal(t, pcommon.Timestamp(int64(secondTS)*int64(time.Second)), points.At(1).Timestamp())
			require.Equal(t, secondValue, points.At(1).DoubleValue())
		})
	}
}
