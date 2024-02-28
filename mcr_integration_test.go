package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	TEST_MCR_TEST_LOCATION_MARKET = "AU"
)

// MCRIntegrationTestSuite is the integration test suite for the MCR service
type MCRIntegrationTestSuite IntegrationTestSuite

func TestMCRIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(MCRIntegrationTestSuite))
	}
}

func (suite *MCRIntegrationTestSuite) SetupSuite() {
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

// TestMCRLifecycle tests the full lifecycle of an MCR
func (suite *MCRIntegrationTestSuite) TestMCRLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	logger.DebugContext(ctx, "Buying MCR Port.")
	mcrSvc := suite.client.MCRService
	testLocation, locErr := GetRandomLocation(ctx, suite.client.LocationService, TEST_MCR_TEST_LOCATION_MARKET)
	if locErr != nil {
		suite.FailNowf("could not get random location", "could not get random location %v", locErr)
	}
	if !suite.NotNil(testLocation) {
		suite.FailNow("invalid test location")
	}

	logger.DebugContext(ctx, "Test location determined", slog.String("location", testLocation.Name))
	mcrRes, portErr := suite.client.MCRService.BuyMCR(ctx, &BuyMCRRequest{
		LocationID:       testLocation.ID,
		Name:             "Buy MCR",
		Term:             1,
		PortSpeed:        1000,
		MCRAsn:           0,
		DiversityZone:    "red",
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("error buying mcr", "error buying mcr %v", portErr)
	}
	mcrId := mcrRes.TechnicalServiceUID
	if !IsGuid(mcrId) {
		suite.FailNowf("invalid mcr id", "invalid mcr id %s", mcrId)
	}

	logger.DebugContext(ctx, "MCR Purchased", slog.String("mcr_id", mcrId))

	// Testing MCR Modify
	mcr, getErr := mcrSvc.GetMCR(ctx, mcrId)
	if getErr != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", getErr)
	}

	logger.DebugContext(ctx, "Modifying MCR.")
	newMCRName := "Buy MCR [Modified]"

	_, modifyErr := mcrSvc.ModifyMCR(ctx, &ModifyMCRRequest{
		MCRID:                 mcrId,
		Name:                  newMCRName,
		CostCentre:            "",
		MarketplaceVisibility: mcr.MarketplaceVisibility,
		WaitForUpdate:         true,
		WaitForTime:           5 * time.Minute,
	})
	if modifyErr != nil {
		suite.FailNowf("could not modify mcr", "could not modify mcr %v", modifyErr)
	}

	mcr, getErr = mcrSvc.GetMCR(ctx, mcrId)
	if getErr != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", getErr)
	}
	suite.EqualValues(newMCRName, mcr.Name)

	// Testing MCR Cancel
	logger.InfoContext(ctx, "Scheduling MCR for deletion (30 days).", slog.String("mcr_id", mcrId))

	// This is a soft Delete
	softDeleteRes, deleteErr := mcrSvc.DeleteMCR(ctx, &DeleteMCRRequest{
		MCRID:     mcrId,
		DeleteNow: false,
	})
	if deleteErr != nil {
		suite.FailNowf("could not soft delete mcr", "could not soft delete mcr %v", deleteErr)
	}
	suite.True(softDeleteRes.IsDeleting, true)

	mcrCancelInfo, getErr := mcrSvc.GetMCR(ctx, mcrId)
	if getErr != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", getErr)
	}
	suite.EqualValues(STATUS_CANCELLED, mcrCancelInfo.ProvisioningStatus)
	logger.DebugContext(ctx, "MCR Canceled", slog.String("provisioning_status", mcrCancelInfo.ProvisioningStatus))
	restoreRes, restoreErr := mcrSvc.RestoreMCR(ctx, mcrId)
	if restoreErr != nil {
		suite.FailNowf("could not restore mcr", "could not restore mcr %v", getErr)
	}
	suite.True(restoreRes.IsRestored)

	// Testing MCR Delete
	logger.Info("Deleting MCR now.")

	// This is a Hard Delete
	hardDeleteRes, deleteErr := mcrSvc.DeleteMCR(ctx, &DeleteMCRRequest{
		MCRID:     mcrId,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("could not delete mcr", "could not delete mcr %v", deleteErr)
	}
	suite.True(hardDeleteRes.IsDeleting)

	mcrDeleteInfo, getErr := mcrSvc.GetMCR(ctx, mcrId)
	if getErr != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", getErr)
	}
	suite.EqualValues(STATUS_DECOMMISSIONED, mcrDeleteInfo.ProvisioningStatus)
	logger.DebugContext(ctx, "mcr deleted", slog.String("provisioning_status", mcrDeleteInfo.ProvisioningStatus), slog.String("mcr_id", mcrId))
}

