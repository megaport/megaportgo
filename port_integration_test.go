package megaport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/suite"
)

const (
	TEST_LOCATION_ID_A = 19 // 	Interactive 437 Williamstown
)

type PortIntegrationTestSuite IntegrationTestSuite

func TestPortIntegrationTestSuite(t *testing.T) {
	if os.Getenv("CI") != "true" {
		suite.Run(t, new(PortIntegrationTestSuite))
	}
}

func (suite *PortIntegrationTestSuite) SetupSuite() {
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

func (suite *PortIntegrationTestSuite) SetupTest() {
	suite.client.Logger.Debug("logging in oauth")
	if accessKey == "" {
		suite.FailNow("MEGAPORT_ACCESS_KEY environment variable not set.")
	}

	if secretKey == "" {
		suite.FailNow("MEGAPORT_SECRET_KEY environment variable not set.")
	}

	ctx := context.Background()
	token, loginErr := suite.client.AuthenticationService.LoginOauth(ctx, accessKey, secretKey)
	if loginErr != nil {
		suite.client.Logger.Error("login error", "error", loginErr.Error())
		suite.FailNowf("login error", "login error %v", loginErr)
	}

	// Session Token is not empty
	if !suite.NotEmpty(token) {
		suite.FailNow("empty token")
	}

	// SessionToken is a valid guid
	if !suite.NotNil(shared.IsGuid(token)) {
		suite.FailNowf("invalid guid for token", "invalid guid for token %v", token)
	}

	suite.client.SessionToken = token
}

// TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestSinglePort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	createRes, portErr := suite.testCreatePort(suite.client, ctx, types.SINGLE_PORT, *testLocation)
	suite.NoError(portErr)
	suite.Greater(len(createRes.PortOrderConfirmations), 0)

	portId := createRes.PortOrderConfirmations[0].TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(shared.IsGuid(portId)) {
		suite.FailNow("")
	}

	portCreated, err := suite.client.PortService.WaitForPortProvisioning(ctx, portId)

	if !suite.NoError(err) || !portCreated {
		suite.FailNow("could not create port")
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

	suite.testModifyPort(suite.client, ctx, portId, types.SINGLE_PORT)
	suite.testLockPort(suite.client, ctx, portId)
	suite.testCancelPort(suite.client, ctx, portId, types.SINGLE_PORT)
	suite.testDeletePort(suite.client, ctx, portId, types.SINGLE_PORT)

}

// TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestLAGPort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	orderRes, portErr := suite.testCreatePort(suite.client, ctx, types.LAG_PORT, *testLocation)
	suite.NoError(portErr)
	suite.Greater(len(orderRes.PortOrderConfirmations), 1)

	mainPortId := orderRes.PortOrderConfirmations[0].TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(shared.IsGuid(mainPortId)) {
		suite.FailNow("")
	}

	portCreated, err := suite.client.PortService.WaitForPortProvisioning(ctx, mainPortId)

	if !suite.NoError(err) || !portCreated {
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
		suite.client.Logger.Debug("Failed to find port we just created in ports list", "port_id", mainPortId)
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", mainPortId)
	}

	suite.testModifyPort(suite.client, ctx, mainPortId, types.LAG_PORT)
	suite.testCancelPort(suite.client, ctx, mainPortId, types.LAG_PORT)
	suite.testDeletePort(suite.client, ctx, mainPortId, types.LAG_PORT)
}

