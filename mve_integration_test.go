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
	TEST_MVE_TEST_LOCATION_MARKET = "AU"
)

// MVEIntegrationTestSuite is the integration test suite for the MVE service
type MVEIntegrationTestSuite IntegrationTestSuite

func TestMVEIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(MVEIntegrationTestSuite))
	}
}

func (suite *MVEIntegrationTestSuite) SetupSuite() {
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

// TestArubaMVE tests the lifecycle of an Aruba SD-WAN MVE
func (suite *MVEIntegrationTestSuite) TestArubaMVE() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	locSvc := suite.client.LocationService
	logger := suite.client.Logger

	logger.DebugContext(ctx, "Buying MVE")
	testLocation, err := GetRandomLocation(ctx, locSvc, TEST_MVE_TEST_LOCATION_MARKET)
	if err != nil {
		suite.FailNowf("could not get location", "could not get location %v", err)
	}
	logger.DebugContext(ctx, "test location determined", slog.String("location", testLocation.Name))
	mveConfig := &ArubaConfig{
		Vendor:      "aruba",
		ProductSize: "MEDIUM",
		ImageID:     23,
		AccountName: "test",
		AccountKey:  "test",
		SystemTag:   "test",
	}

	buyMVERes, err := mveSvc.BuyMVE(ctx, &BuyMVERequest{
		LocationID:       testLocation.ID,
		Name:             "MVE Test",
		Term:             12,
		VendorConfig:     mveConfig,
		Vnics:            nil,
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
		DiversityZone:    "red",
		ResourceTags:     testResourceTags,
	})
	if err != nil {
		suite.FailNowf("error buying mve", "error buying mve %v", err)
	}
	mveUid := buyMVERes.TechnicalServiceUID
	if !IsGuid(mveUid) {
		suite.FailNowf("invalid mve uid", "invalid mve uid %s", mveUid)
	}

	logger.DebugContext(ctx, "MVE Purchased", slog.String("mve_id", mveUid))

	mveDetails, err := mveSvc.GetMVE(ctx, mveUid)
	if err != nil {
		suite.FailNowf("could not get mve", "could not get mve %v", err)
	}
	suite.Equal(mveDetails.ResourceTags, testResourceTags)

	logger.InfoContext(ctx, "Deleting MVE now", slog.String("mve_id", mveUid))

	deleteRes, err := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{
		MVEID: mveUid,
	})
	if err != nil {
		suite.FailNowf("could not delete mve", "could not delete mve %v", err)
	}
	suite.True(deleteRes.IsDeleted)

	mveDetails, err = mveSvc.GetMVE(ctx, mveUid)
	if err != nil {
		suite.FailNowf("could not get mve", "could not get mve %v", err)
	}
	suite.EqualValues(STATUS_DECOMMISSIONED, mveDetails.ProvisioningStatus)

	logger.DebugContext(ctx, "MVE deleted", slog.String("provisioning_status", mveDetails.ProvisioningStatus))
}
