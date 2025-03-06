package megaport

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	TEST_LOCATION_ID_A = 19 // 	Interactive 437 Williamstown
)

// PortIntegrationTestSuite tests the Port Service.
type PortIntegrationTestSuite IntegrationTestSuite

func TestPortIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(PortIntegrationTestSuite))
	}
}

func (suite *PortIntegrationTestSuite) SetupSuite() {
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

// TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestSinglePort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	createRes, portErr := suite.testCreatePort(suite.client, ctx, 0, *testLocation)
	suite.NoError(portErr)

	portID := createRes.TechnicalServiceUIDs[0]

	if !suite.NoError(portErr) && !suite.True(IsGuid(portID)) {
		suite.FailNow("")
	}

	portsListPostCreate, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	portIsActuallyNew := true
	for _, p := range portsListInitial {
		if p.UID == portID {
			portIsActuallyNew = false
		}
	}

	if !portIsActuallyNew {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %s", portID)
	}

	foundNewPort := false
	for _, p := range portsListPostCreate {
		if p.UID == portID {
			foundNewPort = true
			tags, err := suite.client.PortService.ListPortResourceTags(ctx, p.UID)
			if err != nil {
				suite.FailNowf("could not list port resource tags", "could not list port resource tags %v", err)
			}
			suite.EqualValues(testResourceTags, tags)
			err = suite.client.PortService.UpdatePortResourceTags(ctx, p.UID, testUpdatedResourceTags)
			if err != nil {
				suite.FailNowf("could not update port resource tags", "could not update port resource tags %v", err)
			}
			tags, err = suite.client.PortService.ListPortResourceTags(ctx, p.UID)
			if err != nil {
				suite.FailNowf("could not list port resource tags", "could not list port resource tags %v", err)
			}
			suite.EqualValues(testUpdatedResourceTags, tags)
		}
	}

	if !foundNewPort {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list: %v", portID)
	}

	createdServiceKey, err := suite.testCreateServiceKey(suite.client, ctx, portID)
	if err != nil {
		suite.FailNowf("could not create service key", "could not create service key %v", err)
	}
	suite.testUpdateServiceKey(suite.client, ctx, createdServiceKey.Key, createdServiceKey.ProductID)

	suite.testModifyPort(suite.client, ctx, portID)
	suite.testLockPort(suite.client, ctx, portID)
	suite.testCancelPort(suite.client, ctx, portID)
	suite.testDeletePort(suite.client, ctx, portID)

}

// TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func (suite *PortIntegrationTestSuite) TestLAGPort() {
	ctx := context.Background()

	testLocation, err := suite.client.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)
	suite.NoError(err)

	portsListInitial, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	orderRes, portErr := suite.testCreatePort(suite.client, ctx, 2, *testLocation)
	suite.NoError(portErr)

	mainPortIDs := orderRes.TechnicalServiceUIDs

	if !suite.NoError(portErr) && !suite.True(IsGuid(mainPortIDs...)) {
		suite.FailNow("")
	}

	portsListPostCreate, err := suite.client.PortService.ListPorts(ctx)
	suite.NoError(err)

	portIsActuallyNew := true
	for _, p := range portsListInitial {
		if slices.Contains(mainPortIDs, p.UID) {
			portIsActuallyNew = false
		}
	}
	if !portIsActuallyNew {
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", mainPortIDs)
	}

	foundNewPort := false
	for _, p := range portsListPostCreate {
		if slices.Contains(mainPortIDs, p.UID) {
			foundNewPort = true
			if p.AggregationID != 0 {
				lagPort, err := suite.client.PortService.GetPort(ctx, p.UID)
				suite.NoError(err)
				suite.Equal(2, lagPort.LagCount)
				suite.ElementsMatch(mainPortIDs, lagPort.LagPortUIDs)
			}
		}
	}

	if !foundNewPort {
		suite.client.Logger.DebugContext(ctx, "Failed to find port we just created in ports list", slog.String("port_ids", mainPortIDs[0]))
		suite.FailNowf("Failed to find port we just created in ports list", "Failed to find port we just created in ports list %v", mainPortIDs)
	}

	key, err := suite.testCreateServiceKey(suite.client, ctx, mainPortIDs[0])
	if err != nil {
		suite.FailNowf("could not create service key", "could not create service key %v", err)
	}
	suite.testUpdateServiceKey(suite.client, ctx, key.Key, key.ProductID)
	suite.testModifyPort(suite.client, ctx, mainPortIDs[0])
	suite.testCancelPort(suite.client, ctx, mainPortIDs[0])
	suite.testDeletePort(suite.client, ctx, mainPortIDs[0])
}

