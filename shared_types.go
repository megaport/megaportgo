package megaport

import (
	"encoding/json"
	"time"
)

const (
	SERVICE_CONFIGURED = "CONFIGURED" // The CONFIGURED service state.
	SERVICE_LIVE       = "LIVE"       // The LIVE service state.

	PRODUCT_MEGAPORT = "megaport"
	PRODUCT_VXC      = "vxc"
	PRODUCT_MCR      = "mcr2"
	PRODUCT_MVE      = "mve"
	PRODUCT_IX       = "ix"

	STATUS_DECOMMISSIONED = "DECOMMISSIONED"
	STATUS_CANCELLED      = "CANCELLED"

	SINGLE_PORT                        = "Single"
	LAG_PORT                           = "LAG"
	CONNECT_TYPE_AWS_VIF               = "AWS"
	CONNECT_TYPE_AWS_HOSTED_CONNECTION = "AWSHC"
)

var (
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

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

type Time struct {
 	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	t.Time = GetTime(timestamp) // Divide by 1000 to convert from milliseconds to seconds
	return nil
}
