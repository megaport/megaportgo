package megaport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"time"
)

// natGatewayProvisionResult bundles the resources created by
// provisionNATGatewayForTest and its teardown function.
type natGatewayProvisionResult struct {
	ProductUID   string
	Speed        int
	SessionCount int
	LocationID   int
	ASN          int
	// Teardown deletes the gateway via whichever path matches its current
	// state (DESIGN hard-delete vs CANCEL_NOW). Always safe to defer; logs
	// and continues on best-effort failure.
	Teardown func()
}

// provisionNATGatewayForTest provisions a NAT Gateway and waits for it to
// reach CONFIGURED/LIVE. Used by the packet filter, prefix list, VXC and
// diagnostics integration tests so each test does not have to re-implement
// the buy-and-poll dance.
//
// The caller MUST defer result.Teardown() — otherwise the gateway leaks
// into the staging account and bills until manually cleaned up.
func provisionNATGatewayForTest(ctx context.Context, suite *NATGatewayIntegrationTestSuite, productName string) (*natGatewayProvisionResult, error) {
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list NAT Gateway sessions: %w", err)
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no NAT Gateway session configurations available")
	}
	minSession := sessions[0]
	for _, s := range sessions[1:] {
		if s.SpeedMbps < minSession.SpeedMbps {
			minSession = s
		}
	}
	testSpeed := minSession.SpeedMbps
	testSessionCount := minSession.SessionCount[0]

	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list locations: %w", err)
	}
	marketLocations, err := suite.client.LocationService.FilterLocationsByMarketCodeV3(ctx, TEST_NAT_GATEWAY_LOCATION_MARKET, locations)
	if err != nil {
		return nil, fmt.Errorf("could not filter locations by market: %w", err)
	}
	eligible := suite.client.LocationService.FilterLocationsByNATGatewaySpeedV3(ctx, testSpeed, marketLocations)
	if len(eligible) == 0 {
		return nil, fmt.Errorf("no location in market %q advertises NAT Gateway speed %d", TEST_NAT_GATEWAY_LOCATION_MARKET, testSpeed)
	}

	// Shuffle candidates so parallel test runs don't all race for the same site.
	//nolint:gosec // test-only shuffle; cryptographic randomness not required
	candidates := slices.Clone(eligible)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })

	const asn = 64512
	var (
		gw           *NATGateway
		testLocation *LocationV3
	)
	for _, loc := range candidates {
		var createErr error
		gw, createErr = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
			AutoRenewTerm: false,
			Config: NATGatewayNetworkConfig{
				ASN:                asn,
				BGPShutdownDefault: false,
				DiversityZone:      "red",
				SessionCount:       testSessionCount,
			},
			LocationID:  loc.ID,
			ProductName: productName,
			Speed:       testSpeed,
			Term:        1,
		})
		if createErr != nil {
			logger.WarnContext(ctx, "NAT Gateway create failed, trying next location",
				slog.Int("location_id", loc.ID),
				slog.String("error", createErr.Error()),
			)
			continue
		}
		testLocation = loc
		break
	}
	if gw == nil || testLocation == nil {
		return nil, fmt.Errorf("could not create NAT Gateway: all %d candidate locations failed", len(candidates))
	}
	productUID := gw.ProductUID
	logger.InfoContext(ctx, "NAT Gateway design created",
		slog.String("product_uid", productUID),
		slog.String("provisioning_status", gw.ProvisioningStatus),
	)

	teardown := makeNATGatewayTeardown(ctx, suite, productUID)

	if _, err := natSvc.ValidateNATGatewayOrder(ctx, productUID); err != nil {
		teardown()
		return nil, fmt.Errorf("could not validate NAT Gateway: %w", err)
	}
	if _, err := natSvc.BuyNATGateway(ctx, productUID); err != nil {
		teardown()
		return nil, fmt.Errorf("could not buy NAT Gateway: %w", err)
	}

	const (
		pollInterval = 10 * time.Second
		pollTimeout  = 15 * time.Minute
	)
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		fetched, getErr := natSvc.GetNATGateway(pollCtx, productUID)
		if getErr != nil {
			teardown()
			return nil, fmt.Errorf("could not poll NAT Gateway: %w", getErr)
		}
		if slices.Contains(SERVICE_STATE_READY, fetched.ProvisioningStatus) {
			logger.InfoContext(ctx, "NAT Gateway provisioned",
				slog.String("product_uid", productUID),
				slog.String("provisioning_status", fetched.ProvisioningStatus),
			)
			return &natGatewayProvisionResult{
				ProductUID:   productUID,
				Speed:        testSpeed,
				SessionCount: testSessionCount,
				LocationID:   testLocation.ID,
				ASN:          asn,
				Teardown:     teardown,
			}, nil
		}
		if fetched.ProvisioningStatus == STATUS_DECOMMISSIONED ||
			fetched.ProvisioningStatus == STATUS_CANCELLED {
			teardown()
			return nil, fmt.Errorf("NAT Gateway %s reached terminal state %s", productUID, fetched.ProvisioningStatus)
		}
		select {
		case <-pollCtx.Done():
			teardown()
			return nil, fmt.Errorf("timed out waiting for NAT Gateway %s to provision (last status %q)", productUID, fetched.ProvisioningStatus)
		case <-ticker.C:
		}
	}
}

// makeNATGatewayTeardown returns a best-effort cleanup function that
// dispatches to the right deletion path based on the gateway's current
// state. Modeled on the inline teardown in TestNATGatewayFullLifecycle.
func makeNATGatewayTeardown(ctx context.Context, suite *NATGatewayIntegrationTestSuite, productUID string) func() {
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	deleteDesign := func() error {
		path := fmt.Sprintf("/v3/products/nat_gateways/%s", url.PathEscape(productUID))
		req, err := suite.client.NewRequest(ctx, http.MethodDelete, path, nil)
		if err != nil {
			return err
		}
		resp, err := suite.client.Do(ctx, req, nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	cancelNow := func() error {
		_, err := suite.client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
			ProductID: productUID,
			DeleteNow: true,
		})
		return err
	}
	return func() {
		current, getErr := natSvc.GetNATGateway(ctx, productUID)
		if getErr != nil {
			logger.WarnContext(ctx, "NAT Gateway teardown: could not inspect state, attempting both paths",
				slog.String("product_uid", productUID),
				slog.String("error", getErr.Error()),
			)
			if err := deleteDesign(); err != nil {
				logger.WarnContext(ctx, "NAT Gateway teardown (DESIGN DELETE) best-effort failed", slog.String("error", err.Error()))
			}
			if err := cancelNow(); err != nil {
				logger.WarnContext(ctx, "NAT Gateway teardown (CANCEL_NOW) best-effort failed", slog.String("error", err.Error()))
			}
			return
		}
		if current.ProvisioningStatus == STATUS_DECOMMISSIONED ||
			current.ProvisioningStatus == STATUS_CANCELLED {
			logger.InfoContext(ctx, "NAT Gateway teardown skipped: already in terminal state",
				slog.String("product_uid", productUID),
				slog.String("provisioning_status", current.ProvisioningStatus),
			)
			return
		}
		var dErr error
		if current.ProvisioningStatus == STATUS_DESIGN {
			dErr = deleteDesign()
		} else {
			dErr = cancelNow()
		}
		if dErr != nil {
			logger.WarnContext(ctx, "NAT Gateway teardown failed",
				slog.String("product_uid", productUID),
				slog.String("provisioning_status", current.ProvisioningStatus),
				slog.String("error", dErr.Error()),
			)
		}
	}
}
