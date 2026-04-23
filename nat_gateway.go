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
	// ValidateNATGatewayOrder validates a NAT Gateway design via
	// POST /v3/networkdesign/validate. The gateway must be in DESIGN state.
	// Returns an order preview including pricing.
	ValidateNATGatewayOrder(ctx context.Context, productUID string) (*NATGatewayValidateResult, error)
	// BuyNATGateway purchases (provisions) a NAT Gateway design via
	// POST /v3/networkdesign/buy. The gateway must be in DESIGN state;
	// after a successful call it transitions through the normal
	// DEPLOYABLE -> CONFIGURED -> LIVE lifecycle. Returns the provisioning
	// service record.
	BuyNATGateway(ctx context.Context, productUID string) (*NATGatewayBuyResult, error)

	// ListNATGatewayPacketFilters returns all packet filter summaries for
	// a NAT Gateway.
	ListNATGatewayPacketFilters(ctx context.Context, productUID string) ([]*NATGatewayPacketFilterSummary, error)
	// CreateNATGatewayPacketFilter creates a new packet filter on a NAT
	// Gateway.
	CreateNATGatewayPacketFilter(ctx context.Context, productUID string, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error)
	// GetNATGatewayPacketFilter returns a packet filter by its numeric ID.
	GetNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) (*NATGatewayPacketFilter, error)
	// UpdateNATGatewayPacketFilter replaces a packet filter's description
	// and entries.
	UpdateNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error)
	// DeleteNATGatewayPacketFilter removes a packet filter from a NAT
	// Gateway. Any VXC interfaces referencing the filter will be detached
	// server-side.
	DeleteNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) error

	// ListNATGatewayPrefixLists returns all prefix list summaries for a
	// NAT Gateway.
	ListNATGatewayPrefixLists(ctx context.Context, productUID string) ([]*NATGatewayPrefixListSummary, error)
	// CreateNATGatewayPrefixList creates a new prefix list on a NAT
	// Gateway.
	CreateNATGatewayPrefixList(ctx context.Context, productUID string, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error)
	// GetNATGatewayPrefixList returns a prefix list by its numeric ID.
	GetNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) (*NATGatewayPrefixList, error)
	// UpdateNATGatewayPrefixList replaces a prefix list's description,
	// address family, and entries.
	UpdateNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error)
	// DeleteNATGatewayPrefixList removes a prefix list from a NAT Gateway.
	DeleteNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) error

	// ListNATGatewayIPRoutesAsync submits an IP routes diagnostics request
	// and returns the operation ID to poll with
	// GetNATGatewayDiagnosticsRoutes. The endpoint is rate-limited and
	// intended for troubleshooting only. If ipAddress is empty, the
	// response will include both IPv4 and IPv6 routes; otherwise the
	// response is narrowed to routes matching the supplied address.
	ListNATGatewayIPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error)
	// ListNATGatewayBGPRoutesAsync submits a BGP routes diagnostics
	// request and returns the operation ID to poll with
	// GetNATGatewayDiagnosticsRoutes. Rate-limited and intended for
	// troubleshooting only.
	ListNATGatewayBGPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error)
	// ListNATGatewayBGPNeighborRoutesAsync submits a BGP neighbor routes
	// diagnostics request and returns the operation ID to poll with
	// GetNATGatewayDiagnosticsRoutes.
	ListNATGatewayBGPNeighborRoutesAsync(ctx context.Context, req *NATGatewayBGPNeighborRoutesRequest) (string, error)
	// GetNATGatewayDiagnosticsRoutes retrieves the routes for a prior
	// asynchronous diagnostics request. Returns the heterogeneous slice
	// of IP and/or BGP routes produced by the async operation.
	GetNATGatewayDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*NATGatewayRoute, error)

	// ListNATGatewayIPRoutes submits an IP routes diagnostics request and
	// polls until the routes are available. The returned slice contains
	// only IP routes extracted from the heterogeneous result.
	ListNATGatewayIPRoutes(ctx context.Context, productUID, ipAddress string) ([]*NATGatewayIPRoute, error)
	// ListNATGatewayBGPRoutes submits a BGP routes diagnostics request
	// and polls until the routes are available.
	ListNATGatewayBGPRoutes(ctx context.Context, productUID, ipAddress string) ([]*NATGatewayBGPRoute, error)
	// ListNATGatewayBGPNeighborRoutes submits a BGP neighbor routes
	// diagnostics request and polls until the routes are available.
	ListNATGatewayBGPNeighborRoutes(ctx context.Context, req *NATGatewayBGPNeighborRoutesRequest) ([]*NATGatewayBGPRoute, error)
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

