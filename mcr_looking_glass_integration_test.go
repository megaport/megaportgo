package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// MCRLookingGlassIntegrationTestSuite is the integration test suite for the MCR Looking Glass service.
type MCRLookingGlassIntegrationTestSuite IntegrationTestSuite

func TestMCRLookingGlassIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(MCRLookingGlassIntegrationTestSuite))
	}
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

	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		suite.FailNowf("", "could not authorize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

// TestLookingGlassWithMCR tests the Looking Glass endpoints with a real MCR.
// This test creates an MCR, queries the Looking Glass, and then cleans up.
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

	// Test ListIPRoutes
	logger.DebugContext(ctx, "Testing ListIPRoutes")
	routes, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err, "ListIPRoutes should not return error")
	// A newly provisioned MCR should have at least connected routes
	logger.InfoContext(ctx, "ListIPRoutes result", slog.Int("route_count", len(routes)))

	// Test ListIPRoutesWithFilter - filter by protocol
	logger.DebugContext(ctx, "Testing ListIPRoutesWithFilter with CONNECTED protocol")
	connectedRoutes, err := lgSvc.ListIPRoutesWithFilter(ctx, &ListIPRoutesRequest{
		MCRID:    mcrUID,
		Protocol: RouteProtocolConnected,
	})
	suite.NoError(err, "ListIPRoutesWithFilter should not return error")
	logger.InfoContext(ctx, "Connected routes", slog.Int("route_count", len(connectedRoutes)))

	// Test ListBGPRoutes - likely empty for a new MCR without VXCs
	logger.DebugContext(ctx, "Testing ListBGPRoutes")
	bgpRoutes, err := lgSvc.ListBGPRoutes(ctx, mcrUID)
	suite.NoError(err, "ListBGPRoutes should not return error")
	logger.InfoContext(ctx, "BGP routes", slog.Int("route_count", len(bgpRoutes)))

	// Test ListBGPSessions - likely empty for a new MCR without VXCs
	logger.DebugContext(ctx, "Testing ListBGPSessions")
	bgpSessions, err := lgSvc.ListBGPSessions(ctx, mcrUID)
	suite.NoError(err, "ListBGPSessions should not return error")
	logger.InfoContext(ctx, "BGP sessions", slog.Int("session_count", len(bgpSessions)))

	// Test async IP routes retrieval
	logger.DebugContext(ctx, "Testing ListIPRoutesAsync")
	asyncJob, err := lgSvc.ListIPRoutesAsync(ctx, mcrUID)
	if err != nil {
		// Async might not be supported or enabled - log and continue
		logger.WarnContext(ctx, "ListIPRoutesAsync returned error (may not be supported)", slog.String("error", err.Error()))
	} else {
		suite.NotNil(asyncJob, "Async job should not be nil")
		suite.NotEmpty(asyncJob.JobID, "Async job ID should not be empty")
		logger.InfoContext(ctx, "Async job created", slog.String("job_id", asyncJob.JobID), slog.String("status", string(asyncJob.Status)))

		// Wait for async results
		asyncRoutes, err := lgSvc.WaitForAsyncIPRoutes(ctx, mcrUID, asyncJob.JobID, 2*time.Minute)
		if err != nil {
			logger.WarnContext(ctx, "WaitForAsyncIPRoutes returned error", slog.String("error", err.Error()))
		} else {
			logger.InfoContext(ctx, "Async routes received", slog.Int("route_count", len(asyncRoutes)))
		}
	}

	logger.InfoContext(ctx, "Looking Glass integration test completed successfully")
}

// TestLookingGlassWithExistingMCR tests Looking Glass with an existing MCR if MCR_UID env var is set.
// This allows testing against a real MCR with active BGP sessions.
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

	// Test ListIPRoutes
	routes, err := lgSvc.ListIPRoutes(ctx, mcrUID)
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
			slog.String("next_hop", route.NextHop),
			slog.String("protocol", string(route.Protocol)),
		)
	}

	// Test ListBGPRoutes
	bgpRoutes, err := lgSvc.ListBGPRoutes(ctx, mcrUID)
	suite.NoError(err, "ListBGPRoutes should not return error")
	logger.InfoContext(ctx, "BGP routes", slog.Int("route_count", len(bgpRoutes)))

	// Test ListBGPSessions
	bgpSessions, err := lgSvc.ListBGPSessions(ctx, mcrUID)
	suite.NoError(err, "ListBGPSessions should not return error")
	logger.InfoContext(ctx, "BGP sessions", slog.Int("session_count", len(bgpSessions)))

	// If we have BGP sessions, test neighbor routes
	if len(bgpSessions) > 0 {
		session := bgpSessions[0]
		logger.InfoContext(ctx, "Testing BGP neighbor routes for session",
			slog.String("session_id", session.SessionID),
			slog.String("neighbor", session.NeighborAddress),
			slog.Int("neighbor_asn", session.NeighborASN),
		)

		// Test received routes
		receivedRoutes, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
			MCRID:     mcrUID,
			SessionID: session.SessionID,
			Direction: LookingGlassRouteDirectionReceived,
		})
		suite.NoError(err, "ListBGPNeighborRoutes (received) should not return error")
		logger.InfoContext(ctx, "Received routes from neighbor", slog.Int("route_count", len(receivedRoutes)))

		// Test advertised routes
		advertisedRoutes, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
			MCRID:     mcrUID,
			SessionID: session.SessionID,
			Direction: LookingGlassRouteDirectionAdvertised,
		})
		suite.NoError(err, "ListBGPNeighborRoutes (advertised) should not return error")
		logger.InfoContext(ctx, "Advertised routes to neighbor", slog.Int("route_count", len(advertisedRoutes)))
	}

	logger.InfoContext(ctx, "Looking Glass test with existing MCR completed")
}
