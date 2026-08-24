package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

// routeDiagnosticsTimeout bounds a single route diagnostics call so a stuck
// operation fails the test instead of running to the suite timeout.
const routeDiagnosticsTimeout = 2 * time.Minute

// MCRLookingGlassIntegrationTestSuite is the integration test suite for the MCR Looking Glass service.
type MCRLookingGlassIntegrationTestSuite IntegrationTestSuite

func TestMCRLookingGlassIntegrationTestSuite(t *testing.T) {
	runIntegrationMethods[MCRLookingGlassIntegrationTestSuite](t)
}

func (suite *MCRLookingGlassIntegrationTestSuite) SetupSuite() {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	megaportClient, err := New(nil, WithBaseURL(MEGAPORTURL), WithLogHandler(handler), WithCredentials(accessKey, secretKey))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	ctx := context.Background()
	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		suite.FailNowf("", "could not authorize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

// TestLookingGlassWithMCR tests the route diagnostics endpoints with a real MCR.
// This test creates an MCR, queries the routing tables, and then cleans up.
func (suite *MCRLookingGlassIntegrationTestSuite) TestLookingGlassWithMCR() {
	ctx := context.Background()
	logger := suite.client.Logger
	mcrSvc := suite.client.MCRService
	lgSvc := suite.client.MCRLookingGlassService

	// Get a random test location
	testLocation, locErr := GetRandomLocation(ctx, suite.client.LocationService, TEST_MCR_TEST_LOCATION_MARKET)
	if locErr != nil {
		suite.FailNowf("could not get random location", "could not get random location %v", locErr)
	}
	if !suite.NotNil(testLocation) {
		suite.FailNow("invalid test location")
	}

	logger.DebugContext(ctx, "Test location determined", slog.String("location", testLocation.Name))

	// Buy an MCR for testing
	mcrRes, err := mcrSvc.BuyMCR(ctx, &BuyMCRRequest{
		LocationID:       testLocation.ID,
		Name:             "Looking Glass Test MCR",
		Term:             1,
		PortSpeed:        1000,
		MCRAsn:           65000,
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
	})
	if err != nil {
		suite.FailNowf("could not buy mcr", "could not buy mcr %v", err)
	}

	mcrUID := mcrRes.TechnicalServiceUID
	if !IsGuid(mcrUID) {
		suite.FailNowf("invalid mcr id", "invalid mcr id %s", mcrUID)
	}

	logger.InfoContext(ctx, "MCR Purchased for Looking Glass test", slog.String("mcr_id", mcrUID))

	// Cleanup function to delete the MCR after the test
	defer func() {
		logger.InfoContext(ctx, "Cleaning up test MCR", slog.String("mcr_id", mcrUID))
		_, deleteErr := mcrSvc.DeleteMCR(ctx, &DeleteMCRRequest{
			MCRID:     mcrUID,
			DeleteNow: true,
		})
		if deleteErr != nil {
			logger.ErrorContext(ctx, "Failed to delete test MCR", slog.String("error", deleteErr.Error()))
		}
	}()

	routeCtx, cancel := context.WithTimeout(ctx, routeDiagnosticsTimeout)
	defer cancel()

	// Test ListIPRoutes
	logger.DebugContext(ctx, "Testing ListIPRoutes")
	routes, err := lgSvc.ListIPRoutes(routeCtx, mcrUID)
	suite.NoError(err, "ListIPRoutes should not return error")
	// A newly provisioned MCR should have at least connected routes
	logger.InfoContext(ctx, "ListIPRoutes result", slog.Int("route_count", len(routes)))

	// Test ListIPRoutesWithFilter, restricted to one prefix
	logger.DebugContext(ctx, "Testing ListIPRoutesWithFilter")
	filteredRoutes, err := lgSvc.ListIPRoutesWithFilter(routeCtx, &ListIPRoutesRequest{
		MCRID:    mcrUID,
		IPFilter: "0.0.0.0/0",
	})
	suite.NoError(err, "ListIPRoutesWithFilter should not return error")
	logger.InfoContext(ctx, "Filtered routes", slog.Int("route_count", len(filteredRoutes)))

	// Test ListBGPRoutes - likely empty for a new MCR without VXCs
	logger.DebugContext(ctx, "Testing ListBGPRoutes")
	bgpRoutes, err := lgSvc.ListBGPRoutes(routeCtx, mcrUID)
	suite.NoError(err, "ListBGPRoutes should not return error")
	logger.InfoContext(ctx, "BGP routes", slog.Int("route_count", len(bgpRoutes)))

	logger.InfoContext(ctx, "Looking Glass integration test completed successfully")
}

// TestLookingGlassWithExistingMCR tests route diagnostics against an existing
// MCR when TEST_MCR_UID is set. Set TEST_MCR_BGP_PEER_IP as well to also
// exercise the BGP neighbor routes endpoint, which needs a live peer.
func (suite *MCRLookingGlassIntegrationTestSuite) TestLookingGlassWithExistingMCR() {
	mcrUID := os.Getenv("TEST_MCR_UID")
	if mcrUID == "" {
		suite.T().Skip("TEST_MCR_UID not set, skipping test with existing MCR")
		return
	}

	ctx := context.Background()
	logger := suite.client.Logger
	lgSvc := suite.client.MCRLookingGlassService

	logger.InfoContext(ctx, "Testing Looking Glass with existing MCR", slog.String("mcr_id", mcrUID))

	routeCtx, cancel := context.WithTimeout(ctx, routeDiagnosticsTimeout)
	defer cancel()

	// Test ListIPRoutes
	routes, err := lgSvc.ListIPRoutes(routeCtx, mcrUID)
	suite.NoError(err, "ListIPRoutes should not return error")
	logger.InfoContext(ctx, "ListIPRoutes result", slog.Int("route_count", len(routes)))

	// Log some route details for debugging
	for i, route := range routes {
		if i >= 5 {
			logger.DebugContext(ctx, "... and more routes")
			break
		}
		logger.DebugContext(ctx, "Route",
			slog.String("prefix", route.Prefix),
			slog.String("next_hop", route.NextHop.IP),
			slog.String("protocol", route.Protocol),
		)
	}

	// Test ListBGPRoutes
	bgpRoutes, err := lgSvc.ListBGPRoutes(routeCtx, mcrUID)
	suite.NoError(err, "ListBGPRoutes should not return error")
	logger.InfoContext(ctx, "BGP routes", slog.Int("route_count", len(bgpRoutes)))

	// The neighbor endpoint rejects a peer IP the MCR does not peer with, so it
	// needs an MCR with a live BGP session named by the caller.
	peerIP := os.Getenv("TEST_MCR_BGP_PEER_IP")
	if peerIP == "" {
		logger.InfoContext(ctx, "TEST_MCR_BGP_PEER_IP not set, skipping BGP neighbor routes")
		return
	}

	receivedRoutes, err := lgSvc.ListBGPNeighborRoutes(routeCtx, &ListBGPNeighborRoutesRequest{
		MCRID:         mcrUID,
		PeerIPAddress: peerIP,
		Direction:     BGPRouteDirectionReceived,
	})
	suite.NoError(err, "ListBGPNeighborRoutes (received) should not return error")
	logger.InfoContext(ctx, "Received routes from neighbor", slog.Int("route_count", len(receivedRoutes)))

	advertisedRoutes, err := lgSvc.ListBGPNeighborRoutes(routeCtx, &ListBGPNeighborRoutesRequest{
		MCRID:         mcrUID,
		PeerIPAddress: peerIP,
		Direction:     BGPRouteDirectionAdvertised,
	})
	suite.NoError(err, "ListBGPNeighborRoutes (advertised) should not return error")
	logger.InfoContext(ctx, "Advertised routes to neighbor", slog.Int("route_count", len(advertisedRoutes)))

	logger.InfoContext(ctx, "Looking Glass test with existing MCR completed")
}
