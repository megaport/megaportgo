package megaport

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

var runMveIntegrationTests = flag.Bool("integration", false, "perform integration tests")

const (
	TEST_MVE_TEST_LOCATION_MARKET = "AU"
)

type MVEIntegrationTestSuite IntegrationTestSuite

func TestMVEIntegrationTestSuite(t *testing.T) {
	if *runMveIntegrationTests {
		suite.Run(t, new(MVEIntegrationTestSuite))
	}
}

func (suite *MVEIntegrationTestSuite) SetupSuite() {
	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

	httpClient := NewHttpClient()

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err = New(httpClient, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

func (suite *MVEIntegrationTestSuite) SetupTest() {
	suite.client.Logger.Debug("logging in oauth")
	if accessKey == "" {
		suite.FailNow("MEGAPORT_ACCESS_KEY environment variable not set.")
	}

	if secretKey == "" {
		suite.FailNow("MEGAPORT_SECRET_KEY environment variable not set.")
	}

	ctx := context.Background()
	loginResp, loginErr := suite.client.AuthenticationService.LoginOauth(ctx, &LoginOauthRequest{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if loginErr != nil {
		suite.client.Logger.Error("login error", slog.String("error", loginErr.Error()))
		suite.FailNowf("login error", "login error %v", loginErr)
	}

	// Session Token is not empty
	if !suite.NotEmpty(loginResp.Token) {
		suite.FailNow("empty token")
	}

	// SessionToken is a valid guid
	if !suite.NotNil(IsGuid(loginResp.Token)) {
		suite.FailNowf("invalid guid for token", "invalid guid for token %v", loginResp.Token)
	}

	suite.client.SessionToken = loginResp.Token
}

func readSSHPubKey() string {
	key, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa.pub")
	if err != nil {
		panic(err)
	}
	return string(key)
}

func (suite *MVEIntegrationTestSuite) TestC8KVAutoLifecycle() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	locSvc := suite.client.LocationService
	logger := suite.client.Logger

	logger.DebugContext(ctx, "Buying MVE")
	testLocation, err := locSvc.GetRandom(ctx, TEST_MVE_TEST_LOCATION_MARKET)
	if err != nil {
		suite.FailNowf("could not get location", "could not get location %v", err)
	}
	logger.DebugContext(ctx, "test location determined", slog.String("location", testLocation.Name))
	mveConfig := &CiscoConfig{
		Vendor: "cisco",
		ProductSize: "SMALL",
		ImageID: 42,
		AdminSSHPublicKey: readSSHPubKey(),
	}

	buyMVERes, err := mveSvc.BuyMVE(ctx, &BuyMVERequest{
		LocationID:   testLocation.ID,
		Name:         "MVE Test",
		Term:         12,
		VendorConfig: mveConfig,
		Vnics:        nil,
	})
	if err != nil {
		suite.FailNowf("error buying mve", "error buying mve %v", err)
	}
	mveUid := buyMVERes.MVEOrderConfirmations[0].TechnicalServiceUID
	if !IsGuid(mveUid) {
		suite.FailNowf("invalid mve uid", "invalid mve uid %s", mveUid)
	}

	logger.DebugContext(ctx, "MVE Purchased", slog.String("mve_id", mveUid))
	logger.DebugContext(ctx, "Waiting for MVE to provision", slog.String("mve_id", mveUid))

	_, err = mveSvc.WaitForMVEProvisioning(ctx, mveUid)
	if err != nil {
		suite.FailNowf("could not provision mve", "could not provision mve %v", err)
	}

	logger.InfoContext(ctx, "Deleting MVE now", slog.String("mve_id", mveUid))

	deleteRes, err := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{
		MVEID: mveUid,
	})
	if err != nil {
		suite.FailNowf("could not delete mve", "could not delete mve %v", err)
	}
	suite.True(deleteRes.IsDeleted)

	mveDetails, err := mveSvc.GetMVE(ctx, mveUid)
	if err != nil {
		suite.FailNowf("could not get mve", "could not get mve %v", err)
	}
	suite.EqualValues(STATUS_DECOMMISSIONED, mveDetails.ProvisioningStatus)

	logger.DebugContext(ctx, "MVE deleted", slog.String("provisioning_status", mveDetails.ProvisioningStatus))

}
