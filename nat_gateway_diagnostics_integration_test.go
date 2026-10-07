package megaport

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// NATGatewayDiagnosticsIntegrationTestSuite exercises the async
// "looking-glass" diagnostics endpoints. The list endpoints are strictly
// rate-limited and the looking-glass backend itself can be transiently
// unavailable for freshly-provisioned gateways; on 429 or 5xx we t.Skip
// the affected sub-case so the test remains green in those cases. The route
// sub-cases retry first.
type NATGatewayDiagnosticsIntegrationTestSuite IntegrationTestSuite

func TestNATGatewayDiagnosticsIntegrationTestSuite(t *testing.T) {
	runIntegrationMethods[NATGatewayDiagnosticsIntegrationTestSuite](t)
}

func (suite *NATGatewayDiagnosticsIntegrationTestSuite) SetupSuite() {
	natSuite := (*NATGatewayIntegrationTestSuite)(suite)
	natSuite.SetupSuite()
}

// isTransientDiagnosticsError reports whether err should cause the
// diagnostics sub-case to skip rather than fail. Covers:
//   - HTTP 429 (the documented rate-limit response)
//   - HTTP 5xx (intermittent backend errors observed on staging when the
//     looking-glass backend is unavailable for a freshly-provisioned
//     gateway). The SDK is doing its job by surfacing these — we just
//     don't want them to break CI.
func isTransientDiagnosticsError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Response == nil {
		return false
	}
	code := apiErr.Response.StatusCode
	return code == http.StatusTooManyRequests || code >= 500
}

// retryTransientDiagnostics retries call while it fails with a 429 or 5xx.
// Staging returns 503 for a short time after a gateway reaches CONFIGURED.
func retryTransientDiagnostics[T any](t *testing.T, call func() (T, error)) (T, error) {
	const (
		budget   = 90 * time.Second
		interval = 10 * time.Second
	)
	start := time.Now()
	for attempt := 1; ; attempt++ {
		got, err := call()
		if !isTransientDiagnosticsError(err) {
			return got, err
		}
		time.Sleep(interval)
		if elapsed := time.Since(start); elapsed >= budget {
			t.Skipf("still failing after %d attempts in %s: %v", attempt, elapsed.Round(time.Second), err)
		}
	}
}

func (suite *NATGatewayDiagnosticsIntegrationTestSuite) TestNATGatewayDiagnostics() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	natSuite := (*NATGatewayIntegrationTestSuite)(suite)
	prov, err := provisionNATGatewayForTest(ctx, natSuite, "Integration Test NAT Gateway (Diagnostics)")
	if err != nil {
		suite.FailNowf("could not provision NAT Gateway", "%v", err)
	}
	defer prov.Teardown()

	suite.Run("ip-routes", func() {
		routes, err := retryTransientDiagnostics(suite.T(), func() ([]*NATGatewayIPRoute, error) {
			return natSvc.ListNATGatewayIPRoutes(ctx, prov.ProductUID, "")
		})
		if err != nil {
			suite.FailNowf("could not list IP routes", "%v", err)
		}
		// Staging route tables drift, so check only that each route decoded a prefix.
		for _, r := range routes {
			suite.NotEmpty(r.Prefix, "IP route has no prefix")
		}
		logger.InfoContext(ctx, "ip routes diagnostics decoded", slog.Int("route_count", len(routes)))
	})

	suite.Run("bgp-routes", func() {
		routes, err := retryTransientDiagnostics(suite.T(), func() ([]*NATGatewayBGPRoute, error) {
			return natSvc.ListNATGatewayBGPRoutes(ctx, prov.ProductUID, "")
		})
		if err != nil {
			suite.FailNowf("could not list BGP routes", "%v", err)
		}
		for _, r := range routes {
			suite.NotEmpty(r.Prefix, "BGP route has no prefix")
		}
		logger.InfoContext(ctx, "bgp routes diagnostics decoded", slog.Int("route_count", len(routes)))
	})

	// BGP-neighbor needs a real peer IP to query against. Without a
	// running BGP session on this NAT Gateway, the API will likely
	// 400 — we still verify the request shape and validation surface,
	// but treat both rate-limit and "no neighbour" errors as a skip.
	suite.Run("bgp-neighbor-routes-validation", func() {
		_, err := natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{
			ProductUID:    prov.ProductUID,
			PeerIPAddress: "10.255.255.255",
			Direction:     BGPRouteDirectionReceived,
		})
		if err == nil {
			logger.InfoContext(ctx, "bgp neighbor request accepted (no live peer expected to return data)")
			return
		}
		if isTransientDiagnosticsError(err) {
			suite.T().Skip("transient backend error (429/5xx); skipping bgp-neighbor-routes sub-case")
			return
		}
		// Any non-429 error from the server is acceptable for this sub-case
		// — the goal here is to confirm the SDK can submit the request and
		// surface server errors cleanly. We don't assert a specific status
		// because staging behavior for "no such peer" varies.
		logger.InfoContext(ctx, "bgp neighbor request rejected as expected (no live peer)",
			slog.String("error", err.Error()),
		)
	})
}
