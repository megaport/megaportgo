package megaport

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	SERVICE_CONFIGURED = "CONFIGURED" // The CONFIGURED service state.
	SERVICE_LIVE       = "LIVE"       // The LIVE service state.
	// STATUS_DESIGN is the pre-order state for products that are created but
	// not yet validated or purchased (currently: NAT Gateways).
	STATUS_DESIGN = "DESIGN"

	// Product types
	PRODUCT_MEGAPORT    = "megaport"
	PRODUCT_VXC         = "vxc"
	PRODUCT_MCR         = "mcr2"
	PRODUCT_MVE         = "mve"
	PRODUCT_IX          = "ix"
	PRODUCT_NAT_GATEWAY = "nat_gateway"

	// Cancellation states
	STATUS_DECOMMISSIONED = "DECOMMISSIONED"
	STATUS_CANCELLED      = "CANCELLED"

	// Port Types
	SINGLE_PORT = "Single"
	LAG_PORT    = "LAG"

	// AWS VXC Types
	CONNECT_TYPE_AWS_VIF               = "AWS"
	CONNECT_TYPE_AWS_HOSTED_CONNECTION = "AWSHC"

	// InterfaceTypeSubInterface and InterfaceTypeIPSecTunnel are the
	// interface type values accepted by the Megaport API. They are
	// camelCase and the API matches them exactly.
	InterfaceTypeSubInterface = "subInterface"
	InterfaceTypeIPSecTunnel  = "ipSecTunnel"

	// maxCostCentreLength is the maximum number of characters the API accepts for a cost centre.
	maxCostCentreLength = 255
)

var (
	// VALID_CONTRACT_TERMS lists the valid contract terms in months.
	VALID_CONTRACT_TERMS = []int{1, 12, 24, 36, 48, 60}

	VALID_MCR_PORT_SPEEDS = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000, 400000}

	// SERVICE_STATE_READY is a list of service states that are considered ready for use.
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

// Time is a custom time type for Megaport API timestamps.
type Time struct {
	time.Time
}

// timeStringLayouts are the date string formats the Megaport API is known to
// return across environments, tried in order.
var timeStringLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// minEpochMillis and maxEpochMillis bound the range in which a quoted digit
// string is accepted as an epoch-millisecond timestamp (2000-01-01 to
// 2100-01-01 UTC). Outside it, a digit string is likelier a compact date
// (e.g. "20260629" or "20260629010203") than an epoch. No date layout matches
// a bare digit string, so such input errors rather than silently decoding to a
// 1970-era or far-future timestamp.
const (
	minEpochMillis = 946684800000  // 2000-01-01T00:00:00Z
	maxEpochMillis = 4102444800000 // 2100-01-01T00:00:00Z
)

// epochMillisToTime converts Unix epoch milliseconds to a time.Time, keeping
// the numeric path's historical host-local zone so existing responses decode
// exactly as before; only the new string and null forms are additive.
func epochMillisToTime(ms int64) time.Time {
	return time.Unix(ms/1000, 0)
}

// UnmarshalJSON parses a Megaport API timestamp. The API usually sends Unix
// epoch milliseconds as a JSON number, but some endpoints and environments
// send an ISO 8601 date string (or null) instead. Accepting all of these stops
// a single string-valued date from failing the decode of an entire response.
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}

	// Numeric form: Unix epoch milliseconds.
	if s[0] != '"' {
		var ms int64
		if err := json.Unmarshal(b, &ms); err != nil {
			return err
		}
		t.Time = epochMillisToTime(ms)
		return nil
	}

	// String form: a quoted epoch or a date string.
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	str = strings.TrimSpace(str)
	if str == "" {
		t.Time = time.Time{}
		return nil
	}
	if ms, err := strconv.ParseInt(str, 10, 64); err == nil && ms >= minEpochMillis && ms < maxEpochMillis {
		t.Time = epochMillisToTime(ms)
		return nil
	}
	for _, layout := range timeStringLayouts {
		if parsed, err := time.Parse(layout, str); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("megaport: cannot parse %q as a timestamp", str)
}
