package megaport

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetrySampleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TelemetrySample
		wantErr bool
	}{
		{
			name:  "timestamp and value",
			input: `[1700000000000, 12.5]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 12.5},
		},
		{
			name:  "integer value",
			input: `[1700000000000, 12]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 12},
		},
		{
			// The API sends null when a metric had no reading in the interval.
			name:  "null value decodes as zero",
			input: `[1700000000000, null]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 0},
		},
		{name: "null timestamp", input: `[null, 12.5]`, wantErr: true},
		{name: "too few elements", input: `[1700000000000]`, wantErr: true},
		{name: "too many elements", input: `[1700000000000, 12.5, 1]`, wantErr: true},
		{name: "not an array", input: `{"timestamp":1700000000000}`, wantErr: true},
		{name: "non-numeric value", input: `[1700000000000, "high"]`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TelemetrySample
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// A single null value must not fail the decode of the surrounding series.
func TestTelemetrySampleNullWithinSeries(t *testing.T) {
	var metric TelemetryMetricData
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "Bits",
		"subtype": "in",
		"samples": [[1700000000000, 1.5], [1700000060000, null], [1700000120000, 2.5]],
		"unit": {"name": "bps", "fullName": "bits per second"}
	}`), &metric))

	require.Len(t, metric.Samples, 3)
	assert.Equal(t, 1.5, metric.Samples[0].Value)
	assert.Equal(t, 0.0, metric.Samples[1].Value)
	assert.Equal(t, int64(1700000060000), metric.Samples[1].Timestamp)
	assert.Equal(t, 2.5, metric.Samples[2].Value)
}