// Packet filter validation errors.
var (
	ErrNATGatewayPacketFilterIDRequired       = errors.New("packet filter ID must be greater than 0")
	ErrNATGatewayPacketFilterDescriptionEmpty = errors.New("packet filter description is required")
	ErrNATGatewayPacketFilterEntriesEmpty     = errors.New("packet filter requires at least one entry")
)

// Prefix list validation errors.
var (
	ErrNATGatewayPrefixListIDRequired         = errors.New("prefix list ID must be greater than 0")
	ErrNATGatewayPrefixListDescriptionEmpty   = errors.New("prefix list description is required")
	ErrNATGatewayPrefixListAddressFamilyEmpty = errors.New("prefix list addressFamily is required")
	ErrNATGatewayPrefixListEntriesEmpty       = errors.New("prefix list requires at least one entry")
)

// Diagnostics validation errors.
var (
	ErrNATGatewayDiagnosticsPeerIPRequired   = errors.New("BGP neighbor diagnostics require a peer IP address")
	ErrNATGatewayDiagnosticsDirectionInvalid = errors.New("BGP neighbor diagnostics require a direction (RECEIVED or ADVERTISED)")
	ErrNATGatewayDiagnosticsOperationEmpty   = errors.New("operation ID is required")
	ErrNATGatewayDiagnosticsTimeout          = errors.New("timed out waiting for diagnostics operation to complete")
)

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

// DeleteNATGateway deletes a NAT Gateway by its product UID. It handles
// both lifecycle stages transparently:
//
//   - DESIGN-state designs that have never been purchased use
//     DELETE /v3/products/nat_gateways/{uid} (the design-only endpoint).
//     This hard-removes the record — the gateway disappears from list.
//   - Any non-DESIGN gateway (e.g. DEPLOYABLE / CONFIGURED / LIVE) is
//     cancelled via the generic product action
//     POST /v3/product/{uid}/action/CANCEL_NOW, matching the teardown path
//     used for Ports, MCRs, MVEs, and VXCs. The record is retained rather
//     than being hard-deleted, typically transitioning to
//     DECOMMISSIONED / CANCELLED.
//
// Callers do not need to inspect state themselves. The design endpoint
// returns 400 for non-DESIGN gateways, and CANCEL_NOW rolls back against
// DESIGN-state records — so a single unified endpoint is not available from
// the API side, and the SDK routes based on a pre-flight GET. Errors from
// the pre-flight GET (including 404 for an unknown UID) are wrapped with
// a "nat gateway delete: could not inspect lifecycle state" prefix but
// preserve the underlying error chain (use errors.Is / errors.As).
//
// The routing is not atomic: if a DESIGN-state gateway transitions to
// DEPLOYABLE between the GET and the DELETE (e.g., another caller has just
// purchased it), the design endpoint will return 400. Retrying the delete
// will route through the provisioned path on the next attempt.
//
// Unlike DeletePort / DeleteMCR / DeleteMVE, this method does not currently
// accept a SafeDelete (end-of-term cancellation) option — provisioned
// gateways are always cancelled immediately with DeleteNow: true.
func (svc *NATGatewayServiceOp) DeleteNATGateway(ctx context.Context, productUID string) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}

	gw, err := svc.GetNATGateway(ctx, productUID)
	if err != nil {
		return fmt.Errorf("nat gateway delete: could not inspect lifecycle state: %w", err)
	}

	if gw.ProvisioningStatus == STATUS_DESIGN {
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

	_, err = svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: productUID,
		DeleteNow: true,
	})
	return err
}

// natGatewayOrderItem is the minimal payload expected by
// /v3/networkdesign/validate and /v3/networkdesign/buy for NAT Gateway
// designs. The endpoints accept an array of items.
type natGatewayOrderItem struct {
	ProductUID string `json:"productUid"`
}

