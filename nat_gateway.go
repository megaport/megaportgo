package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"
)

// NATGatewayService is an interface for interfacing with the NAT Gateway endpoints of the Megaport API.
type NATGatewayService interface {
	// CreateNATGateway creates a new NAT Gateway resource.
	CreateNATGateway(ctx context.Context, req *CreateNATGatewayRequest) (*NATGateway, error)
	// ListNATGateways retrieves all NAT Gateways for the authenticated company.
	ListNATGateways(ctx context.Context) ([]*NATGateway, error)
	// GetNATGateway retrieves a NAT Gateway by its product UID.
	GetNATGateway(ctx context.Context, productUID string) (*NATGateway, error)
	// UpdateNATGateway updates a NAT Gateway by its product UID.
	UpdateNATGateway(ctx context.Context, req *UpdateNATGatewayRequest) (*NATGateway, error)
	// DeleteNATGateway deletes a NAT Gateway by its product UID.
	DeleteNATGateway(ctx context.Context, productUID string) error
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
	ProductUID string     // The product UID of the NAT Gateway.
	Types      []string   // Telemetry types to retrieve, e.g. "BITS", "PACKETS", "SPEED".
	From       *time.Time // Start time. Mutually exclusive with Days.
	To         *time.Time // End time. Mutually exclusive with Days.
	Days       *int32     // Number of days of telemetry (1-180). Mutually exclusive with From/To.
}

// ErrNATGatewayProductUIDRequired is returned when a ProductUID is not provided.
var ErrNATGatewayProductUIDRequired = errors.New("product UID is required")

// ErrNATGatewayTelemetryTypesRequired is returned when no telemetry types are provided.
var ErrNATGatewayTelemetryTypesRequired = errors.New("at least one telemetry type is required")

// ErrNATGatewayTelemetryTimeExclusive is returned when both Days and From/To are provided.
var ErrNATGatewayTelemetryTimeExclusive = errors.New("days and from/to are mutually exclusive")

// ErrNATGatewayTelemetryDaysOutOfRange is returned when Days is not between 1 and 180.
var ErrNATGatewayTelemetryDaysOutOfRange = errors.New("days must be between 1 and 180")

// ErrNATGatewayTelemetryFromToIncomplete is returned when only one of From/To is provided.
var ErrNATGatewayTelemetryFromToIncomplete = errors.New("both from and to must be provided together")

// ErrNATGatewayProductNameRequired is returned when a ProductName is not provided.
var ErrNATGatewayProductNameRequired = errors.New("product name is required")

// ErrNATGatewayLocationIDRequired is returned when a LocationID is not provided or is invalid.
var ErrNATGatewayLocationIDRequired = errors.New("location ID must be greater than 0")

// ErrNATGatewaySpeedRequired is returned when a Speed is not provided or is invalid.
var ErrNATGatewaySpeedRequired = errors.New("speed must be greater than 0")

// ErrNATGatewayInvalidTerm is returned when a Term is not a valid contract term.
var ErrNATGatewayInvalidTerm = errors.New("term must be one of: 1, 12, 24, 36, 48, 60")

// validateCreateNATGatewayRequest validates the request parameters for creating a NAT Gateway.
func validateCreateNATGatewayRequest(req *CreateNATGatewayRequest) error {
	if req.ProductName == "" {
		return ErrNATGatewayProductNameRequired
	}
	if req.LocationID < 1 {
		return ErrNATGatewayLocationIDRequired
	}
	if req.Speed < 1 {
		return ErrNATGatewaySpeedRequired
	}
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return ErrNATGatewayInvalidTerm
	}
	return nil
}

// validateUpdateNATGatewayRequest validates the request parameters for updating a NAT Gateway.
func validateUpdateNATGatewayRequest(req *UpdateNATGatewayRequest) error {
	if req.ProductUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if req.ProductName == "" {
		return ErrNATGatewayProductNameRequired
	}
	if req.LocationID < 1 {
		return ErrNATGatewayLocationIDRequired
	}
	if req.Speed < 1 {
		return ErrNATGatewaySpeedRequired
	}
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return ErrNATGatewayInvalidTerm
	}
	return nil
}

// CreateNATGateway creates a new NAT Gateway resource.
func (svc *NATGatewayServiceOp) CreateNATGateway(ctx context.Context, req *CreateNATGatewayRequest) (*NATGateway, error) {
	if err := validateCreateNATGatewayRequest(req); err != nil {
		return nil, err
	}

	path := "/v3/products/nat_gateways"
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var natResp NATGatewayResponse
	if err := json.Unmarshal(buf.Bytes(), &natResp); err != nil {
		return nil, err
	}
	return &natResp.Data, nil
}

// ListNATGateways retrieves all NAT Gateways for the authenticated company.
func (svc *NATGatewayServiceOp) ListNATGateways(ctx context.Context) ([]*NATGateway, error) {
	path := "/v3/products/nat_gateways"
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var listResp NATGatewayListResponse
	if err := json.Unmarshal(buf.Bytes(), &listResp); err != nil {
		return nil, err
	}
	return listResp.Data, nil
}

// GetNATGateway retrieves a NAT Gateway by its product UID.
func (svc *NATGatewayServiceOp) GetNATGateway(ctx context.Context, productUID string) (*NATGateway, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", url.PathEscape(productUID))
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var natResp NATGatewayResponse
	if err := json.Unmarshal(buf.Bytes(), &natResp); err != nil {
		return nil, err
	}
	return &natResp.Data, nil
}

// UpdateNATGateway updates a NAT Gateway by its product UID.
func (svc *NATGatewayServiceOp) UpdateNATGateway(ctx context.Context, req *UpdateNATGatewayRequest) (*NATGateway, error) {
	if err := validateUpdateNATGatewayRequest(req); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", url.PathEscape(req.ProductUID))
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var natResp NATGatewayResponse
	if err := json.Unmarshal(buf.Bytes(), &natResp); err != nil {
		return nil, err
	}
	return &natResp.Data, nil
}

// DeleteNATGateway deletes a NAT Gateway by its product UID.
func (svc *NATGatewayServiceOp) DeleteNATGateway(ctx context.Context, productUID string) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", url.PathEscape(productUID))
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

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
	if req.Days != nil && (*req.Days < 1 || *req.Days > 180) {
		return ErrNATGatewayTelemetryDaysOutOfRange
	}
	if (req.From != nil) != (req.To != nil) {
		return ErrNATGatewayTelemetryFromToIncomplete
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
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
		params.Set("from", strconv.FormatInt(req.From.UnixMilli(), 10))
	}
	if req.To != nil {
		params.Set("to", strconv.FormatInt(req.To.UnixMilli(), 10))
	}
	if req.Days != nil {
		params.Set("days", strconv.FormatInt(int64(*req.Days), 10))
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	telemetryResp := &ServiceTelemetryResponse{}
	if err := json.Unmarshal(buf.Bytes(), telemetryResp); err != nil {
		return nil, err
	}
	return telemetryResp, nil
}
