package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// MCRLookingGlassService reads MCR route and connectivity diagnostics.
// The route methods start a route diagnostics operation and poll for the result.
// They block until it completes. When the caller's context has no deadline, the
// SDK stops polling after 5 minutes. The submit leg is not bounded by that.
type MCRLookingGlassService interface {
	// ListIPRoutes retrieves the IP routing table from the MCR.
	ListIPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassIPRoute, error)
	// ListIPRoutesWithFilter retrieves the IP routing table, optionally narrowed to one IP address or prefix.
	ListIPRoutesWithFilter(ctx context.Context, req *ListIPRoutesRequest) ([]*LookingGlassIPRoute, error)
	// ListBGPRoutes retrieves the BGP routes the MCR knows about.
	ListBGPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassBGPRoute, error)
	// ListBGPRoutesWithFilter retrieves BGP routes, optionally narrowed to one IP address or prefix.
	ListBGPRoutesWithFilter(ctx context.Context, req *ListBGPRoutesRequest) ([]*LookingGlassBGPRoute, error)
	// ListBGPNeighborRoutes retrieves the routes advertised to or received from one BGP peer.
	ListBGPNeighborRoutes(ctx context.Context, req *ListBGPNeighborRoutesRequest) ([]*LookingGlassBGPRoute, error)
	// PingMCR initiates an ICMP ping from the MCR and returns the operation ID to poll with GetMCRPingResult.
	PingMCR(ctx context.Context, req *MCRPingRequest) (string, error)
	// TracerouteMCR initiates a traceroute from the MCR and returns the operation ID to poll with GetMCRTracerouteResult.
	TracerouteMCR(ctx context.Context, req *MCRTracerouteRequest) (string, error)
	// GetMCRPingResult retrieves the result of a pending ping operation. Returns nil result when still pending.
	GetMCRPingResult(ctx context.Context, mcrUID, operationID string) (*LookingGlassPingResult, error)
	// GetMCRTracerouteResult retrieves the result of a pending traceroute operation. Returns nil result when still pending.
	GetMCRTracerouteResult(ctx context.Context, mcrUID, operationID string) (*LookingGlassTracerouteResult, error)
	// WaitForMCRPing polls until the ping result is available or context is cancelled.
	WaitForMCRPing(ctx context.Context, mcrUID, operationID string) (*LookingGlassPingResult, error)
	// WaitForMCRTraceroute polls until the traceroute result is available or context is cancelled.
	WaitForMCRTraceroute(ctx context.Context, mcrUID, operationID string) (*LookingGlassTracerouteResult, error)
}

// mcrDiagnosticsPollTimeout bounds the poll loop when the caller's context has
// no deadline.
const mcrDiagnosticsPollTimeout = 5 * time.Minute

// mcrDiagnosticsPollInterval is the interval between poll attempts for MCR diagnostics.
const mcrDiagnosticsPollInterval = 3 * time.Second

// MCRLookingGlassServiceOp handles communication with MCR Looking Glass methods of the Megaport API.
type MCRLookingGlassServiceOp struct {
	Client *Client
	// pollInterval overrides mcrDiagnosticsPollInterval when non-zero.
	// Intended for tests that want to avoid real-time waits.
	pollInterval time.Duration
	// pollTimeout overrides mcrDiagnosticsPollTimeout when non-zero.
	// Intended for tests that want to avoid real-time waits.
	pollTimeout time.Duration
}

// NewMCRLookingGlassService creates a new instance of the MCR Looking Glass Service.
func NewMCRLookingGlassService(c *Client) *MCRLookingGlassServiceOp {
	return &MCRLookingGlassServiceOp{
		Client: c,
	}
}

// diagnosticsPollInterval returns the effective poll interval for MCR diagnostics.
func (svc *MCRLookingGlassServiceOp) diagnosticsPollInterval() time.Duration {
	if svc.pollInterval != 0 {
		return svc.pollInterval
	}
	return mcrDiagnosticsPollInterval
}

// diagnosticsPollTimeout returns the effective SDK-managed poll timeout for MCR diagnostics.
func (svc *MCRLookingGlassServiceOp) diagnosticsPollTimeout() time.Duration {
	if svc.pollTimeout != 0 {
		return svc.pollTimeout
	}
	return mcrDiagnosticsPollTimeout
}