func (suite *PortIntegrationTestSuite) testCreatePort(c *Client, ctx context.Context, lagCount int, location Location) (*BuyPortResponse, error) {
	suite.client.Logger.DebugContext(ctx, "Buying Port", slog.Int("lag_count", lagCount))
	orderRes, err := c.PortService.BuyPort(ctx, &BuyPortRequest{
		Name:                  "Buy Port (LAG) Test",
		Term:                  1,
		PortSpeed:             10000,
		LocationId:            location.ID,
		Market:                location.Market,
		LagCount:              lagCount,
		MarketPlaceVisibility: true,
		DiversityZone:         "red",
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
		ResourceTags:          testResourceTags,
	})

	if err != nil {
		return nil, err
	}
	return orderRes, nil
}

//nolint:unparam
func (suite *PortIntegrationTestSuite) testCreateServiceKey(c *Client, ctx context.Context, portId string) (*ServiceKey, error) {
	suite.client.Logger.DebugContext(ctx, "Creating Service Key", slog.String("port_id", portId))
	createReq := &CreateServiceKeyRequest{
		ProductUID:  portId,
		Description: "Test Service Key",
		Active:      true,
		SingleUse:   true,
		MaxSpeed:    500,
		PreApproved: true,
		VLAN:        3,
	}
	validFor := &ValidFor{
		StartTime: &Time{
			Time: time.Now(),
		},
		EndTime: &Time{
			Time: time.Now().Add(24 * time.Hour),
		},
	}
	createReq.ValidFor = validFor
	createRes, err := c.ServiceKeyService.CreateServiceKey(ctx, createReq)
	if err != nil {
		suite.FailNowf("could not create service key", "could not create service key %v", err)
	}
	suite.True(IsGuid(createRes.ServiceKeyUID))
	foundKey, err := c.ServiceKeyService.GetServiceKey(ctx, createRes.ServiceKeyUID)
	if err != nil {
		suite.FailNowf("could not get service key", "could not get service key %v", err)
	}
	suite.Equal(createRes.ServiceKeyUID, foundKey.Key)
	suite.Equal(foundKey.Description, createReq.Description)
	suite.Equal(foundKey.ProductUID, portId)
	suite.Equal(foundKey.Active, createReq.Active)
	suite.Equal(foundKey.SingleUse, createReq.SingleUse)
	suite.Equal(foundKey.MaxSpeed, createReq.MaxSpeed)
	suite.Equal(foundKey.PreApproved, createReq.PreApproved)
	suite.Equal(foundKey.VLAN, createReq.VLAN)

	createReq2 := &CreateServiceKeyRequest{
		ProductUID:  portId,
		Description: "Test Service Key 2",
		Active:      true,
		SingleUse:   true,
		MaxSpeed:    500,
		PreApproved: true,
		VLAN:        3,
	}
	validFor2 := &ValidFor{
		StartTime: &Time{
			Time: time.Now(),
		},
		EndTime: &Time{
			Time: time.Now().Add(24 * time.Hour),
		},
	}
	createReq2.ValidFor = validFor2
	_, err = c.ServiceKeyService.CreateServiceKey(ctx, createReq2)
	if err != nil {
		suite.FailNowf("could not create service key", "could not create service key %v", err)
	}
	listRes, err := c.ServiceKeyService.ListServiceKeys(ctx, &ListServiceKeysRequest{
		ProductUID: &portId,
	})
	if err != nil {
		suite.FailNowf("could not list service keys", "could not list service keys %v", err)
	}
	suite.Equal(len(listRes.ServiceKeys), 2)
	return foundKey, nil
}