// TestPortSpeedValidation tests the port speed validation
func (suite *MCRIntegrationTestSuite) TestPortSpeedValidation() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	mcrSvc := suite.client.MCRService

	testLocation, locErr := locSvc.GetLocationByName(ctx, "Global Switch Sydney West")
	if locErr != nil {
		suite.FailNowf("could not get location", "could not get location %v", locErr)
	}
	if !suite.NotNil(testLocation) {
		suite.FailNow("invalid test location")
	}
	_, buyErr := mcrSvc.BuyMCR(ctx, &BuyMCRRequest{
		LocationID: testLocation.ID,
		Name:       "Test MCR",
		Term:       1,
		PortSpeed:  500,
		MCRAsn:     0,
	})
	suite.Equal(buyErr, ErrMCRInvalidPortSpeed)
}

// TestCreatePrefixFilterList tests the creation of a prefix filter list for an MCR.
func (suite *MCRIntegrationTestSuite) TestCreatePrefixFilterList() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	mcrSvc := suite.client.MCRService
	logger := suite.client.Logger

	logger.Info("Buying MCR Port.")
	testLocation, locErr := GetRandomLocation(ctx, locSvc, TEST_MCR_TEST_LOCATION_MARKET)
	if locErr != nil {
		suite.FailNowf("could not get location", "could not get location %v", locErr)
	}

	logger.InfoContext(ctx, "Test location determined", slog.String("location", testLocation.Name))
	mcrRes, portErr := mcrSvc.BuyMCR(ctx, &BuyMCRRequest{
		LocationID:       testLocation.ID,
		Name:             "Buy MCR",
		Term:             1,
		PortSpeed:        1000,
		MCRAsn:           0,
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("could not buy mcr", "could not buy mcr %v", portErr)
	}
	mcrId := mcrRes.TechnicalServiceUID

	if !IsGuid(mcrId) {
		suite.FailNowf("invalid mcr id", "invalid mcr id %s", mcrId)
	}

	logger.InfoContext(ctx, "MCR Purchased", slog.String("mcr_id", mcrId))

	logger.InfoContext(ctx, "Creating prefix filter list")

	prefixFilterEntries := []*MCRPrefixListEntry{
		{
			Action: "permit",
			Prefix: "10.0.1.0/24",
			Ge:     24,
			Le:     24,
		},
		{
			Action: "deny",
			Prefix: "10.0.2.0/24",
			Ge:     24,
			Le:     24,
		},
	}

	validatedPrefixFilterList := MCRPrefixFilterList{
		Description:   "Test Prefix Filter List",
		AddressFamily: "IPv4",
		Entries:       prefixFilterEntries,
	}

	_, prefixErr := mcrSvc.CreatePrefixFilterList(ctx, &CreateMCRPrefixFilterListRequest{
		MCRID:            mcrId,
		PrefixFilterList: validatedPrefixFilterList,
	})

	if prefixErr != nil {
		suite.FailNowf("could not create prefix filter list", "could not create prefix filter list %v", prefixErr)
	}

	logger.InfoContext(ctx, "Deleting MCR now.", slog.String("mcr_id", mcrId))
	hardDeleteRes, deleteErr := mcrSvc.DeleteMCR(ctx, &DeleteMCRRequest{
		MCRID:     mcrId,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("could not delete mcr", "could not delete mcr %v", deleteErr)
	}
	suite.True(hardDeleteRes.IsDeleting)

	mcrDeleteInfo, getErr := mcrSvc.GetMCR(ctx, mcrId)
	if getErr != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", getErr)
	}
	suite.EqualValues(STATUS_DECOMMISSIONED, mcrDeleteInfo.ProvisioningStatus)

	logger.DebugContext(ctx, "mcr deleted", slog.String("status", mcrDeleteInfo.ProvisioningStatus))
}