// pollMCRDiagnostics is a generic helper that polls fetch until it returns a
// non-nil result, pollCtx is done, or pollDoneErr fires. It performs an
// immediate poll before starting the ticker so callers receive results without
// delay when the operation is already complete.
func pollMCRDiagnostics[T any](
	pollCtx context.Context,
	pollDoneErr func() error,
	fetch func(context.Context) (*T, error),
	interval time.Duration,
) (*T, error) {
	// When the poll context expires mid-request, Client.Do returns a wrapped
	// context error. Attribute it to the deadline/cancellation via pollDoneErr
	// so callers get a consistent error instead of a raw context error, while
	// leaving genuine API/network failures untouched.
	mapErr := func(err error) error {
		if pollCtx.Err() != nil && (errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
			return pollDoneErr()
		}
		return err
	}

	result, err := fetch(pollCtx)
	if err != nil {
		return nil, mapErr(err)
	}
	if result != nil {
		return result, nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return nil, pollDoneErr()
		case <-ticker.C:
			result, err = fetch(pollCtx)
			if err != nil {
				return nil, mapErr(err)
			}
			if result != nil {
				return result, nil
			}
		}
	}
}

var _ MCRLookingGlassService = (*MCRLookingGlassServiceOp)(nil)

// submitRouteDiagnostics starts a route diagnostics operation and returns the
// operation ID to poll. The spec says these endpoints must run in async mode and
// that sync mode is deprecated, so the operation always asks for async.
func (svc *MCRLookingGlassServiceOp) submitRouteDiagnostics(ctx context.Context, path string, params url.Values) (string, error) {
	query := url.Values{}
	for key, values := range params {
		query[key] = values
	}
	query.Set("async", "true")

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path+"?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	apiResponse := &mcrDiagnosticsStringResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return "", err
	}
	if apiResponse.Data == "" {
		return "", ErrMCRDiagnosticsOperationEmpty
	}

	return apiResponse.Data, nil
}

func routeDiagnosticsPath(mcrUID, suffix string) string {
	return fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes%s", url.PathEscape(mcrUID), suffix)
}

// getRouteOperationResult returns nil while the result is not ready.
func getRouteOperationResult[T any](ctx context.Context, svc *MCRLookingGlassServiceOp, mcrUID, operationID string) (*[]*T, error) {
	params := url.Values{}
	params.Set("operationId", operationID)
	path := routeDiagnosticsPath(mcrUID, "/operation") + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// The API answers 202 while it still collects the result. An empty route list
	// is also a 200, so only the status code separates the two.
	if response.StatusCode == http.StatusAccepted {
		return nil, nil
	}

	apiResponse := &mcrDiagnosticsRoutesResponse[T]{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	// Normalize a null or absent data key to an empty slice, so "no routes" has
	// one shape rather than two.
	routes := apiResponse.Data
	if routes == nil {
		routes = []*T{}
	}

	return &routes, nil
}

// listRouteDiagnostics submits the operation, then polls until the result is ready.
func listRouteDiagnostics[T any](ctx context.Context, svc *MCRLookingGlassServiceOp, mcrUID, suffix string, params url.Values) ([]*T, error) {
	operationID, err := svc.submitRouteDiagnostics(ctx, routeDiagnosticsPath(mcrUID, suffix), params)
	if err != nil {
		return nil, err
	}

	// pollCtx carries the SDK-managed deadline; ctx is the caller's original
	// context. pollDoneErr distinguishes caller cancellation from SDK timeout.
	pollCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, svc.diagnosticsPollTimeout())
		defer cancel()
	}

	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err // caller cancelled or deadline exceeded
		}
		return ErrMCRDiagnosticsTimeout // SDK-managed timeout fired
	}

	routes, err := pollMCRDiagnostics(
		pollCtx,
		pollDoneErr,
		func(c context.Context) (*[]*T, error) {
			return getRouteOperationResult[T](c, svc, mcrUID, operationID)
		},
		svc.diagnosticsPollInterval(),
	)
	if err != nil {
		return nil, err
	}

	return *routes, nil
}