func (suite *PortIntegrationTestSuite) testCreatePort(c *Client, ctx context.Context, portType string, location types.Location) (*BuyPortResponse, error) {
	var portErr error
	var orderRes *BuyPortResponse

	suite.client.Logger.Debug("Buying Port", "port_type", portType)
	if portType == types.LAG_PORT {
		orderRes, portErr = c.PortService.BuyLAGPort(ctx, &BuyLAGPortRequest{
			Name:       "Buy Port (LAG) Test",
			Term:       1,
			PortSpeed:  10000,
			LocationId: location.ID,
			Market:     location.Market,
			LagCount:   2,
			IsPrivate:  true,
		})
	} else {
		orderRes, portErr = c.PortService.BuySinglePort(ctx, &BuySinglePortRequest{
			Name:       "Buy Port (Single) Test",
			Term:       1,
			PortSpeed:  10000,
			LocationId: location.ID,
			Market:     location.Market,
			IsPrivate:  true,
		})
	}
	if portErr != nil {
		return nil, portErr
	}
	return orderRes, nil
}

func (suite *PortIntegrationTestSuite) testModifyPort(c *Client, ctx context.Context, portId string, portType string) {
	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)

	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portType)

	suite.client.Logger.Debug("Modifying Port", "port_id", portId, "port_type", portType)
	_, modifyErr := c.PortService.ModifyPort(ctx, &ModifyPortRequest{
		PortID:                portId,
		Name:                  newPortName,
		CostCentre:            "",
		MarketplaceVisibility: portInfo.MarketplaceVisibility,
	})
	suite.NoError(modifyErr)

	secondGetPortInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)
	suite.EqualValues(newPortName, secondGetPortInfo.Name)
}

// PortScript tests the remaining lifecycle for a Port (not dependant on port-type), Go-Live, Modification,
// and Soft/Hard Deletes.
func (suite *PortIntegrationTestSuite) testCancelPort(c *Client, ctx context.Context, portId string, portType string) {
	// Soft Delete
	suite.client.Logger.Debug("Scheduling Port for deletion (30 days).", "port_id", portId, "port_type", portType)
	resp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
		PortID:    portId,
		DeleteNow: false,
	})
	suite.NoError(deleteErr)
	suite.True(resp.IsDeleting)

	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{PortID: portId})
	suite.NoError(err)
	suite.EqualValues(types.STATUS_CANCELLED, portInfo.ProvisioningStatus)

	suite.client.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
	restoreResp, restoreErr := c.PortService.RestorePort(ctx, &RestorePortRequest{PortID: portId})
	suite.NoError(restoreErr)
	suite.True(restoreResp.IsRestoring)

}

func (suite *PortIntegrationTestSuite) testDeletePort(c *Client, ctx context.Context, portId string, portType string) {
	// Hard Delete
	suite.client.Logger.Debug("Deleting Port now.", "port_type", portType, "port_id", portId)
	hardDeleteResp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
		PortID:    portId,
		DeleteNow: true,
	})
	suite.True(hardDeleteResp.IsDeleting)
	suite.NoError(deleteErr)

	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)

	suite.EqualValues(types.STATUS_DECOMMISSIONED, portInfo.ProvisioningStatus)
	suite.client.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
}

func (suite *PortIntegrationTestSuite) testLockPort(c *Client, ctx context.Context, portId string) {
	suite.client.Logger.Debug("Locking Port now.", "port_id", portId)
	lockResp, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
	suite.True(lockResp.IsLocking)
	suite.NoError(lockErr)

	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)
	suite.EqualValues(true, portInfo.Locked)

	suite.client.Logger.Debug("Test lock of an already locked port.", "port_id", portId)
	lockRes, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
	suite.Nil(lockRes)
	suite.Error(errors.New(mega_err.ERR_PORT_ALREADY_LOCKED), lockErr)

	suite.client.Logger.Debug("Unlocking Port now.", "port_id", portId)
	unlockResp, unlockErr := c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
	suite.True(unlockResp.IsUnlocking)
	suite.NoError(unlockErr)

	suite.client.Logger.Debug("Test unlocking of a port that doesn't have a lock.", "port_id", portId)
	unlockResp, unlockErr = c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
	suite.Nil(unlockResp)
	suite.Error(errors.New(mega_err.ERR_PORT_NOT_LOCKED), unlockErr)
}
