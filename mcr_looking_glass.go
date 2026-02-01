package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	// WaitForAsyncIPRoutes polls for async IP routes results until complete or timeout.
	WaitForAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string, timeout time.Duration) ([]*LookingGlassIPRoute, error)
	// WaitForAsyncBGPNeighborRoutes polls for async BGP neighbor routes results until complete or timeout.
	WaitForAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string, timeout time.Duration) ([]*LookingGlassBGPNeighborRoute, error)
}

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
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", req.MCRID)

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

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassIPRoutesResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
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
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgp", req.MCRID)

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

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassBGPRoutesResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPSessions retrieves all BGP sessions configured on the MCR.
func (svc *MCRLookingGlassServiceOp) ListBGPSessions(ctx context.Context, mcrUID string) ([]*LookingGlassBGPSession, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions", mcrUID)

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassBGPSessionsResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPNeighborRoutes retrieves routes advertised to or received from a specific BGP neighbor.
func (svc *MCRLookingGlassServiceOp) ListBGPNeighborRoutes(ctx context.Context, req *ListBGPNeighborRoutesRequest) ([]*LookingGlassBGPNeighborRoute, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/%s",
		req.MCRID, req.SessionID, req.Direction)

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

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassBGPNeighborRoutesResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListIPRoutesAsync initiates an async query for IP routes.
func (svc *MCRLookingGlassServiceOp) ListIPRoutesAsync(ctx context.Context, mcrUID string) (*LookingGlassAsyncJob, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", mcrUID)

	// Build query parameters
	params := url.Values{}
	params.Set("async", "true")
	path = path + "?" + params.Encode()

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassAsyncJobResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// GetAsyncIPRoutes retrieves the results of an async IP routes query.
func (svc *MCRLookingGlassServiceOp) GetAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncIPRoutesData, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes/async/%s", mcrUID, jobID)

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassAsyncIPRoutesResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// ListBGPNeighborRoutesAsync initiates an async query for BGP neighbor routes.
func (svc *MCRLookingGlassServiceOp) ListBGPNeighborRoutesAsync(ctx context.Context, req *ListBGPNeighborRoutesRequest) (*LookingGlassAsyncJob, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/%s",
		req.MCRID, req.SessionID, req.Direction)

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

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassAsyncJobResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// GetAsyncBGPNeighborRoutes retrieves the results of an async BGP neighbor routes query.
func (svc *MCRLookingGlassServiceOp) GetAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string) (*AsyncBGPNeighborRoutesData, error) {
	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/async/%s", mcrUID, jobID)

	clientReq, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &LookingGlassAsyncBGPNeighborRoutesResponse{}
	if err := json.Unmarshal(body, apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

// WaitForAsyncIPRoutes polls for async IP routes results until complete or timeout.
func (svc *MCRLookingGlassServiceOp) WaitForAsyncIPRoutes(ctx context.Context, mcrUID string, jobID string, timeout time.Duration) ([]*LookingGlassIPRoute, error) {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Looking Glass async jobs are diagnostic and typically complete faster than provisioning
	// workflows (which use a 30s polling interval). We poll more frequently (5s) here to
	// return results sooner while still avoiding excessive request volume.
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(timeout)
	defer ticker.Stop()
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for async IP routes job %s", jobID)
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled waiting for async IP routes job %s", jobID)
		case <-ticker.C:
			result, err := svc.GetAsyncIPRoutes(ctx, mcrUID, jobID)
			if err != nil {
				return nil, err
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

// WaitForAsyncBGPNeighborRoutes polls for async BGP neighbor routes results until complete or timeout.
func (svc *MCRLookingGlassServiceOp) WaitForAsyncBGPNeighborRoutes(ctx context.Context, mcrUID string, jobID string, timeout time.Duration) ([]*LookingGlassBGPNeighborRoute, error) {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Looking Glass async jobs are diagnostic and typically complete faster than provisioning
	// workflows (which use a 30s polling interval). We poll more frequently (5s) here to
	// return results sooner while still avoiding excessive request volume.
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(timeout)
	defer ticker.Stop()
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for async BGP neighbor routes job %s", jobID)
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled waiting for async BGP neighbor routes job %s", jobID)
		case <-ticker.C:
			result, err := svc.GetAsyncBGPNeighborRoutes(ctx, mcrUID, jobID)
			if err != nil {
				return nil, err
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