// ErrNATGatewayOrderResponseEmpty is returned when the API response data
// array is empty (the endpoints are expected to return one entry per
// submitted productUid).
var ErrNATGatewayOrderResponseEmpty = errors.New("nat gateway order response contained no data")

// ValidateNATGatewayOrder validates a NAT Gateway design without purchasing.
// The returned result includes a pricing preview.
func (svc *NATGatewayServiceOp) ValidateNATGatewayOrder(ctx context.Context, productUID string) (*NATGatewayValidateResult, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	var envelope natGatewayValidateEnvelope
	if err := svc.postNetworkDesign(ctx, "/v3/networkdesign/validate", productUID, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Data) == 0 {
		return nil, ErrNATGatewayOrderResponseEmpty
	}
	return envelope.Data[0], nil
}

// BuyNATGateway purchases a NAT Gateway design, kicking off provisioning.
// The returned result contains the initial provisioning service record.
func (svc *NATGatewayServiceOp) BuyNATGateway(ctx context.Context, productUID string) (*NATGatewayBuyResult, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	var envelope natGatewayBuyEnvelope
	if err := svc.postNetworkDesign(ctx, "/v3/networkdesign/buy", productUID, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Data) == 0 {
		return nil, ErrNATGatewayOrderResponseEmpty
	}
	return envelope.Data[0], nil
}

func (svc *NATGatewayServiceOp) postNetworkDesign(ctx context.Context, path, productUID string, out interface{}) error {
	body := []natGatewayOrderItem{{ProductUID: productUID}}
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.Unmarshal(buf.Bytes(), out); err != nil {
			return err
		}
	}
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

// --- Packet filters -------------------------------------------------------

func validateNATGatewayPacketFilterRequest(req *NATGatewayPacketFilterRequest) error {
	if req == nil || req.Description == "" {
		return ErrNATGatewayPacketFilterDescriptionEmpty
	}
	if len(req.Entries) == 0 {
		return ErrNATGatewayPacketFilterEntriesEmpty
	}
	return nil
}

// ListNATGatewayPacketFilters returns all packet filter summaries for a NAT Gateway.
func (svc *NATGatewayServiceOp) ListNATGatewayPacketFilters(ctx context.Context, productUID string) ([]*NATGatewayPacketFilterSummary, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filter_summaries", url.PathEscape(productUID))
	var envelope natGatewayPacketFilterSummariesResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// CreateNATGatewayPacketFilter creates a new packet filter on a NAT Gateway.
func (svc *NATGatewayServiceOp) CreateNATGatewayPacketFilter(ctx context.Context, productUID string, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if err := validateNATGatewayPacketFilterRequest(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters", url.PathEscape(productUID))
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodPost, path, req, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// GetNATGatewayPacketFilter returns a packet filter by its numeric ID.
func (svc *NATGatewayServiceOp) GetNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return nil, ErrNATGatewayPacketFilterIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// UpdateNATGatewayPacketFilter replaces a packet filter's description and entries.
func (svc *NATGatewayServiceOp) UpdateNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return nil, ErrNATGatewayPacketFilterIDRequired
	}
	if err := validateNATGatewayPacketFilterRequest(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodPut, path, req, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// DeleteNATGatewayPacketFilter removes a packet filter from a NAT Gateway.
func (svc *NATGatewayServiceOp) DeleteNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return ErrNATGatewayPacketFilterIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	return svc.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

// --- Prefix lists ---------------------------------------------------------

func validateNATGatewayPrefixList(req *NATGatewayPrefixList) error {
	if req == nil || req.Description == "" {
		return ErrNATGatewayPrefixListDescriptionEmpty
	}
	if req.AddressFamily == "" {
		return ErrNATGatewayPrefixListAddressFamilyEmpty
	}
	if len(req.Entries) == 0 {
		return ErrNATGatewayPrefixListEntriesEmpty
	}
	return nil
}

// ListNATGatewayPrefixLists returns all prefix list summaries for a NAT Gateway.
func (svc *NATGatewayServiceOp) ListNATGatewayPrefixLists(ctx context.Context, productUID string) ([]*NATGatewayPrefixListSummary, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_list_summaries", url.PathEscape(productUID))
	var envelope natGatewayPrefixListSummariesResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// CreateNATGatewayPrefixList creates a new prefix list on a NAT Gateway.
func (svc *NATGatewayServiceOp) CreateNATGatewayPrefixList(ctx context.Context, productUID string, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if err := validateNATGatewayPrefixList(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists", url.PathEscape(productUID))
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodPost, path, req.toAPI(), &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, nil
	}
	return envelope.Data.toPrefixList()
}

// GetNATGatewayPrefixList returns a prefix list by its numeric ID.
func (svc *NATGatewayServiceOp) GetNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return nil, ErrNATGatewayPrefixListIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, nil
	}
	return envelope.Data.toPrefixList()
}

// UpdateNATGatewayPrefixList replaces a prefix list's description, address family, and entries.
func (svc *NATGatewayServiceOp) UpdateNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return nil, ErrNATGatewayPrefixListIDRequired
	}
	if err := validateNATGatewayPrefixList(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodPut, path, req.toAPI(), &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, nil
	}
	return envelope.Data.toPrefixList()
}

