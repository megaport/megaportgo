package megaport

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var fromNumber, fromString Time
	require.NoError(t, json.Unmarshal([]byte(`1700000000000`), &fromNumber))
	require.NoError(t, json.Unmarshal([]byte(`"2023-11-14T22:13:20Z"`), &fromString))

	assert.True(t, fromNumber.Equal(time.Unix(1700000000, 0)))
	assert.True(t, fromString.Equal(fromNumber.Time))

	var withOffset Time
	require.NoError(t, json.Unmarshal([]byte(`"2023-11-15T09:13:20+11:00"`), &withOffset))
	assert.True(t, withOffset.Equal(fromNumber.Time))

	var invalid Time
	assert.Error(t, json.Unmarshal([]byte(`"not a date"`), &invalid))
}
