package megaport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	TEST_LOCATION_ID_A = 19 // 	Interactive 437 Williamstown
)

type PortIntegrationTestSuite IntegrationTestSuite

func TestPortIntegrationTestSuite(t *testing.T) {
	if *runIntegrationTests {
		suite.Run(t, new(PortIntegrationTestSuite))
	}
}

func (suite *PortIntegrationTestSuite) SetupSuite() {
	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err = New(nil, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

func (suite *PortIntegrationTestSuite) SetupTest() {
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

// TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestSinglePort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	createRes, portErr := suite.testCreatePort(suite.client, ctx, SINGLE_PORT, *testLocation)
	suite.NoError(portErr)

	portId := createRes.TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(IsGuid(portId)) {
		suite.FailNow("")
	}

	portsListPostCreate, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	portIsActuallyNew := true
	for _, p := range portsListInitial {
		if p.UID == portId {
			portIsActuallyNew = false
		}
	}

	if !portIsActuallyNew {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %s", portId)
	}

	foundNewPort := false
	for _, p := range portsListPostCreate {
		if p.UID == portId {
			foundNewPort = true
		}
	}

	if !foundNewPort {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list: %v", portId)
	}

	suite.testModifyPort(suite.client, ctx, portId, SINGLE_PORT)
	suite.testLockPort(suite.client, ctx, portId)
	suite.testCancelPort(suite.client, ctx, portId, SINGLE_PORT)
	suite.testDeletePort(suite.client, ctx, portId, SINGLE_PORT)

}

// TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestLAGPort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	orderRes, portErr := suite.testCreatePort(suite.client, ctx, LAG_PORT, *testLocation)
	suite.NoError(portErr)

	mainPortId := orderRes.TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(IsGuid(mainPortId)) {
		suite.FailNow("")
	}

	portsListPostCreate, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	portIsActuallyNew := true
	for _, p := range portsListInitial {
		if p.UID == mainPortId {
			portIsActuallyNew = false
		}
	}

	if !portIsActuallyNew {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", mainPortId)
	}

	foundNewPort := false
	for _, p := range portsListPostCreate {
		if p.UID == mainPortId {
			foundNewPort = true
		}
	}

	if !foundNewPort {
		suite.client.Logger.Debug("Failed to find port we just created in ports list", slog.String("port_id", mainPortId))
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", mainPortId)
	}

	suite.testModifyPort(suite.client, ctx, mainPortId, LAG_PORT)
	suite.testCancelPort(suite.client, ctx, mainPortId, LAG_PORT)
	suite.testDeletePort(suite.client, ctx, mainPortId, LAG_PORT)
}

func (suite *PortIntegrationTestSuite) testCreatePort(c *Client, ctx context.Context, portType string, location Location) (*BuyPortResponse, error) {
	var portErr error
	var orderRes *BuyPortResponse

	suite.client.Logger.Debug("Buying Port", slog.String("port_type", portType))
	if portType == LAG_PORT {
		orderRes, portErr = c.PortService.BuyLAGPort(ctx, &BuyLAGPortRequest{
			Name:       "Buy Port (LAG) Test",
			Term:       1,
			PortSpeed:  10000,
			LocationId: location.ID,
			Market:     location.Market,
			LagCount:   2,
			IsPrivate:  true,
			WaitForProvision: true,
			WaitForTime: 5 * time.Minute,
		})
	} else {
		orderRes, portErr = c.PortService.BuySinglePort(ctx, &BuySinglePortRequest{
			Name:       "Buy Port (Single) Test",
			Term:       1,
			PortSpeed:  10000,
			LocationId: location.ID,
			Market:     location.Market,
			IsPrivate:  true,
			WaitForProvision: true,
			WaitForTime: 5 * time.Minute,
		})
	}
	if portErr != nil {
		suite.FailNowf("could not find port", "could not find port %v", portErr)
	}
	return orderRes, nil
}

func (suite *PortIntegrationTestSuite) testModifyPort(c *Client, ctx context.Context, portId string, portType string) {
	portInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}

	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portType)

	suite.client.Logger.Debug("Modifying Port", slog.String("port_id", portId), slog.String("port_type", portType))
	_, modifyErr := c.PortService.ModifyPort(ctx, &ModifyPortRequest{
		PortID:                portId,
		Name:                  newPortName,
		CostCentre:            "",
		MarketplaceVisibility: portInfo.MarketplaceVisibility,
		WaitForUpdate: true,
		WaitForTime: 5 * time.Minute,
	})
	if err != nil {
		suite.FailNowf("could not modify port", "could not modify port %v", modifyErr)
	}

	secondGetPortInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}
	suite.EqualValues(newPortName, secondGetPortInfo.Name)
}