// DeleteNATGatewayPrefixList removes a prefix list from a NAT Gateway.
func (svc *NATGatewayServiceOp) DeleteNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return ErrNATGatewayPrefixListIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	return svc.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

// --- Diagnostics ----------------------------------------------------------

// diagnosticsPollInterval and diagnosticsPollTimeout are the defaults used by
// the convenience diagnostics wrappers.
const (
	diagnosticsPollInitialDelay = 2 * time.Second
	diagnosticsPollInterval     = 3 * time.Second
	diagnosticsPollTimeout      = 60 * time.Second
)

// ListNATGatewayIPRoutesAsync submits an IP routes diagnostics request.
func (svc *NATGatewayServiceOp) ListNATGatewayIPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	if productUID == "" {
		return "", ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/diagnostics/routes/ip", url.PathEscape(productUID))
	if ipAddress != "" {
		params := url.Values{}
		params.Set("ip_address", ipAddress)
		path = path + "?" + params.Encode()
	}
	var envelope natGatewayDiagnosticsAsyncResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return "", err
	}
	return envelope.Data, nil
}

// ListNATGatewayBGPRoutesAsync submits a BGP routes diagnostics request.
func (svc *NATGatewayServiceOp) ListNATGatewayBGPRoutesAsync(ctx context.Context, productUID, ipAddress string) (string, error) {
	if productUID == "" {
		return "", ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/diagnostics/routes/bgp", url.PathEscape(productUID))
	if ipAddress != "" {
		params := url.Values{}
		params.Set("ip_address", ipAddress)
		path = path + "?" + params.Encode()
	}
	var envelope natGatewayDiagnosticsAsyncResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return "", err
	}
	return envelope.Data, nil
}

// ListNATGatewayBGPNeighborRoutesAsync submits a BGP neighbor routes diagnostics request.
func (svc *NATGatewayServiceOp) ListNATGatewayBGPNeighborRoutesAsync(ctx context.Context, req *NATGatewayBGPNeighborRoutesRequest) (string, error) {
	if req == nil || req.ProductUID == "" {
		return "", ErrNATGatewayProductUIDRequired
	}
	if req.PeerIPAddress == "" {
		return "", ErrNATGatewayDiagnosticsPeerIPRequired
	}
	if req.Direction != BGPRouteDirectionReceived && req.Direction != BGPRouteDirectionAdvertised {
		return "", ErrNATGatewayDiagnosticsDirectionInvalid
	}
	params := url.Values{}
	params.Set("direction", req.Direction)
	params.Set("peer_ip_address", req.PeerIPAddress)
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/diagnostics/routes/bgp/neighbor?%s", url.PathEscape(req.ProductUID), params.Encode())
	var envelope natGatewayDiagnosticsAsyncResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return "", err
	}
	return envelope.Data, nil
}

