package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Diagnostics validation errors.
var (
	ErrNATGatewayDiagnosticsPeerIPRequired   = errors.New("BGP neighbor diagnostics require a peer IP address")
	ErrNATGatewayDiagnosticsDirectionInvalid = errors.New("BGP neighbor diagnostics require a direction (RECEIVED or ADVERTISED)")
	ErrNATGatewayDiagnosticsOperationEmpty   = errors.New("operation ID is required")
	ErrNATGatewayDiagnosticsTimeout          = errors.New("timed out waiting for diagnostics operation to complete")
)

// Default polling cadence used by the convenience diagnostics wrappers:
// an initial delay before the first poll, the interval between polls, and
// the overall timeout for the operation.
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
// operation returns a non-empty result, the SDK-managed
// diagnosticsPollTimeout elapses, or the caller's context is cancelled.
// Empty responses are treated as "still processing".
func (svc *NATGatewayServiceOp) pollDiagnosticsRoutes(ctx context.Context, productUID, operationID string) ([]*NATGatewayRoute, error) {
	pollCtx, cancel := context.WithTimeout(ctx, diagnosticsPollTimeout)
	defer cancel()
	// pollDoneErr returns ctx.Err() when the caller's context is the one that
	// fired (cancellation or caller-imposed deadline) and
	// ErrNATGatewayDiagnosticsTimeout when the SDK-managed
	// diagnosticsPollTimeout is what elapsed. This lets callers tell
	// "my deadline hit" from "the diagnostics op never completed".
	pollDoneErr := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return ErrNATGatewayDiagnosticsTimeout
	}
	select {
	case <-pollCtx.Done():
		return nil, pollDoneErr()
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
			return nil, pollDoneErr()
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
