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
}

// defaultAsyncJobTimeout is applied to WaitForAsync* calls when the caller
// passes a context without a deadline, to avoid blocking forever if the
// Looking Glass async job never transitions out of a pending state.
const defaultAsyncJobTimeout = 5 * time.Minute

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
