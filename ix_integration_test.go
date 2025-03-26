package megaport

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	TEST_IX_LOCATION_MARKET = "AU"
)

// IXIntegrationTestSuite is the integration test suite for the IX service
type IXIntegrationTestSuite IntegrationTestSuite

func TestIXIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(IXIntegrationTestSuite))
	}
}

func (suite *IXIntegrationTestSuite) SetupSuite() {
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

// TestIXLifecycle tests the full lifecycle of an IX
func (suite *IXIntegrationTestSuite) TestIXLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	ixSvc := suite.client.IXService
	portSvc := suite.client.PortService
	locSvc := suite.client.LocationService

	// First, we need a port to attach the IX to
	logger.InfoContext(ctx, "Finding a suitable location for the port")
	testLocation, locErr := GetRandomLocation(ctx, locSvc, TEST_IX_LOCATION_MARKET)
	if locErr != nil {
		suite.FailNowf("could not get random location", "could not get random location %v", locErr)
	}
	if !suite.NotNil(testLocation) {
		suite.FailNow("invalid test location")
	}

	logger.InfoContext(ctx, "Test location determined", slog.String("location", testLocation.Name))

	// Now create the port
	logger.InfoContext(ctx, "Buying port for IX attachment")
	portRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "IX Test Port",
		Term:                  1,
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Market:                TEST_IX_LOCATION_MARKET,
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("error buying port", "error buying port %v", portErr)
	}

	portID := portRes.TechnicalServiceUIDs[0]
	if !IsGuid(portID) {
		suite.FailNowf("invalid port id", "invalid port id %s", portID)
	}

	logger.InfoContext(ctx, "Port purchased", slog.String("port_id", portID))

	// Now create an IX attached to the port
	logger.InfoContext(ctx, "Buying IX")
	ixRes, ixErr := ixSvc.BuyIX(ctx, &BuyIXRequest{
		ProductUID:         portID,
		Name:               "Test IX",
		NetworkServiceType: "Sydney IX", // Adjust based on location
		ASN:                12345,
		MACAddress:         "AA:BB:CC:DD:EE:FF",
		RateLimit:          500,
		VLAN:               2001,
		Shutdown:           false,
		WaitForProvision:   true,
		WaitForTime:        10 * time.Minute,
	})

	if ixErr != nil {
		suite.FailNowf("error buying ix", "error buying ix %v", ixErr)
	}

	ixID := ixRes.TechnicalServiceUID
	if !IsGuid(ixID) {
		suite.FailNowf("invalid ix id", "invalid ix id %s", ixID)
	}

	logger.InfoContext(ctx, "IX purchased", slog.String("ix_id", ixID))

	// Get the IX details to verify it was created properly
	ix, getErr := ixSvc.GetIX(ctx, ixID)
	if getErr != nil {
		suite.FailNowf("could not get ix", "could not get ix %v", getErr)
	}

	suite.Equal("Test IX", ix.ProductName)
	suite.Equal(500, ix.RateLimit)
	suite.Equal(2001, ix.VLAN)
	// Use strings.EqualFold for case-insensitive comparison of MAC addresses
	suite.True(strings.EqualFold("AA:BB:CC:DD:EE:FF", ix.MACAddress),
		"MAC addresses should match case-insensitively. Expected: AA:BB:CC:DD:EE:FF, Got: %s", ix.MACAddress)
	suite.Equal(12345, ix.ASN)

	// Test updating the IX
	logger.InfoContext(ctx, "Updating IX")
	newName := "Updated Test IX"
	newRateLimit := 750
	newVLAN := 2002

	_, updateErr := ixSvc.UpdateIX(ctx, ixID, &UpdateIXRequest{
		Name:          &newName,
		RateLimit:     &newRateLimit,
		VLAN:          &newVLAN,
		WaitForUpdate: true,
		WaitForTime:   10 * time.Minute,
	})

	if updateErr != nil {
		suite.FailNowf("could not update ix", "could not update ix %v", updateErr)
	}

	updatedIX, err := ixSvc.GetIX(ctx, ixID)
	if err != nil {
		suite.FailNowf("could not get ix", "could not get ix %v", err)
	}

	// Verify the update was successful
	suite.Equal(newName, updatedIX.ProductName)
	suite.Equal(newRateLimit, updatedIX.RateLimit)
	suite.Equal(newVLAN, updatedIX.VLAN)

	// Double check by getting the IX details again
	ix, getErr = ixSvc.GetIX(ctx, ixID)
	if getErr != nil {
		suite.FailNowf("could not get ix", "could not get ix %v", getErr)
	}

	suite.Equal(newName, ix.ProductName)
	suite.Equal(newRateLimit, ix.RateLimit)
	suite.Equal(newVLAN, ix.VLAN)

	// Testing IX Cancel (soft delete)
	logger.InfoContext(ctx, "Scheduling IX for deletion (30 days)", slog.String("ix_id", ixID))

	deleteErr := ixSvc.DeleteIX(ctx, ixID, &DeleteIXRequest{
		DeleteNow: false,
	})
	if deleteErr != nil {
		suite.FailNowf("could not schedule ix for deletion", "could not schedule ix for deletion %v", deleteErr)
	}

	ix, getErr = ixSvc.GetIX(ctx, ixID)
	if getErr != nil {
		suite.FailNowf("could not get ix", "could not get ix %v", getErr)
	}
	suite.Equal(STATUS_CANCELLED, ix.ProvisioningStatus)
	logger.InfoContext(ctx, "IX scheduled for cancellation", slog.String("status", ix.ProvisioningStatus))

	// Hard delete the IX
	logger.InfoContext(ctx, "Deleting IX now", slog.String("ix_id", ixID))
	deleteErr = ixSvc.DeleteIX(ctx, ixID, &DeleteIXRequest{
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("could not delete ix", "could not delete ix %v", deleteErr)
	}

	// Verify IX was deleted
	ix, getErr = ixSvc.GetIX(ctx, ixID)
	if getErr != nil {
		suite.FailNowf("could not get ix", "could not get ix %v", getErr)
	}
	suite.Equal(STATUS_DECOMMISSIONED, ix.ProvisioningStatus)
	logger.InfoContext(ctx, "IX deleted", slog.String("status", ix.ProvisioningStatus))

	// Clean up by deleting the port
	logger.InfoContext(ctx, "Deleting port", slog.String("port_id", portID))
	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    portID,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("could not delete port", "could not delete port %v", deleteErr)
	}

	// Verify port was deleted
	port, getErr := portSvc.GetPort(ctx, portID)
	if getErr != nil {
		suite.FailNowf("could not get port", "could not get port %v", getErr)
	}
	suite.Equal(STATUS_DECOMMISSIONED, port.ProvisioningStatus)
	logger.InfoContext(ctx, "Port deleted", slog.String("status", port.ProvisioningStatus))
}
