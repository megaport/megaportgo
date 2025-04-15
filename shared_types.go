package megaport

import (
	"encoding/json"
	"time"
)

const (
	SERVICE_CONFIGURED = "CONFIGURED" // The CONFIGURED service state.
	SERVICE_LIVE       = "LIVE"       // The LIVE service state.

	// Product types
	PRODUCT_MEGAPORT = "megaport"
	PRODUCT_VXC      = "vxc"
	PRODUCT_MCR      = "mcr2"
	PRODUCT_MVE      = "mve"
	PRODUCT_IX       = "ix"

	// Cancellation states
	STATUS_DECOMMISSIONED = "DECOMMISSIONED"
	STATUS_CANCELLED      = "CANCELLED"

	// Port Types
	SINGLE_PORT = "Single"
	LAG_PORT    = "LAG"

	// AWS VXC Types
	CONNECT_TYPE_AWS_VIF               = "AWS"
	CONNECT_TYPE_AWS_HOSTED_CONNECTION = "AWSHC"
)

var (
	// VALID_CONTRACT_TERMS lists the valid contract terms in months.
	VALID_CONTRACT_TERMS = []int{1, 12, 24, 36}

	VALID_MCR_PORT_SPEEDS = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000}

	// SERVICE_STATE_READY is a list of service states that are considered ready for use.
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

// Time is a custom time type that allows for unmarshalling of Unix timestamps.
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshals a Unix timestamp into a Time type.
func (t *Time) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	t.Time = time.Unix(timestamp/1000, 0) // Divide by 1000 to convert from milliseconds to seconds
	return nil
}