// PortScript tests the remaining lifecycle for a Port (not dependant on port-type), Go-Live, Modification,
// and Soft/Hard Deletes.
func (suite *PortIntegrationTestSuite) testCancelPort(c *Client, ctx context.Context, portId string, portType string) {
	// Soft Delete
	suite.client.Logger.Debug("Scheduling Port for deletion (30 days).", slog.String("port_id", portId), slog.String("port_type", portType))
	resp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
		PortID:    portId,
		DeleteNow: false,
	})
	if deleteErr != nil {
		suite.FailNowf("could not cancel port", "could not cancel port %v", deleteErr)
	}
	suite.True(resp.IsDeleting)

	portInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}
	suite.EqualValues(STATUS_CANCELLED, portInfo.ProvisioningStatus)

	suite.client.Logger.Debug("port scheduled for cancellation", slog.String("status", portInfo.ProvisioningStatus), slog.String("port_id", portId))
	restoreResp, restoreErr := c.PortService.RestorePort(ctx, portId)
	if restoreErr != nil {
		suite.FailNowf("could not restore port", "could not restore port %v", restoreErr)
	}
	suite.True(restoreResp.IsRestored)

}

func (suite *PortIntegrationTestSuite) testDeletePort(c *Client, ctx context.Context, portId string, portType string) {
	// Hard Delete
	suite.client.Logger.Debug("Deleting Port now.", slog.String("port_type", portType), slog.String("port_id", portId))
	hardDeleteResp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
		PortID:    portId,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("could not delete port", "could not delete port %v", deleteErr)
	}
	suite.True(hardDeleteResp.IsDeleting)

	portInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}
	suite.EqualValues(STATUS_DECOMMISSIONED, portInfo.ProvisioningStatus)
	suite.client.Logger.Debug("port deleted", slog.String("status", portInfo.ProvisioningStatus), slog.String("port_id", portId))
}

func (suite *PortIntegrationTestSuite) testLockPort(c *Client, ctx context.Context, portId string) {
	suite.client.Logger.Debug("Locking Port now.", slog.String("port_id", portId))
	lockResp, lockErr := c.PortService.LockPort(ctx, portId)
	if lockErr != nil {
		suite.FailNowf("could not lock port", "could not lock port %v", lockErr)
	}
	suite.True(lockResp.IsLocking)

	portInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}
	suite.EqualValues(true, portInfo.Locked)

	suite.client.Logger.Debug("Test lock of an already locked port.", slog.String("port_id", portId))
	lockRes, lockErr := c.PortService.LockPort(ctx, portId)
	suite.Nil(lockRes)
	suite.Error(errors.New(ERR_PORT_ALREADY_LOCKED), lockErr)

	suite.client.Logger.Debug("Unlocking Port now.", slog.String("port_id", portId))
	unlockResp, unlockErr := c.PortService.UnlockPort(ctx, portId)
	if unlockErr != nil {
		suite.FailNowf("could not unlock port", "could not unlock port %v", unlockErr)
	}
	suite.True(unlockResp.IsUnlocking)

	suite.client.Logger.Debug("Test unlocking of a port that doesn't have a lock.", slog.String("port_id", portId))
	unlockResp, unlockErr = c.PortService.UnlockPort(ctx, portId)
	suite.Nil(unlockResp)
	suite.Error(errors.New(ERR_PORT_NOT_LOCKED), unlockErr)
}
