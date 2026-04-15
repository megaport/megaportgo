package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	TEST_NAT_GATEWAY_LOCATION_MARKET = "AU"
)

// NATGatewayIntegrationTestSuite is the integration test suite for the NAT Gateway service.
type NATGatewayIntegrationTestSuite IntegrationTestSuite

func TestNATGatewayIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(NATGatewayIntegrationTestSuite))
	}
}

func (suite *NATGatewayIntegrationTestSuite) SetupSuite() {
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

// TestNATGatewayLifecycle tests the full CRUD lifecycle of a NAT Gateway.
func (suite *NATGatewayIntegrationTestSuite) TestNATGatewayLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	// Step 1: List available sessions to pick a valid speed/session count.
	logger.DebugContext(ctx, "Listing NAT Gateway sessions.")
	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	if err != nil {
		suite.FailNowf("could not list sessions", "could not list NAT Gateway sessions: %v", err)
	}
	suite.NotEmpty(sessions, "expected at least one session configuration")

	testSpeed := sessions[0].SpeedMbps
	testSessionCount := sessions[0].SessionCount[0]
	logger.DebugContext(ctx, "Selected session config",
		slog.Int("speed", testSpeed),
		slog.Int("session_count", testSessionCount),
	)

	// Step 2: Pick a location.
	testLocation, locErr := GetRandomLocation(ctx, suite.client.LocationService, TEST_NAT_GATEWAY_LOCATION_MARKET)
	if locErr != nil {
		suite.FailNowf("could not get random location", "could not get random location: %v", locErr)
	}
	suite.NotNil(testLocation)
	logger.DebugContext(ctx, "Test location determined", slog.String("location", testLocation.Name), slog.Int("location_id", testLocation.ID))

	// Step 3: Create a NAT Gateway (stays in NEW status, not provisioned).
	logger.DebugContext(ctx, "Creating NAT Gateway.")
	createReq := &CreateNATGatewayRequest{
		AutoRenewTerm: true,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: false,
			DiversityZone:      "red",
			SessionCount:       testSessionCount,
		},
		LocationID:  testLocation.ID,
		ProductName: "Integration Test NAT Gateway",
		Speed:       testSpeed,
		Term:        1,
	}
	gw, err := natSvc.CreateNATGateway(ctx, createReq)
	if err != nil {
		suite.FailNowf("could not create NAT Gateway", "could not create NAT Gateway: %v", err)
	}
	suite.NotEmpty(gw.ProductUID)
	suite.Equal("Integration Test NAT Gateway", gw.ProductName)
	suite.Equal(testSpeed, gw.Speed)
	suite.Equal(1, gw.Term)
	logger.DebugContext(ctx, "NAT Gateway created", slog.String("product_uid", gw.ProductUID), slog.String("provisioning_status", gw.ProvisioningStatus))

	productUID := gw.ProductUID

	// Step 4: Get the NAT Gateway by UID.
	logger.DebugContext(ctx, "Retrieving NAT Gateway by UID.")
	fetched, err := natSvc.GetNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not get NAT Gateway", "could not get NAT Gateway: %v", err)
	}
	suite.Equal(productUID, fetched.ProductUID)
	suite.Equal("Integration Test NAT Gateway", fetched.ProductName)
	suite.Equal(testLocation.ID, fetched.LocationID)

	// Step 5: List NAT Gateways and verify ours appears.
	logger.DebugContext(ctx, "Listing NAT Gateways.")
	gateways, err := natSvc.ListNATGateways(ctx)
	if err != nil {
		suite.FailNowf("could not list NAT Gateways", "could not list NAT Gateways: %v", err)
	}
	found := false
	for _, g := range gateways {
		if g.ProductUID == productUID {
			found = true
			break
		}
	}
	suite.True(found, "created NAT Gateway not found in list")

	// Step 6: Update the NAT Gateway.
	logger.DebugContext(ctx, "Updating NAT Gateway.")
	updated, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID:    productUID,
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: true,
			DiversityZone:      "red",
			SessionCount:       testSessionCount,
		},
		LocationID:  testLocation.ID,
		ProductName: "Integration Test NAT Gateway [Updated]",
		Speed:       testSpeed,
		Term:        1,
	})
	if err != nil {
		suite.FailNowf("could not update NAT Gateway", "could not update NAT Gateway: %v", err)
	}
	suite.Equal("Integration Test NAT Gateway [Updated]", updated.ProductName)
	suite.False(updated.AutoRenewTerm)
	logger.DebugContext(ctx, "NAT Gateway updated", slog.String("product_name", updated.ProductName))

	// Step 7: Delete the NAT Gateway (allowed while provisioningStatus is NEW).
	logger.DebugContext(ctx, "Deleting NAT Gateway.")
	err = natSvc.DeleteNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not delete NAT Gateway", "could not delete NAT Gateway: %v", err)
	}
	logger.DebugContext(ctx, "NAT Gateway deleted", slog.String("product_uid", productUID))
}