// ListIPRoutes retrieves the IP routing table from the MCR.
func (svc *MCRLookingGlassServiceOp) ListIPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassIPRoute, error) {
	return svc.ListIPRoutesWithFilter(ctx, &ListIPRoutesRequest{MCRID: mcrUID})
}

// ListIPRoutesWithFilter retrieves the IP routing table, optionally narrowed to one IP address or prefix.
func (svc *MCRLookingGlassServiceOp) ListIPRoutesWithFilter(ctx context.Context, req *ListIPRoutesRequest) ([]*LookingGlassIPRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list IP routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}

	params := url.Values{}
	if req.IPFilter != "" {
		params.Set("ip_address", req.IPFilter)
	}

	return listRouteDiagnostics[LookingGlassIPRoute](ctx, svc, req.MCRID, "/ip", params)
}

// ListBGPRoutes retrieves the BGP routes the MCR knows about.
func (svc *MCRLookingGlassServiceOp) ListBGPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassBGPRoute, error) {
	return svc.ListBGPRoutesWithFilter(ctx, &ListBGPRoutesRequest{MCRID: mcrUID})
}

// ListBGPRoutesWithFilter retrieves BGP routes, optionally narrowed to one IP address or prefix.
func (svc *MCRLookingGlassServiceOp) ListBGPRoutesWithFilter(ctx context.Context, req *ListBGPRoutesRequest) ([]*LookingGlassBGPRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list BGP routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}

	params := url.Values{}
	if req.IPFilter != "" {
		params.Set("ip_address", req.IPFilter)
	}

	return listRouteDiagnostics[LookingGlassBGPRoute](ctx, svc, req.MCRID, "/bgp", params)
}

// ListBGPNeighborRoutes retrieves the routes advertised to or received from one BGP peer.
func (svc *MCRLookingGlassServiceOp) ListBGPNeighborRoutes(ctx context.Context, req *ListBGPNeighborRoutesRequest) ([]*LookingGlassBGPRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list BGP neighbor routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}
	if req.PeerIPAddress == "" {
		return nil, ErrMCRDiagnosticsPeerIPRequired
	}
	if req.Direction != BGPRouteDirectionReceived && req.Direction != BGPRouteDirectionAdvertised {
		return nil, ErrMCRDiagnosticsDirectionInvalid
	}

	params := url.Values{}
	params.Set("direction", req.Direction)
	params.Set("peer_ip_address", req.PeerIPAddress)

	return listRouteDiagnostics[LookingGlassBGPRoute](ctx, svc, req.MCRID, "/bgp/neighbor", params)
}

// PingMCR initiates an ICMP ping from the MCR and returns the operation ID.
func (svc *MCRLookingGlassServiceOp) PingMCR(ctx context.Context, req *MCRPingRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("ping request cannot be nil")
	}
	if req.MCRID == "" {
		return "", ErrMCRDiagnosticsMCRUIDRequired
	}
	if req.DestinationAddress == "" {
		return "", ErrMCRPingDestinationRequired
	}
	if req.PacketCount != nil && (*req.PacketCount < 1 || *req.PacketCount > 60) {
		return "", ErrMCRPingPacketCountOutOfRange
	}
	if req.PacketSize != nil && (*req.PacketSize < 1 || *req.PacketSize > 9186) {
		return "", ErrMCRPingPacketSizeOutOfRange
	}

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/ping", url.PathEscape(req.MCRID))
	params := url.Values{}
	params.Set("destination_address", req.DestinationAddress)
	if req.SourceAddress != "" {
		params.Set("source_address", req.SourceAddress)
	}
	if req.PacketCount != nil {
		params.Set("packet_count", fmt.Sprintf("%d", *req.PacketCount))
	}
	if req.PacketSize != nil {
		params.Set("packet_size", fmt.Sprintf("%d", *req.PacketSize))
	}
	path = path + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	apiResponse := &mcrDiagnosticsStringResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return "", err
	}
	if apiResponse.Data == "" {
		return "", ErrMCRDiagnosticsOperationEmpty
	}

	return apiResponse.Data, nil
}

