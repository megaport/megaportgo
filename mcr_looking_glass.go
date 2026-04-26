package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// MCRLookingGlassService is an interface for interfacing with the MCR Looking Glass endpoints
// of the Megaport API. The Looking Glass provides visibility into traffic routing,
// helping you troubleshoot connections by showing the status of protocols and
// routing tables in the MCR.
type MCRLookingGlassService interface {
	// ListIPRoutes retrieves the IP routing table from the MCR Looking Glass.
	// This returns all routes (BGP, static, connected, local) that the MCR knows about.
	ListIPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassIPRoute, error)
	// ListIPRoutesWithFilter retrieves the IP routing table with optional filtering.
	ListIPRoutesWithFilter(ctx context.Context, req *ListIPRoutesRequest) ([]*LookingGlassIPRoute, error)
	// ListBGPRoutes retrieves BGP routes from the MCR Looking Glass.
	// This returns routes learned via BGP with full BGP attributes.
	ListBGPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassBGPRoute, error)
	// ListBGPRoutesWithFilter retrieves BGP routes with optional filtering.
	ListBGPRoutesWithFilter(ctx context.Context, req *ListBGPRoutesRequest) ([]*LookingGlassBGPRoute, error)
	// ListBGPSessions retrieves all BGP sessions configured on the MCR.
	ListBGPSessions(ctx context.Context, mcrUID string) ([]*LookingGlassBGPSession, error)
	// ListBGPNeighborRoutes retrieves routes advertised to or received from a specific BGP neighbor.
	ListBGPNeighborRoutes(ctx context.Context, req *ListBGPNeighborRoutesRequest) ([]*LookingGlassBGPNeighborRoute, error)
	// ListIPRoutesAsync initiates an async query for IP routes and returns the job ID.
	// Use GetAsyncIPRoutes to poll for results.
	ListIPRoutesAsync(ctx context.Context, mcrUID string) (*LookingGlassAsyncJob, error)
	// GetAsyncIPRoutes retrieves the results of an async IP routes query.
	GetAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncIPRoutesData, error)
	// ListBGPNeighborRoutesAsync initiates an async query for BGP neighbor routes.
	// Use GetAsyncBGPNeighborRoutes to poll for results.
	ListBGPNeighborRoutesAsync(ctx context.Context, req *ListBGPNeighborRoutesRequest) (*LookingGlassAsyncJob, error)
	// GetAsyncBGPNeighborRoutes retrieves the results of an async BGP neighbor routes query.
	GetAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncBGPNeighborRoutesData, error)
	// WaitForAsyncIPRoutes polls for async IP routes results until the job
	// completes or the context is cancelled. Callers control the overall
	// timeout by passing a context with a deadline; if the context has no
	// deadline, a default of defaultAsyncJobTimeout is applied.
	WaitForAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string) ([]*LookingGlassIPRoute, error)
	// WaitForAsyncBGPNeighborRoutes polls for async BGP neighbor routes
	// results until the job completes or the context is cancelled. Callers
	// control the overall timeout by passing a context with a deadline; if
	// the context has no deadline, a default of defaultAsyncJobTimeout is
	// applied.
	WaitForAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string) ([]*LookingGlassBGPNeighborRoute, error)
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

// defaultAsyncJobTimeout is applied to WaitForAsync* calls when the caller
// passes a context without a deadline, to avoid blocking forever if the
// Looking Glass async job never transitions out of a pending state.
const defaultAsyncJobTimeout = 5 * time.Minute

// mcrDiagnosticsPollTimeout is the SDK-managed timeout for WaitForMCRPing and
// WaitForMCRTraceroute when the caller does not provide a context with a deadline.
const mcrDiagnosticsPollTimeout = 5 * time.Minute

// mcrDiagnosticsPollInterval is the interval between poll attempts for MCR diagnostics.
const mcrDiagnosticsPollInterval = 3 * time.Second

// MCRLookingGlassServiceOp handles communication with MCR Looking Glass methods of the Megaport API.
type MCRLookingGlassServiceOp struct {
	Client *Client
}

