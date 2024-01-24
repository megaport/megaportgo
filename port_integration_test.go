package megaport

// const (
// 	TEST_LOCATION_ID_A = 19 // 	Interactive 437 Williamstown
// )

// // TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
// func TestSinglePort(t *testing.T) {
// 	ctx := context.Background()

// 	testLocation, err := megaportClient.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}

// 	portsListInitial, err := megaportClient.PortService.ListPorts(ctx)
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}

// 	portConfirmation, portErr := testCreatePort(megaportClient, ctx, types.SINGLE_PORT, *testLocation)
// 	if !assert.NoError(t, portErr) {
// 		t.FailNow()
// 	}

// 	portId := portConfirmation.TechnicalServiceUID

// 	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
// 		t.FailNow()
// 	}

// 	portCreated, err := megaportClient.PortService.WaitForPortProvisioning(ctx, portId)

// 	if !assert.NoError(t, err) || !portCreated {
// 		t.FailNow()
// 	}

// 	portsListPostCreate, err := megaportClient.PortService.ListPorts(ctx)
// 	if err != nil {
// 		megaportClient.Logger.Debug("Failed to get ports list", "error", err)
// 		t.FailNow()
// 	}

// 	portIsActuallyNew := true
// 	for _, p := range portsListInitial {
// 		if p.UID == portId {
// 			portIsActuallyNew = false
// 		}
// 	}

// 	if !portIsActuallyNew {
// 		megaportClient.Logger.Debug("Failed to find port we just created in ports list", "port_id", portId)
// 		t.FailNow()
// 	}

// 	foundNewPort := false
// 	for _, p := range portsListPostCreate {
// 		if p.UID == portId {
// 			foundNewPort = true
// 		}
// 	}

// 	if !foundNewPort {
// 		megaportClient.Logger.Debug("Failed to find port we just created in ports list", "port_id", portId)
// 		t.FailNow()
// 	}

// 	testModifyPort(megaportClient, ctx, portId, types.SINGLE_PORT, t)
// 	testLockPort(megaportClient, ctx, portId, t)
// 	testCancelPort(megaportClient, ctx, portId, types.SINGLE_PORT, t)
// 	testDeletePort(megaportClient, ctx, portId, types.SINGLE_PORT, t)

// }

// // TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
// func TestLAGPort(t *testing.T) {
// 	ctx := context.Background()

// 	testLocation, err := megaportClient.LocationService.GetLocationByID(ctx, TEST_LOCATION_ID_A)

// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}

// 	portsListInitial, err := megaportClient.PortService.ListPorts(ctx)
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}

// 	portConfirmation, portErr := testCreatePort(megaportClient, ctx, types.LAG_PORT, *testLocation)

// 	portId := portConfirmation.TechnicalServiceUID

// 	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
// 		t.FailNow()
// 	}

// 	portCreated, err := megaportClient.PortService.WaitForPortProvisioning(ctx, portId)

// 	if !assert.NoError(t, err) || !portCreated {
// 		t.FailNow()
// 	}

// 	portsListPostCreate, err := megaportClient.PortService.ListPorts(ctx)
// 	if err != nil {
// 		megaportClient.Logger.Error("Failed to get ports list", "error", err)
// 		t.FailNow()
// 	}

// 	portIsActuallyNew := true
// 	for _, p := range portsListInitial {
// 		if p.UID == portId {
// 			portIsActuallyNew = false
// 		}
// 	}

// 	if !portIsActuallyNew {
// 		megaportClient.Logger.Debug("Failed to find port we just created in ports list", "port_id", portId)
// 		t.FailNow()
// 	}

// 	foundNewPort := false
// 	for _, p := range portsListPostCreate {
// 		if p.UID == portId {
// 			foundNewPort = true
// 		}
// 	}

// 	if !foundNewPort {
// 		megaportClient.Logger.Debug("Failed to find port we just created in ports list", "port_id", portId)
// 		t.FailNow()
// 	}

// 	testModifyPort(megaportClient, ctx, portId, types.LAG_PORT, t)
// 	testCancelPort(megaportClient, ctx, portId, types.LAG_PORT, t)
// }

// func testCreatePort(c *Client, ctx context.Context, portType string, location types.Location) (*types.PortOrderConfirmation, error) {
// 	var portConfirm *types.PortOrderConfirmation
// 	var portErr error