// TracerouteMCR initiates a traceroute from the MCR and returns the operation ID.
func (svc *MCRLookingGlassServiceOp) TracerouteMCR(ctx context.Context, req *MCRTracerouteRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("traceroute request cannot be nil")
	}
	if req.MCRID == "" {
		return "", ErrMCRDiagnosticsMCRUIDRequired
	}
	if req.DestinationAddress == "" {
		return "", ErrMCRTracerouteDestinationRequired
	}

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/traceroute", url.PathEscape(req.MCRID))
	params := url.Values{}
	params.Set("destination_address", req.DestinationAddress)
	if req.SourceAddress != "" {
		params.Set("source_address", req.SourceAddress)
	}
	path = path + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	apiResponse := &mcrDiagnosticsStringResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return "", err
	}
	if apiResponse.Data == "" {
		return "", ErrMCRDiagnosticsOperationEmpty
	}

	return apiResponse.Data, nil
}

// GetMCRPingResult retrieves the result of a pending ping operation. Returns nil when still pending.
func (svc *MCRLookingGlassServiceOp) GetMCRPingResult(ctx context.Context, mcrUID, operationID string) (*LookingGlassPingResult, error) {
	if mcrUID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}
	if operationID == "" {
		return nil, ErrMCRDiagnosticsOperationEmpty
	}

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", url.PathEscape(mcrUID))
	params := url.Values{}
	params.Set("operationId", operationID)
	path = path + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	apiResponse := &mcrDiagnosticsPingResultResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// GetMCRTracerouteResult retrieves the result of a pending traceroute operation. Returns nil when still pending.
func (svc *MCRLookingGlassServiceOp) GetMCRTracerouteResult(ctx context.Context, mcrUID, operationID string) (*LookingGlassTracerouteResult, error) {
	if mcrUID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}
	if operationID == "" {
		return nil, ErrMCRDiagnosticsOperationEmpty
	}

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", url.PathEscape(mcrUID))
	params := url.Values{}
	params.Set("operationId", operationID)
	path = path + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	response, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	apiResponse := &mcrDiagnosticsTracerouteResultResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// WaitForMCRPing polls until the ping result is available or context is cancelled.
// If the context has no deadline, mcrDiagnosticsPollTimeout is applied.
func (svc *MCRLookingGlassServiceOp) WaitForMCRPing(ctx context.Context, mcrUID, operationID string) (*LookingGlassPingResult, error) {
	if mcrUID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}
	if operationID == "" {
		return nil, ErrMCRDiagnosticsOperationEmpty
	}

	// pollCtx carries the SDK-managed deadline; ctx is the caller's original
	// context. pollDoneErr distinguishes caller cancellation from SDK timeout.
	pollCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, svc.diagnosticsPollTimeout())
		defer cancel()
	}

	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err // caller cancelled or deadline exceeded
		}
		return ErrMCRDiagnosticsTimeout // SDK-managed timeout fired
	}

	return pollMCRDiagnostics(
		pollCtx,
		pollDoneErr,
		func(c context.Context) (*LookingGlassPingResult, error) {
			return svc.GetMCRPingResult(c, mcrUID, operationID)
		},
		svc.diagnosticsPollInterval(),
	)
}

// WaitForMCRTraceroute polls until the traceroute result is available or context is cancelled.
// If the context has no deadline, mcrDiagnosticsPollTimeout is applied.
func (svc *MCRLookingGlassServiceOp) WaitForMCRTraceroute(ctx context.Context, mcrUID, operationID string) (*LookingGlassTracerouteResult, error) {
	if mcrUID == "" {
		return nil, ErrMCRDiagnosticsMCRUIDRequired
	}
	if operationID == "" {
		return nil, ErrMCRDiagnosticsOperationEmpty
	}

	// pollCtx carries the SDK-managed deadline; ctx is the caller's original
	// context. pollDoneErr distinguishes caller cancellation from SDK timeout.
	pollCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, svc.diagnosticsPollTimeout())
		defer cancel()
	}

	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err // caller cancelled or deadline exceeded
		}
		return ErrMCRDiagnosticsTimeout // SDK-managed timeout fired
	}

	return pollMCRDiagnostics(
		pollCtx,
		pollDoneErr,
		func(c context.Context) (*LookingGlassTracerouteResult, error) {
			return svc.GetMCRTracerouteResult(c, mcrUID, operationID)
		},
		svc.diagnosticsPollInterval(),
	)
}
