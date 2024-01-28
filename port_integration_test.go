package megaport

import (
	"context"
	"errors"
	"fmt"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
)

const (
	TEST_LOCATION_ID_A = 19 // 	Interactive 437 Williamstown
)

// TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *IntegrationTestSuite) TestSinglePort() {
	ctx := context.Background()

	testLocation, err := megaportClient.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := megaportClient.PortService.ListPorts(ctx)
	suite.NoError(err)

	portConfirmation, portErr := suite.testCreatePort(megaportClient, ctx, types.SINGLE_PORT, *testLocation)
	suite.NoError(portErr)

	portId := portConfirmation.TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(shared.IsGuid(portId)) {
		suite.FailNow("")
	}

	portCreated, err := megaportClient.PortService.WaitForPortProvisioning(ctx, portId)

	if !suite.NoError(err) || !portCreated {
		suite.FailNow("could not create port")
	}

	portsListPostCreate, err := megaportClient.PortService.ListPorts(ctx)
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

	suite.testModifyPort(megaportClient, ctx, portId, types.SINGLE_PORT)
	suite.testLockPort(megaportClient, ctx, portId)
	suite.testCancelPort(megaportClient, ctx, portId, types.SINGLE_PORT)
	suite.testDeletePort(megaportClient, ctx, portId, types.SINGLE_PORT)

}

// TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *IntegrationTestSuite) TestLAGPort() {
	ctx := context.Background()

	testLocation, err := megaportClient.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := megaportClient.PortService.ListPorts(ctx)
	suite.NoError(err)

	portConfirmation, portErr := suite.testCreatePort(megaportClient, ctx, types.LAG_PORT, *testLocation)

	portId := portConfirmation.TechnicalServiceUID

	if !suite.NoError(portErr) && !suite.True(shared.IsGuid(portId)) {
		suite.FailNow("")
	}

	portCreated, err := megaportClient.PortService.WaitForPortProvisioning(ctx, portId)

	if !suite.NoError(err) || !portCreated {
		suite.FailNow("")
	}

	portsListPostCreate, err := megaportClient.PortService.ListPorts(ctx)
	suite.NoError(err)

	portIsActuallyNew := true
	for _, p := range portsListInitial {
		if p.UID == portId {
			portIsActuallyNew = false
		}
	}

	if !portIsActuallyNew {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", portId)
	}

	foundNewPort := false
	for _, p := range portsListPostCreate {
		if p.UID == portId {
			foundNewPort = true
		}
	}

	if !foundNewPort {
		megaportClient.Logger.Debug("Failed to find port we just created in ports list", "port_id", portId)
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", portId)
	}

	suite.testModifyPort(megaportClient, ctx, portId, types.LAG_PORT)
	suite.testCancelPort(megaportClient, ctx, portId, types.LAG_PORT)
}

func (suite *IntegrationTestSuite) testCreatePort(c *Client, ctx context.Context, portType string, location types.Location) (*types.PortOrderConfirmation, error) {
	var portConfirm *types.PortOrderConfirmation
	var portErr error

	megaportClient.Logger.Debug("Buying Port", "port_type", portType)
	if portType == types.LAG_PORT {
		portConfirm, portErr = c.PortService.BuyLAGPort(ctx, &BuyLAGPortRequest{
			Name:       "Buy Port (LAG) Test",
			Term:       1,
			PortSpeed:  10000,
			LocationId: location.ID,
			Market:     location.Market,
			LagCount:   2,
			IsPrivate:  true,
		})
	} else {
		portConfirm, portErr = c.PortService.BuySinglePort(ctx, &BuySinglePortRequest{
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
	megaportClient.Logger.Debug("Port Purchased", "port_confirmation", portConfirm)
	return portConfirm, portErr
}

func (suite *IntegrationTestSuite) testModifyPort(c *Client, ctx context.Context, portId string, portType string) {
	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)

	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portType)

	megaportClient.Logger.Debug("Modifying Port", "port_id", portId, "port_type", portType)
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
func (suite *IntegrationTestSuite) testCancelPort(c *Client, ctx context.Context, portId string, portType string) {
	// Soft Delete
	megaportClient.Logger.Debug("Scheduling Port for deletion (30 days).", "port_id", portId, "port_type", portType)
	resp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
		PortID:    portId,
		DeleteNow: false,
	})
	suite.NoError(deleteErr)
	suite.True(resp.IsDeleting)

	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{PortID: portId})
	suite.NoError(err)
	suite.EqualValues(types.STATUS_CANCELLED, portInfo.ProvisioningStatus)

	megaportClient.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
	restoreResp, restoreErr := c.PortService.RestorePort(ctx, &RestorePortRequest{PortID: portId})
	suite.NoError(restoreErr)
	suite.True(restoreResp.IsRestoring)

}

func (suite *IntegrationTestSuite) testDeletePort(c *Client, ctx context.Context, portId string, portType string) {
	// Hard Delete
	megaportClient.Logger.Debug("Deleting Port now.", "port_type", portType, "port_id", portId)
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
	megaportClient.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
}

func (suite *IntegrationTestSuite) testLockPort(c *Client, ctx context.Context, portId string) {
	megaportClient.Logger.Debug("Locking Port now.", "port_id", portId)
	lockResp, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
	suite.True(lockResp.IsLocking)
	suite.NoError(lockErr)

	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
		PortID: portId,
	})
	suite.NoError(err)
	suite.EqualValues(true, portInfo.Locked)

	megaportClient.Logger.Debug("Test lock of an already locked port.", "port_id", portId)
	lockRes, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
	suite.Nil(lockRes)
	suite.Error(errors.New(mega_err.ERR_PORT_ALREADY_LOCKED), lockErr)

	megaportClient.Logger.Debug("Unlocking Port now.", "port_id", portId)
	unlockResp, unlockErr := c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
	suite.True(unlockResp.IsUnlocking)
	suite.NoError(unlockErr)

	megaportClient.Logger.Debug("Test unlocking of a port that doesn't have a lock.", "port_id", portId)
	unlockResp, unlockErr = c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
	suite.Nil(unlockResp)
	suite.Error(errors.New(mega_err.ERR_PORT_NOT_LOCKED), unlockErr)
}
