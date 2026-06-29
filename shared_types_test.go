package megaport

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		zero    bool
		wantErr bool
	}{
		{name: "epoch milliseconds number", input: `1700000000000`, want: time.Unix(1700000000, 0)},
		{name: "zero number", input: `0`, want: time.Unix(0, 0)},
		{name: "epoch milliseconds string", input: `"1700000000000"`, want: time.Unix(1700000000, 0)},
		{name: "rfc3339 string", input: `"2026-06-29T01:02:03Z"`, want: time.Date(2026, 6, 29, 1, 2, 3, 0, time.UTC)},
		{name: "rfc3339 with offset", input: `"2026-06-29T01:02:03+10:00"`, want: time.Date(2026, 6, 29, 1, 2, 3, 0, time.FixedZone("", 10*3600))},
		{name: "date only string", input: `"2026-06-29"`, want: time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC)},
		{name: "null", input: `null`, zero: true},
		{name: "empty string", input: `""`, zero: true},
		{name: "unparseable string", input: `"not-a-date"`, wantErr: true},
		{name: "compact digit string is not treated as epoch", input: `"20260629"`, wantErr: true},
		{name: "compact datetime string is not treated as epoch", input: `"20260629010203"`, wantErr: true},
		{name: "quoted epoch at min bound", input: `"946684800000"`, want: time.Unix(946684800, 0)},
		{name: "quoted epoch just below min bound errors", input: `"946684799999"`, wantErr: true},
		{name: "quoted epoch just below max bound", input: `"4102444799999"`, want: time.Unix(4102444799, 0)},
		{name: "quoted epoch at max bound errors", input: `"4102444800000"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Time
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.zero {
				assert.True(t, got.IsZero(), "expected zero time, got %s", got)
				return
			}
			assert.True(t, got.Equal(tt.want), "got %s, want %s", got, tt.want)
		})
	}
}

// A string-valued date must not fail decoding of the whole struct, which was
// the original fragility (the unmarshaller only accepted numbers). Covers both
// a pointer field (e.g. Port.CreateDate) and a value field (UserActivity), and
// pins the null handling for each.
func TestTimeUnmarshalInStruct(t *testing.T) {
	type ptrField struct {
		Name       string `json:"name"`
		CreateDate *Time  `json:"createDate"`
	}
	type valueField struct {
		CreateDate Time `json:"createDate"`
	}

	t.Run("pointer field with string date", func(t *testing.T) {
		var p ptrField
		require.NoError(t, json.Unmarshal([]byte(`{"name":"port-1","createDate":"2026-06-29T00:00:00Z"}`), &p))
		assert.Equal(t, "port-1", p.Name)
		require.NotNil(t, p.CreateDate)
		assert.Equal(t, 2026, p.CreateDate.Year())
	})

	t.Run("pointer field null stays nil", func(t *testing.T) {
		var p ptrField
		require.NoError(t, json.Unmarshal([]byte(`{"name":"port-1","createDate":null}`), &p))
		assert.Nil(t, p.CreateDate)
	})

	t.Run("value field with string date", func(t *testing.T) {
		var v valueField
		require.NoError(t, json.Unmarshal([]byte(`{"createDate":"2026-06-29T00:00:00Z"}`), &v))
		assert.Equal(t, 2026, v.CreateDate.Year())
	})

	// Deliberate change: null on a value Time field now yields the Go zero
	// time (IsZero), where older code decoded null to the 1970 epoch.
	t.Run("value field null is zero time", func(t *testing.T) {
		var v valueField
		require.NoError(t, json.Unmarshal([]byte(`{"createDate":null}`), &v))
		assert.True(t, v.CreateDate.IsZero())
	})
}

// Decoding must not normalize zones: epoch inputs keep the host-local zone of
// the historical numeric path, and string inputs keep the zone they were
// written in. Guards against re-introducing UTC normalization, which would
// churn downstream consumers that format these times into stored state.
func TestTimeUnmarshalPreservesZone(t *testing.T) {
	decode := func(in string) Time {
		var got Time
		require.NoError(t, json.Unmarshal([]byte(in), &got))
		return got
	}

	// Epoch, as a number and as a quoted string, stays host-local (time.Unix).
	assert.Equal(t, time.Local, decode(`1700000000000`).Location())
	assert.Equal(t, time.Local, decode(`"1700000000000"`).Location())

	// Zulu and date-only strings parse as UTC.
	assert.Equal(t, time.UTC, decode(`"2026-06-29T01:02:03Z"`).Location())
	assert.Equal(t, time.UTC, decode(`"2026-06-29"`).Location())

	// An explicit offset is preserved, not converted to UTC.
	_, offset := decode(`"2026-06-29T01:02:03+10:00"`).Zone()
	assert.Equal(t, 10*3600, offset)
}