// GetNATGatewayDiagnosticsRoutes fetches the routes for a prior async request.
func (svc *NATGatewayServiceOp) GetNATGatewayDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*NATGatewayRoute, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if operationID == "" {
		return nil, ErrNATGatewayDiagnosticsOperationEmpty
	}
	params := url.Values{}
	params.Set("operationId", operationID)
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/diagnostics/routes/operation?%s", url.PathEscape(productUID), params.Encode())
	var envelope natGatewayDiagnosticsRoutesResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// pollDiagnosticsRoutes polls GetNATGatewayDiagnosticsRoutes until the
// operation returns a non-empty result, or the caller-supplied timeout
// elapses. Empty responses are treated as "still processing".
func (svc *NATGatewayServiceOp) pollDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*NATGatewayRoute, error) {
	pollCtx, cancel := context.WithTimeout(ctx, diagnosticsPollTimeout)
	defer cancel()
	select {
	case <-pollCtx.Done():
		return nil, ErrNATGatewayDiagnosticsTimeout
	case <-time.After(diagnosticsPollInitialDelay):
	}
	ticker := time.NewTicker(diagnosticsPollInterval)
	defer ticker.Stop()
	for {
		routes, err := svc.GetNATGatewayDiagnosticsRoutes(pollCtx, productUID, operationID)
		if err != nil {
			return nil, err
		}
		if len(routes) > 0 {
			return routes, nil
		}
		select {
		case <-pollCtx.Done():
			return nil, ErrNATGatewayDiagnosticsTimeout
		case <-ticker.C:
		}
	}
}

// ListNATGatewayIPRoutes submits an IP routes request and polls until results are available.
func (svc *NATGatewayServiceOp) ListNATGatewayIPRoutes(ctx context.Context, productUID, ipAddress string) ([]*NATGatewayIPRoute, error) {
	opID, err := svc.ListNATGatewayIPRoutesAsync(ctx, productUID, ipAddress)
	if err != nil {
		return nil, err
	}
	routes, err := svc.pollDiagnosticsRoutes(ctx, productUID, opID)
	if err != nil {
		return nil, err
	}
	out := make([]*NATGatewayIPRoute, 0, len(routes))
	for _, r := range routes {
		if r.IP != nil {
			out = append(out, r.IP)
		}
	}
	return out, nil
}

// ListNATGatewayBGPRoutes submits a BGP routes request and polls until results are available.
func (svc *NATGatewayServiceOp) ListNATGatewayBGPRoutes(ctx context.Context, productUID, ipAddress string) ([]*NATGatewayBGPRoute, error) {
	opID, err := svc.ListNATGatewayBGPRoutesAsync(ctx, productUID, ipAddress)
	if err != nil {
		return nil, err
	}
	routes, err := svc.pollDiagnosticsRoutes(ctx, productUID, opID)
	if err != nil {
		return nil, err
	}
	out := make([]*NATGatewayBGPRoute, 0, len(routes))
	for _, r := range routes {
		if r.BGP != nil {
			out = append(out, r.BGP)
		}
	}
	return out, nil
}

// ListNATGatewayBGPNeighborRoutes submits a BGP neighbor routes request and polls for results.
func (svc *NATGatewayServiceOp) ListNATGatewayBGPNeighborRoutes(ctx context.Context, req *NATGatewayBGPNeighborRoutesRequest) ([]*NATGatewayBGPRoute, error) {
	opID, err := svc.ListNATGatewayBGPNeighborRoutesAsync(ctx, req)
	if err != nil {
		return nil, err
	}
	routes, err := svc.pollDiagnosticsRoutes(ctx, req.ProductUID, opID)
	if err != nil {
		return nil, err
	}
	out := make([]*NATGatewayBGPRoute, 0, len(routes))
	for _, r := range routes {
		if r.BGP != nil {
			out = append(out, r.BGP)
		}
	}
	return out, nil
}

// --- shared helper --------------------------------------------------------

// doJSON sends a JSON request and decodes the response into out (or discards
// the body if out is nil). It centralises the NewRequest/Do/Unmarshal dance
// shared by the packet filter, prefix list, and diagnostics methods.
func (svc *NATGatewayServiceOp) doJSON(ctx context.Context, method, path string, body, out interface{}) error {
	clientReq, err := svc.Client.NewRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil || buf.Len() == 0 {
		return nil
	}
	return json.Unmarshal(buf.Bytes(), out)
}