func (suite *PortIntegrationTestSuite) testUpdateServiceKey(c *Client, ctx context.Context, key string, productID int) {
	suite.client.Logger.DebugContext(ctx, "Updating Service Key", slog.String("key", key))
	updateReq := &UpdateServiceKeyRequest{
		Key:       key,
		ProductID: productID,
		Active:    false,
		SingleUse: true,
	}
	validFor := &ValidFor{
		StartTime: &Time{
			Time: time.Now(),
		},
		EndTime: &Time{
			Time: time.Now().Add(24 * time.Hour),
		},
	}
	updateReq.ValidFor = validFor
	_, err := c.ServiceKeyService.UpdateServiceKey(ctx, updateReq)
	if err != nil {
		suite.FailNowf("could not update service key", "could not update service key %v", err)
	}
}

func (suite *PortIntegrationTestSuite) testModifyPort(c *Client, ctx context.Context, portId string) {
	portInfo, err := c.PortService.GetPort(ctx, portId)
	if err != nil {
		suite.FailNowf("could not find port", "could not find port %v", err)
	}

	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portId)

	suite.client.Logger.DebugContext(ctx, "Modifying Port", slog.String("port_id", portId))
	_, modifyErr := c.PortService.ModifyPort(ctx, &ModifyPortRequest{
		PortID:                portId,
		Name:                  newPortName,
		CostCentre:            "",
		MarketplaceVisibility: &portInfo.MarketplaceVisibility,
		WaitForUpdate:         true,
		WaitForTime:           5 * time.Minute,
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
func (suite *PortIntegrationTestSuite) testCancelPort(c *Client, ctx context.Context, portId string) {
	// Soft Delete
	suite.client.Logger.DebugContext(ctx, "Scheduling Port for deletion (30 days).", slog.String("port_id", portId))
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

	suite.client.Logger.DebugContext(ctx, "port scheduled for cancellation", slog.String("status", portInfo.ProvisioningStatus), slog.String("port_id", portId))
	restoreResp, restoreErr := c.PortService.RestorePort(ctx, portId)
	if restoreErr != nil {
		suite.FailNowf("could not restore port", "could not restore port %v", restoreErr)
	}
	suite.True(restoreResp.IsRestored)

}

// testDeletePort tests the deletion of a port, both hard and soft.
func (suite *PortIntegrationTestSuite) testDeletePort(c *Client, ctx context.Context, portId string) {
	// Hard Delete
	suite.client.Logger.DebugContext(ctx, "Deleting Port now.", slog.String("port_id", portId))
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
	suite.client.Logger.DebugContext(ctx, "port deleted", slog.String("status", portInfo.ProvisioningStatus), slog.String("port_id", portId))
}

// testLockPort tests the locking and unlocking of a port.
func (suite *PortIntegrationTestSuite) testLockPort(c *Client, ctx context.Context, portId string) {
	suite.client.Logger.DebugContext(ctx, "Locking Port now.", slog.String("port_id", portId))
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

	suite.client.Logger.DebugContext(ctx, "Test lock of an already locked port.", slog.String("port_id", portId))
	lockRes, lockErr := c.PortService.LockPort(ctx, portId)
	suite.Nil(lockRes)
	suite.Error(ErrPortAlreadyLocked, lockErr)

	suite.client.Logger.DebugContext(ctx, "Unlocking Port now.", slog.String("port_id", portId))
	unlockResp, unlockErr := c.PortService.UnlockPort(ctx, portId)
	if unlockErr != nil {
		suite.FailNowf("could not unlock port", "could not unlock port %v", unlockErr)
	}
	suite.True(unlockResp.IsUnlocking)

	suite.client.Logger.DebugContext(ctx, "Test unlocking of a port that doesn't have a lock.", slog.String("port_id", portId))
	unlockResp, unlockErr = c.PortService.UnlockPort(ctx, portId)
	suite.Nil(unlockResp)
	suite.Error(ErrPortNotLocked, unlockErr)
}