// NewMCRLookingGlassService creates a new instance of the MCR Looking Glass Service.
func NewMCRLookingGlassService(c *Client) *MCRLookingGlassServiceOp {
	return &MCRLookingGlassServiceOp{
		Client: c,
	}
}

var _ MCRLookingGlassService = (*MCRLookingGlassServiceOp)(nil)

// ListIPRoutes retrieves the IP routing table from the MCR Looking Glass.
func (svc *MCRLookingGlassServiceOp) ListIPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassIPRoute, error) {
	return svc.ListIPRoutesWithFilter(ctx, &ListIPRoutesRequest{MCRID: mcrUID})
}

// ListIPRoutesWithFilter retrieves the IP routing table with optional filtering.
func (svc *MCRLookingGlassServiceOp) ListIPRoutesWithFilter(ctx context.Context, req *ListIPRoutesRequest) ([]*LookingGlassIPRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list IP routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, fmt.Errorf("list IP routes request MCRID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", url.PathEscape(req.MCRID))

	// Build query parameters
	params := url.Values{}
	if req.Protocol != "" {
		params.Set("protocol", string(req.Protocol))
	}
	if req.IPFilter != "" {
		params.Set("ip", req.IPFilter)
	}
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

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

	apiResponse := &LookingGlassIPRoutesResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPRoutes retrieves BGP routes from the MCR Looking Glass.
func (svc *MCRLookingGlassServiceOp) ListBGPRoutes(ctx context.Context, mcrUID string) ([]*LookingGlassBGPRoute, error) {
	return svc.ListBGPRoutesWithFilter(ctx, &ListBGPRoutesRequest{MCRID: mcrUID})
}

// ListBGPRoutesWithFilter retrieves BGP routes with optional filtering.
func (svc *MCRLookingGlassServiceOp) ListBGPRoutesWithFilter(ctx context.Context, req *ListBGPRoutesRequest) ([]*LookingGlassBGPRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list BGP routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, fmt.Errorf("list BGP routes request MCRID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgp", url.PathEscape(req.MCRID))

	// Build query parameters
	params := url.Values{}
	if req.IPFilter != "" {
		params.Set("ip", req.IPFilter)
	}
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

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

	apiResponse := &LookingGlassBGPRoutesResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPSessions retrieves all BGP sessions configured on the MCR.
func (svc *MCRLookingGlassServiceOp) ListBGPSessions(ctx context.Context, mcrUID string) ([]*LookingGlassBGPSession, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("list BGP sessions request MCRID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions", url.PathEscape(mcrUID))

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

	apiResponse := &LookingGlassBGPSessionsResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPNeighborRoutes retrieves routes advertised to or received from a specific BGP neighbor.
func (svc *MCRLookingGlassServiceOp) ListBGPNeighborRoutes(ctx context.Context, req *ListBGPNeighborRoutesRequest) ([]*LookingGlassBGPNeighborRoute, error) {
	if req == nil {
		return nil, fmt.Errorf("list BGP neighbor routes request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, fmt.Errorf("list BGP neighbor routes request MCRID cannot be empty")
	}
	if req.SessionID == "" {
		return nil, fmt.Errorf("list BGP neighbor routes request SessionID cannot be empty")
	}
	switch req.Direction {
	case LookingGlassRouteDirectionAdvertised, LookingGlassRouteDirectionReceived:
	default:
		return nil, fmt.Errorf("list BGP neighbor routes request Direction must be one of: advertised, received")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/%s",
		url.PathEscape(req.MCRID), url.PathEscape(req.SessionID), req.Direction)

	// Build query parameters
	params := url.Values{}
	if req.IPFilter != "" {
		params.Set("ip", req.IPFilter)
	}
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

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

	apiResponse := &LookingGlassBGPNeighborRoutesResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListIPRoutesAsync initiates an async query for IP routes.
func (svc *MCRLookingGlassServiceOp) ListIPRoutesAsync(ctx context.Context, mcrUID string) (*LookingGlassAsyncJob, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("list IP routes async request MCRID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", url.PathEscape(mcrUID))

	// Build query parameters
	params := url.Values{}
	params.Set("async", "true")
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

	apiResponse := &LookingGlassAsyncJobResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// GetAsyncIPRoutes retrieves the results of an async IP routes query.
func (svc *MCRLookingGlassServiceOp) GetAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncIPRoutesData, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("get async IP routes request MCRID cannot be empty")
	}
	if jobID == "" {
		return nil, fmt.Errorf("get async IP routes request jobID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes/async/%s", url.PathEscape(mcrUID), url.PathEscape(jobID))

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

	apiResponse := &LookingGlassAsyncIPRoutesResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPNeighborRoutesAsync initiates an async query for BGP neighbor routes.
func (svc *MCRLookingGlassServiceOp) ListBGPNeighborRoutesAsync(ctx context.Context, req *ListBGPNeighborRoutesRequest) (*LookingGlassAsyncJob, error) {
	if req == nil {
		return nil, fmt.Errorf("list BGP neighbor routes async request cannot be nil")
	}
	if req.MCRID == "" {
		return nil, fmt.Errorf("list BGP neighbor routes async request MCRID cannot be empty")
	}
	if req.SessionID == "" {
		return nil, fmt.Errorf("list BGP neighbor routes async request SessionID cannot be empty")
	}
	switch req.Direction {
	case LookingGlassRouteDirectionAdvertised, LookingGlassRouteDirectionReceived:
	default:
		return nil, fmt.Errorf("list BGP neighbor routes async request Direction must be one of: advertised, received")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/%s",
		url.PathEscape(req.MCRID), url.PathEscape(req.SessionID), req.Direction)

	// Build query parameters
	params := url.Values{}
	params.Set("async", "true")
	if req.IPFilter != "" {
		params.Set("ip", req.IPFilter)
	}
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

	apiResponse := &LookingGlassAsyncJobResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// GetAsyncBGPNeighborRoutes retrieves the results of an async BGP neighbor routes query.
func (svc *MCRLookingGlassServiceOp) GetAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncBGPNeighborRoutesData, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("get async BGP neighbor routes request MCRID cannot be empty")
	}
	if jobID == "" {
		return nil, fmt.Errorf("get async BGP neighbor routes request jobID cannot be empty")
	}
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/async/%s", url.PathEscape(mcrUID), url.PathEscape(jobID))

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

	apiResponse := &LookingGlassAsyncBGPNeighborRoutesResponse{}
	if err := json.Unmarshal(buf.Bytes(), apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// WaitForAsyncIPRoutes polls for async IP routes results until the job
// completes or the context is cancelled. If the context has no deadline,
// defaultAsyncJobTimeout is applied so callers who pass a bare context
// are still protected from hanging indefinitely.
func (svc *MCRLookingGlassServiceOp) WaitForAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string) ([]*LookingGlassIPRoute, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("wait for async IP routes request MCRID cannot be empty")
	}
	if jobID == "" {
		return nil, fmt.Errorf("wait for async IP routes request jobID cannot be empty")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultAsyncJobTimeout)
		defer cancel()
	}

	// Check immediately before starting the ticker to return results without
	// delay when the job is already complete.
	result, err := svc.GetAsyncIPRoutes(ctx, mcrUID, jobID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("async IP routes job %s returned nil result", jobID)
	}
	switch result.Status {
	case LookingGlassAsyncStatusComplete:
		return result.Routes, nil
	case LookingGlassAsyncStatusFailed:
		return nil, fmt.Errorf("async IP routes job %s failed", jobID)
	case LookingGlassAsyncStatusPending, LookingGlassAsyncStatusProcessing:
		// Continue to ticker-based polling below.
	default:
		return nil, fmt.Errorf("unknown async job status: %s", result.Status)
	}

	// Looking Glass async jobs are diagnostic and typically complete faster than provisioning
	// workflows (which use a 30s polling interval). We poll more frequently (5s) here to
	// return results sooner while still avoiding excessive request volume.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for async IP routes job %s: %w", jobID, ctx.Err())
		case <-ticker.C:
			result, err := svc.GetAsyncIPRoutes(ctx, mcrUID, jobID)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, fmt.Errorf("async IP routes job %s returned nil result", jobID)
			}

			switch result.Status {
			case LookingGlassAsyncStatusComplete:
				return result.Routes, nil
			case LookingGlassAsyncStatusFailed:
				return nil, fmt.Errorf("async IP routes job %s failed", jobID)
			case LookingGlassAsyncStatusPending, LookingGlassAsyncStatusProcessing:
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown async job status: %s", result.Status)
			}
		}
	}
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
		pollCtx, cancel = context.WithTimeout(ctx, mcrDiagnosticsPollTimeout)
		defer cancel()
	}

	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err // caller cancelled or deadline exceeded
		}
		return ErrMCRDiagnosticsTimeout // SDK-managed timeout fired
	}

	// Poll immediately — return without delay when the result is already available.
	result, err := svc.GetMCRPingResult(pollCtx, mcrUID, operationID)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	ticker := time.NewTicker(mcrDiagnosticsPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return nil, pollDoneErr()
		case <-ticker.C:
			result, err := svc.GetMCRPingResult(pollCtx, mcrUID, operationID)
			if err != nil {
				return nil, err
			}
			if result != nil {
				return result, nil
			}
		}
	}
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
		pollCtx, cancel = context.WithTimeout(ctx, mcrDiagnosticsPollTimeout)
		defer cancel()
	}

	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err // caller cancelled or deadline exceeded
		}
		return ErrMCRDiagnosticsTimeout // SDK-managed timeout fired
	}

	// Poll immediately — return without delay when the result is already available.
	result, err := svc.GetMCRTracerouteResult(pollCtx, mcrUID, operationID)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	ticker := time.NewTicker(mcrDiagnosticsPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return nil, pollDoneErr()
		case <-ticker.C:
			result, err := svc.GetMCRTracerouteResult(pollCtx, mcrUID, operationID)
			if err != nil {
				return nil, err
			}
			if result != nil {
				return result, nil
			}
		}
	}
}

