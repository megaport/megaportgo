package megaport

import (
	"encoding/json"
	"fmt"
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

	// Interface types for VXC vRouter / NAT Gateway A-End partner configs.
	InterfaceTypeSubInterface = "subInterface"
	InterfaceTypeIPSecTunnel  = "ipSecTunnel"
)

var (
	// VALID_CONTRACT_TERMS lists the valid contract terms in months.
	VALID_CONTRACT_TERMS = []int{1, 12, 24, 36, 48, 60}

	VALID_MCR_PORT_SPEEDS = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000, 400000}

	// SERVICE_STATE_READY is a list of service states that are considered ready for use.
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

// ServiceTelemetryResponse is the API response for service telemetry data.
// This response is NOT wrapped in the standard message/terms/data envelope.
// It is shared across Port, MCR, MVE, VXC, and IX services.
type ServiceTelemetryResponse struct {
	ServiceUID string                 `json:"serviceUid"`
	Type       string                 `json:"type"`
	TimeFrame  TelemetryTimeFrame     `json:"timeFrame"`
	Data       []*TelemetryMetricData `json:"data"`
	PeerUID    string                 `json:"peerUid,omitempty"` // only present for IX flow metrics
}

// TelemetryTimeFrame represents the time range of a telemetry response.
type TelemetryTimeFrame struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// TelemetryMetricData represents a single metric series in a telemetry response.
type TelemetryMetricData struct {
	Type    string            `json:"type"`
	Subtype string            `json:"subtype"`
	Samples []TelemetrySample `json:"samples"`
	Unit    TelemetryUnit     `json:"unit"`
}

// TelemetrySample represents a single data point in a telemetry series.
// The API returns samples as [timestamp, value] tuples.
type TelemetrySample struct {
	Timestamp int64
	Value     float64
}

// UnmarshalJSON handles the [int64, float64] tuple format from the API.
func (s *TelemetrySample) UnmarshalJSON(data []byte) error {
	var tuple []json.Number
	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf("telemetry sample must be a JSON array: %w", err)
	}
	if len(tuple) != 2 {
		return fmt.Errorf("telemetry sample must have exactly 2 elements, got %d", len(tuple))
	}
	ts, err := tuple[0].Int64()
	if err != nil {
		return fmt.Errorf("telemetry sample timestamp: %w", err)
	}
	val, err := tuple[1].Float64()
	if err != nil {
		return fmt.Errorf("telemetry sample value: %w", err)
	}
	s.Timestamp = ts
	s.Value = val
	return nil
}

// TelemetryUnit describes the unit of measurement for a telemetry metric.
type TelemetryUnit struct {
	Name     string `json:"name"`
	FullName string `json:"fullName"`
}

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