// 	megaportClient.Logger.Debug("Buying Port", "port_type", portType)
// 	if portType == types.LAG_PORT {
// 		portConfirm, portErr = c.PortService.BuyLAGPort(ctx, &BuyLAGPortRequest{
// 			Name:       "Buy Port (LAG) Test",
// 			Term:       1,
// 			PortSpeed:  10000,
// 			LocationId: location.ID,
// 			Market:     location.Market,
// 			LagCount:   2,
// 			IsPrivate:  true,
// 		})
// 	} else {
// 		portConfirm, portErr = c.PortService.BuySinglePort(ctx, &BuySinglePortRequest{
// 			Name:       "Buy Port (Single) Test",
// 			Term:       1,
// 			PortSpeed:  10000,
// 			LocationId: location.ID,
// 			Market:     location.Market,
// 			IsPrivate:  true,
// 		})
// 	}
// 	if portErr != nil {
// 		return nil, portErr
// 	}
// 	megaportClient.Logger.Debug("Port Purchased", "port_confirmation", portConfirm)
// 	return portConfirm, portErr
// }

// func testModifyPort(c *Client, ctx context.Context, portId string, portType string, t *testing.T) {
// 	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
// 		PortID: portId,
// 	})
// 	assert.NoError(t, err)

// 	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portType)

// 	megaportClient.Logger.Debug("Modifying Port", "port_id", portId, "port_type", portType)
// 	_, modifyErr := c.PortService.ModifyPort(ctx, &ModifyPortRequest{
// 		PortID:                portId,
// 		Name:                  newPortName,
// 		CostCentre:            "",
// 		MarketplaceVisibility: portInfo.MarketplaceVisibility,
// 	})
// 	assert.NoError(t, modifyErr)

// 	secondGetPortInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
// 		PortID: portId,
// 	})
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, newPortName, secondGetPortInfo.Name)
// }

// // PortScript tests the remaining lifecycle for a Port (not dependant on port-type), Go-Live, Modification,
// // and Soft/Hard Deletes.
// func testCancelPort(c *Client, ctx context.Context, portId string, portType string, t *testing.T) {
// 	// Soft Delete
// 	megaportClient.Logger.Debug("Scheduling Port for deletion (30 days).", "port_id", portId, "port_type", portType)
// 	resp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
// 		PortID:    portId,
// 		DeleteNow: false,
// 	})
// 	assert.NoError(t, deleteErr)
// 	assert.True(t, resp.IsDeleting)

// 	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{PortID: portId})
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, types.STATUS_CANCELLED, portInfo.ProvisioningStatus)

// 	megaportClient.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
// 	restoreResp, restoreErr := c.PortService.RestorePort(ctx, &RestorePortRequest{PortID: portId})
// 	assert.NoError(t, restoreErr)
// 	assert.True(t, restoreResp.IsRestoring)

// }

// func testDeletePort(c *Client, ctx context.Context, portId string, portType string, t *testing.T) {
// 	// Hard Delete
// 	megaportClient.Logger.Debug("Deleting Port now.", "port_type", portType, "port_id", portId)
// 	hardDeleteResp, deleteErr := c.PortService.DeletePort(ctx, &DeletePortRequest{
// 		PortID:    portId,
// 		DeleteNow: true,
// 	})
// 	assert.True(t, hardDeleteResp.IsDeleting)
// 	assert.NoError(t, deleteErr)

// 	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
// 		PortID: portId,
// 	})
// 	assert.NoError(t, err)

// 	assert.EqualValues(t, types.STATUS_DECOMMISSIONED, portInfo.ProvisioningStatus)
// 	megaportClient.Logger.Debug("", "status", portInfo.ProvisioningStatus, "port_id", portId)
// }

// func testLockPort(c *Client, ctx context.Context, portId string, t *testing.T) {
// 	megaportClient.Logger.Debug("Locking Port now.", "port_id", portId)
// 	lockResp, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
// 	assert.True(t, lockResp.IsLocking)
// 	assert.NoError(t, lockErr)

// 	portInfo, err := c.PortService.GetPort(ctx, &GetPortRequest{
// 		PortID: portId,
// 	})
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, true, portInfo.Locked)

// 	megaportClient.Logger.Debug("Test lock of an already locked port.", "port_id", portId)
// 	lockRes, lockErr := c.PortService.LockPort(ctx, &LockPortRequest{PortID: portId})
// 	assert.Nil(t, lockRes)
// 	assert.Error(t, errors.New(mega_err.ERR_PORT_ALREADY_LOCKED), lockErr)

// 	megaportClient.Logger.Debug("Unlocking Port now.", "port_id", portId)
// 	unlockResp, unlockErr := c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
// 	assert.True(t, unlockResp.IsUnlocking)
// 	assert.NoError(t, unlockErr)

// 	megaportClient.Logger.Debug("Test unlocking of a port that doesn't have a lock.", "port_id", portId)
// 	unlockResp, unlockErr = c.PortService.UnlockPort(ctx, &UnlockPortRequest{PortID: portId})
// 	assert.Nil(t, unlockResp)
// 	assert.Error(t, errors.New(mega_err.ERR_PORT_NOT_LOCKED), unlockErr)
// }