// WaitForAsyncBGPNeighborRoutes polls for async BGP neighbor routes results
// until the job completes or the context is cancelled. If the context has
// no deadline, defaultAsyncJobTimeout is applied so callers who pass a
// bare context are still protected from hanging indefinitely.
func (svc *MCRLookingGlassServiceOp) WaitForAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string) ([]*LookingGlassBGPNeighborRoute, error) {
	if mcrUID == "" {
		return nil, fmt.Errorf("wait for async BGP neighbor routes request MCRID cannot be empty")
	}
	if jobID == "" {
		return nil, fmt.Errorf("wait for async BGP neighbor routes request jobID cannot be empty")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultAsyncJobTimeout)
		defer cancel()
	}

	// Check immediately before starting the ticker to return results without
	// delay when the job is already complete.
	result, err := svc.GetAsyncBGPNeighborRoutes(ctx, mcrUID, jobID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("async BGP neighbor routes job %s returned nil result", jobID)
	}
	switch result.Status {
	case LookingGlassAsyncStatusComplete:
		return result.Routes, nil
	case LookingGlassAsyncStatusFailed:
		return nil, fmt.Errorf("async BGP neighbor routes job %s failed", jobID)
	case LookingGlassAsyncStatusPending, LookingGlassAsyncStatusProcessing:
		// Continue to ticker-based polling below.
	default:
		return nil, fmt.Errorf("unknown async job status: %s", result.Status)
	}

	// Looking Glass async jobs are diagnostic and typically complete faster than provisioning
	// workflows (which use a 30s polling interval). We poll more frequently (5s) here to
	// return results sooner while still avoiding excessive request volume.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for async BGP neighbor routes job %s: %w", jobID, ctx.Err())
		case <-ticker.C:
			result, err := svc.GetAsyncBGPNeighborRoutes(ctx, mcrUID, jobID)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, fmt.Errorf("async BGP neighbor routes job %s returned nil result", jobID)
			}

			switch result.Status {
			case LookingGlassAsyncStatusComplete:
				return result.Routes, nil
			case LookingGlassAsyncStatusFailed:
				return nil, fmt.Errorf("async BGP neighbor routes job %s failed", jobID)
			case LookingGlassAsyncStatusPending, LookingGlassAsyncStatusProcessing:
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown async job status: %s", result.Status)
			}
		}
	}
}
