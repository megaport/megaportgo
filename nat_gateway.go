package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// NATGatewayService is an interface for interfacing with the NAT Gateway endpoints of the Megaport API.
type NATGatewayService interface {
	// ListNATGatewaySessions returns the speed/session-count availability matrix for NAT Gateways.
	ListNATGatewaySessions(ctx context.Context) ([]*NATGatewaySession, error)
	// GetNATGatewayTelemetry returns telemetry data for a NAT Gateway product.
	GetNATGatewayTelemetry(ctx context.Context, req *GetNATGatewayTelemetryRequest) (*ServiceTelemetryResponse, error)
}

// NewNATGatewayService creates a new instance of the NAT Gateway Service.
func NewNATGatewayService(c *Client) *NATGatewayServiceOp {
	return &NATGatewayServiceOp{
		Client: c,
	}
}

// NATGatewayServiceOp handles communication with NAT Gateway methods of the Megaport API.
type NATGatewayServiceOp struct {
	Client *Client
}

// GetNATGatewayTelemetryRequest represents a request to get telemetry data for a NAT Gateway.
type GetNATGatewayTelemetryRequest struct {
	ProductUID string   // The product UID of the NAT Gateway.
	Types      []string // Telemetry types to retrieve, e.g. "BITS", "PACKETS", "SPEED".
	From       *int64   // Start time in epoch milliseconds. Mutually exclusive with Days.
	To         *int64   // End time in epoch milliseconds. Mutually exclusive with Days.
	Days       *int32   // Number of days of telemetry (1-180). Mutually exclusive with From/To.
}

// ErrNATGatewayProductUIDRequired is returned when a ProductUID is not provided.
var ErrNATGatewayProductUIDRequired = errors.New("product UID is required")

// ErrNATGatewayTelemetryTypesRequired is returned when no telemetry types are provided.
var ErrNATGatewayTelemetryTypesRequired = errors.New("at least one telemetry type is required")

// ErrNATGatewayTelemetryTimeExclusive is returned when both Days and From/To are provided.
var ErrNATGatewayTelemetryTimeExclusive = errors.New("days and from/to are mutually exclusive")

// validateGetNATGatewayTelemetryRequest validates the request parameters.
func validateGetNATGatewayTelemetryRequest(req *GetNATGatewayTelemetryRequest) error {
	if req.ProductUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if len(req.Types) == 0 {
		return ErrNATGatewayTelemetryTypesRequired
	}
	if req.Days != nil && (req.From != nil || req.To != nil) {
		return ErrNATGatewayTelemetryTimeExclusive
	}
	return nil
}

// ListNATGatewaySessions returns the speed/session-count availability matrix for NAT Gateways.
func (svc *NATGatewayServiceOp) ListNATGatewaySessions(ctx context.Context) ([]*NATGatewaySession, error) {
	path := "/v3/products/nat_gateways/sessions"
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	sessionsResp := NATGatewaySessionsResponse{}
	if err := json.Unmarshal(buf.Bytes(), &sessionsResp); err != nil {
		return nil, err
	}
	return sessionsResp.Data, nil
}

// GetNATGatewayTelemetry returns telemetry data for a NAT Gateway product.
func (svc *NATGatewayServiceOp) GetNATGatewayTelemetry(ctx context.Context, req *GetNATGatewayTelemetryRequest) (*ServiceTelemetryResponse, error) {
	if err := validateGetNATGatewayTelemetryRequest(req); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v3/products/nat_gateways/%s/telemetry", url.PathEscape(req.ProductUID))

	params := url.Values{}
	for _, t := range req.Types {
		params.Add("type", t)
	}
	if req.From != nil {
		params.Set("from", strconv.FormatInt(*req.From, 10))
	}
	if req.To != nil {
		params.Set("to", strconv.FormatInt(*req.To, 10))
	}
	if req.Days != nil {
		params.Set("days", strconv.FormatInt(int64(*req.Days), 10))
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	telemetryResp := &ServiceTelemetryResponse{}
	if err := json.Unmarshal(buf.Bytes(), telemetryResp); err != nil {
		return nil, err
	}
	return telemetryResp, nil
}
