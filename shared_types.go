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
	SINGLE_PORT                        = "Single"
	LAG_PORT                           = "LAG"

	// AWS VXC Types
	CONNECT_TYPE_AWS_VIF               = "AWS"
	CONNECT_TYPE_AWS_HOSTED_CONNECTION = "AWSHC"
)

var (
	// SERVICE_STATE_READY is a list of service states that are considered ready for use.
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

// GenericResponse is a generic response structure for API responses.
type GenericResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    map[string]interface{} `json:"data"`
}

const APPLICATION_SHORT_NAME = "mpt"
const MODIFY_NAME string = "NAME"
const MODIFY_COST_CENTRE = "COST_CENTRE"
const MODIFY_MARKETPLACE_VISIBILITY string = "MARKETPLACE_VISIBILITY"
const MODIFY_RATE_LIMIT = "RATE_LIMIT"
const MODIFY_A_END_VLAN = "A_VLAN"
const MODIFY_B_END_VLAN = "B_VLAN"

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
	t.Time = GetTime(timestamp) // Divide by 1000 to convert from milliseconds to seconds
	return nil
}
