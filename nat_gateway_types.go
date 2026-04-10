package megaport

import (
	"encoding/json"
	"fmt"
)

// NATGatewaySession represents a speed/session-count availability entry for NAT Gateways.
type NATGatewaySession struct {
	SessionCount []int `json:"sessionCount"`
	SpeedMbps    int   `json:"speedMbps"`
}

// NATGatewaySessionsResponse is the API response for listing NAT Gateway sessions.
type NATGatewaySessionsResponse struct {
	Message string               `json:"message"`
	Terms   string               `json:"terms"`
	Data    []*NATGatewaySession `json:"data"`
}

// ServiceTelemetryResponse is the API response for service telemetry data.
// This response is NOT wrapped in the standard message/terms/data envelope.
type ServiceTelemetryResponse struct {
	ServiceUID string                 `json:"serviceUid"`
	Type       string                 `json:"type"`
	TimeFrame  TelemetryTimeFrame     `json:"timeFrame"`
	Data       []*TelemetryMetricData `json:"data"`
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

// NATGateway represents a NAT Gateway product from the Megaport API.
type NATGateway struct {
	AdminLocked           bool                    `json:"adminLocked"`
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	ContractEndDate       string                  `json:"contractEndDate"`
	CreateDate            string                  `json:"createDate"`
	CreatedBy             string                  `json:"createdBy"`
	LocationID            int                     `json:"locationId"`
	Locked                bool                    `json:"locked"`
	OrderApprovalStatus   string                  `json:"orderApprovalStatus"`
	ProductName           string                  `json:"productName"`
	ProductUID            string                  `json:"productUid"`
	PromoCode             string                  `json:"promoCode"`
	ProvisioningStatus    string                  `json:"provisioningStatus"`
	ResourceTags          []ResourceTag           `json:"resourceTags"`
	ServiceLevelReference string                  `json:"serviceLevelReference"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// NATGatewayNetworkConfig represents the network configuration for a NAT Gateway.
type NATGatewayNetworkConfig struct {
	ASN                int    `json:"asn"`
	BGPShutdownDefault bool   `json:"bgpShutdownDefault"`
	DiversityZone      string `json:"diversityZone"`
	SessionCount       int    `json:"sessionCount"`
}

// CreateNATGatewayRequest represents a request to create a NAT Gateway.
type CreateNATGatewayRequest struct {
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	LocationID            int                     `json:"locationId"`
	ProductName           string                  `json:"productName"`
	PromoCode             string                  `json:"promoCode,omitempty"`
	ResourceTags          []ResourceTag           `json:"resourceTags,omitempty"`
	ServiceLevelReference string                  `json:"serviceLevelReference,omitempty"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// UpdateNATGatewayRequest represents a request to update a NAT Gateway.
type UpdateNATGatewayRequest struct {
	ProductUID            string                  `json:"-"` // path parameter, not serialized
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	LocationID            int                     `json:"locationId"`
	ProductName           string                  `json:"productName"`
	PromoCode             string                  `json:"promoCode,omitempty"`
	ResourceTags          []ResourceTag           `json:"resourceTags,omitempty"`
	ServiceLevelReference string                  `json:"serviceLevelReference,omitempty"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// NATGatewayResponse is the API response for a single NAT Gateway.
type NATGatewayResponse struct {
	Message string     `json:"message"`
	Terms   string     `json:"terms"`
	Data    NATGateway `json:"data"`
}

// NATGatewayListResponse is the API response for listing NAT Gateways.
type NATGatewayListResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []*NATGateway `json:"data"`
}

// DeleteNATGatewayResponse is the API response for deleting a NAT Gateway.
type DeleteNATGatewayResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
}
