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
	TEST_LOCATION_A = "Global Switch Sydney"
	TEST_LOCATION_B = "Equinix SY3"
	TEST_LOCATION_C = "90558833-e14f-49cf-84ba-bce1c2c40f2d"
	MCR_LOCATION    = "AU"

	VXC_MVE_TEST_LOCATION = 65
)

// VXCIntegrationTestSuite tests the VXC Service.
type VXCIntegrationTestSuite IntegrationTestSuite

func TestVXCIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(VXCIntegrationTestSuite))
	}
}

func (suite *VXCIntegrationTestSuite) SetupSuite() {
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

// TestVXCBuy tests the VXC buy process.
func (suite *VXCIntegrationTestSuite) TestVXCBuy() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	fuzzySearch, locationErr := locSvc.GetLocationByNameFuzzy(ctx, TEST_LOCATION_A)
	if locationErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locationErr)
	}
	testLocation := fuzzySearch[0]

	logger.InfoContext(ctx, "buying port a end")

	aEndPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port A",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}

	aEndUid := aEndPortRes.TechnicalServiceUIDs[0]

	suite.True(IsGuid(aEndUid), "invalid guid for a end uid")

	serviceKeySvc := suite.client.ServiceKeyService

	logger.InfoContext(ctx, "buying port b end")
	bEndPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port B",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}
	bEndUid := bEndPortRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(bEndUid), "invalid guid for b end uid")

	serviceKeyOrder := &CreateServiceKeyRequest{
		ProductUID:  bEndUid,
		Active:      true,
		SingleUse:   true,
		VLAN:        3,
		PreApproved: true,
		MaxSpeed:    1000,
		ValidFor: &ValidFor{
			StartTime: &Time{
				Time: time.Now(),
			},
			EndTime: &Time{
				Time: time.Now().Add(24 * time.Hour),
			},
		},
	}
	serviceKeyRes, serviceKeyErr := serviceKeySvc.CreateServiceKey(ctx, serviceKeyOrder)
	if serviceKeyErr != nil {
		suite.FailNowf("cannot create service key", "cannot create service key: %s", serviceKeyErr)
	}
	serviceKeyID := serviceKeyRes.ServiceKeyUID

	var aEndVLAN, bEndVLAN int
	var aEndAvailable, bEndAvailable bool
	var err error

	for i := 0; i < 10; i++ {
		aEndVLAN = GenerateRandomVLAN()
		aEndAvailable, err = portSvc.CheckPortVLANAvailability(ctx, aEndUid, aEndVLAN)
		if err != nil {
			suite.FailNowf("cannot check a end vlan availability", "cannot check a end vlan availability %v", err)
		}
		if aEndAvailable {
			break
		}
	}

	if !aEndAvailable {
		suite.FailNowf("a end vlan not available after 10 attempts", "a end vlan %d is not available after 10 attempts", aEndVLAN)
	}

	for i := 0; i < 10; i++ {
		bEndVLAN = GenerateRandomVLAN()
		bEndAvailable, err = portSvc.CheckPortVLANAvailability(ctx, bEndUid, bEndVLAN)
		if err != nil {
			suite.FailNowf("cannot check b end vlan availability", "cannot check b end vlan availability %v", err)
		}
		if bEndAvailable {
			break
		}
	}
	if !bEndAvailable {
		suite.FailNowf("b end vlan not available after 10 attempts", "b end vlan %d is not available after 10 attempts", bEndVLAN)
	}

	logger.InfoContext(ctx, "buying vxc")

	buyVxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:    aEndUid,
		VXCName:    "Test VXC",
		RateLimit:  500,
		Term:       12,
		Shutdown:   false,
		ServiceKey: serviceKeyID,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: aEndVLAN,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN:       bEndVLAN,
			ProductUID: bEndUid,
		},
		WaitForProvision: true,
		WaitForTime:      8 * time.Minute,
		ResourceTags:     testResourceTags,
	})
	if vxcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vxcErr)
	}
	vxcUid := buyVxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUid), "invalid guid for vxc uid")

	var newAEndAvailable, newBEndAvailable bool
	var newAVLAN, newBVLAN int

	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vxcUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	newCostCentre := "Test Cost Centre 2"
	newTerm := 24

	for i := 0; i < 10; i++ {
		newAVLAN = GenerateRandomVLAN()
		newAEndAvailable, err = portSvc.CheckPortVLANAvailability(ctx, aEndUid, newAVLAN)
		if err != nil {
			suite.FailNowf("cannot check new a end vlan availability", "cannot check new a end vlan availability %v", err)
		}
		if newAEndAvailable {
			break
		}
	}
	if !newAEndAvailable {
		suite.FailNowf("new a end vlan not available after 10 attempts", "new a end vlan %d is not available after 10 attempts", newAVLAN)
	}

	for i := 0; i < 10; i++ {
		newBVLAN = GenerateRandomVLAN()
		newBEndAvailable, err = portSvc.CheckPortVLANAvailability(ctx, bEndUid, newBVLAN)
		if err != nil {
			suite.FailNowf("cannot check new b end vlan availability", "cannot check new b end vlan availability %v", err)
		}
		if newBEndAvailable {
			break
		}
	}

	if !newBEndAvailable {
		suite.FailNowf("new b end vlan not available after 10 attempts", "new b end vlan %d is not available after 10 attempts", newBVLAN)
	}

	updateRes, updateErr := vxcSvc.UpdateVXC(ctx, vxcUid, &UpdateVXCRequest{
		AEndVLAN:      &newAVLAN,
		BEndVLAN:      &newBVLAN,
		Name:          PtrTo("Updated VXC"),
		CostCentre:    PtrTo(newCostCentre),
		Term:          PtrTo(newTerm),
		WaitForUpdate: true,
		WaitForTime:   8 * time.Minute,
	})
	if updateErr != nil {
		suite.FailNowf("cannot update vxc", "cannot update vxc %v", updateErr)
	}
	suite.NotNil(updateRes, "update response is nil")

	vxcInfo, getErr := vxcSvc.GetVXC(ctx, vxcUid)
	if getErr != nil {
		suite.FailNowf("cannot get vxc", "cannot get vxc %v", getErr)
	}

	suite.EqualValues("Updated VXC", vxcInfo.Name, "vxc name is not updated")
	suite.EqualValues(newAVLAN, vxcInfo.AEndConfiguration.VLAN, "vxc a end vlan is not updated")
	suite.EqualValues(newBVLAN, vxcInfo.BEndConfiguration.VLAN, "vxc b end vlan is not updated")
	suite.EqualValues(newCostCentre, vxcInfo.CostCentre, "vxc cost centre is not updated")
	suite.EqualValues(newTerm, vxcInfo.ContractTermMonths, "vxc terms are not updated")

	// Test Safe Delete functionality
	// Attempting to delete attached ports prior to deleting VXC. This should result in an error.
	_, deleteErr := portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:     aEndUid,
		DeleteNow:  true,
		SafeDelete: true,
	})
	suite.NotNil(deleteErr, "delete should fail with safe delete")
	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:     bEndUid,
		DeleteNow:  true,
		SafeDelete: true,
	})
	suite.NotNil(deleteErr, "delete should fail with safe delete")

	logger.InfoContext(ctx, "deleting vxc")

	deleteErr = vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting ports")

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    aEndUid,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete a-end port", "cannot delete a-end port %v", deleteErr)
	}

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    bEndUid,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete b-end port", "cannot delete b-end port %v", deleteErr)
	}
}

// TestVXCMoveAEnd tests the VXC move process.
func (suite *VXCIntegrationTestSuite) TestVXCMove() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	fuzzySearch, locationErr := locSvc.GetLocationByNameFuzzy(ctx, TEST_LOCATION_A)
	if locationErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locationErr)
	}
	testLocation := fuzzySearch[0]

	logger.InfoContext(ctx, "buying first port a end")

	aEndFirstPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port A #1",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}

	aEndUidFirst := aEndFirstPortRes.TechnicalServiceUIDs[0]

	suite.True(IsGuid(aEndUidFirst), "invalid guid for a end uid")

	logger.InfoContext(ctx, "buying first port a end")

	aEndSecondPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port A #2",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}

	aEndUidSecond := aEndSecondPortRes.TechnicalServiceUIDs[0]

	suite.True(IsGuid(aEndUidSecond), "invalid guid for a end uid")

	logger.InfoContext(ctx, "buying first port b end")
	bEndFirstPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port B #2",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}
	bEndUidFirst := bEndFirstPortRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(bEndUidFirst), "invalid guid for b end uid")

	logger.InfoContext(ctx, "buying second port b end")
	bEndSecondPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "VXC Port B #2",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}
	bEndUidSecond := bEndSecondPortRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(bEndUidSecond), "invalid guid for b end uid")

	logger.InfoContext(ctx, "buying vxc")

	buyVxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   aEndUidFirst,
		VXCName:   "Test VXC",
		RateLimit: 500,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: GenerateRandomVLAN(),
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN:       GenerateRandomVLAN(),
			ProductUID: bEndUidFirst,
		},
		WaitForProvision: true,
		WaitForTime:      8 * time.Minute,
		ResourceTags:     testResourceTags,
	})
	if vxcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vxcErr)
	}
	vxcUid := buyVxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUid), "invalid guid for vxc uid")

	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vxcUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	logger.InfoContext(ctx, "updating vxc to second a and b end ports")

	updateRes, updateErr := vxcSvc.UpdateVXC(ctx, vxcUid, &UpdateVXCRequest{
		AEndProductUID: &aEndUidSecond,
		BEndProductUID: &bEndUidSecond,
		WaitForUpdate:  true,
		WaitForTime:    10 * time.Minute,
	})
	if updateErr != nil {
		suite.FailNowf("cannot update vxc", "cannot update vxc %v", updateErr)
	}

	suite.Equal(updateRes.AEndConfiguration.UID, aEndUidSecond, "a end port is not moved")
	suite.Equal(updateRes.BEndConfiguration.UID, bEndUidSecond, "b end port is not moved")

	logger.InfoContext(ctx, "deleting vxc")

	deleteErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting ports")

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    aEndUidFirst,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete first a-end port", "cannot delete a-end port %v", deleteErr)
	}

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    bEndUidFirst,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete first b-end port", "cannot delete b-end port %v", deleteErr)
	}

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    aEndUidSecond,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete second a-end port", "cannot delete a-end port %v", deleteErr)
	}

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    bEndUidSecond,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete second b-end port", "cannot delete b-end port %v", deleteErr)
	}
}

// TestAWSVIFConnectionBuy tests the AWS VIF connection buy process.
func (suite *VXCIntegrationTestSuite) TestAWSVIFConnectionBuy() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	testLocation, locErr := locSvc.GetLocationByName(ctx, TEST_LOCATION_B)
	if locErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locErr)
	}

	logger.InfoContext(ctx, "buying port a end")

	portRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "AWS VIF Test Port",
		Term:                  1,
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}
	portUid := portRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(portUid), "invalid guid for port uid")

	logger.InfoContext(ctx, "buying aws vif connection (b-end)")

	var aEndVLAN int
	var vlanAvailable bool
	var err error
	for i := 0; i < 10; i++ {
		aEndVLAN = GenerateRandomVLAN()
		vlanAvailable, err = portSvc.CheckPortVLANAvailability(ctx, portUid, aEndVLAN)
		if err != nil {
			suite.FailNowf("cannot check vlan availability", "cannot check vlan availability %v", err)
		}
		if vlanAvailable {
			break
		}
	}
	if !vlanAvailable {
		suite.FailNowf("vlan not available after 10 attempts", "vlan %d is not available after 10 attempts", aEndVLAN)
	}

	buyVxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   portUid,
		VXCName:   "Hosted AWS VIF Test Connection",
		RateLimit: 500,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: aEndVLAN,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
			PartnerConfig: VXCPartnerConfigAWS{
				ConnectType:  "AWS",
				Type:         "private",
				ASN:          65105,
				AmazonASN:    65106,
				OwnerAccount: "684021030471",
				AuthKey:      "notarealauthkey",
				Prefixes:     "10.0.1.0/24",
			},
		},
		ResourceTags:     testResourceTags,
		WaitForProvision: true,
		WaitForTime:      8 * time.Minute,
	})
	if vxcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vxcErr)
	}
	vxcUid := buyVxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUid), "invalid guid for vxc uid")

	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vxcUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	// Attempt to delete attached ports prior to deleting VXC. This should result in an error.
	_, deleteErr := portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:     portUid,
		DeleteNow:  true,
		SafeDelete: true,
	})
	suite.NotNil(deleteErr, "delete should fail with safe delete")

	logger.InfoContext(ctx, "deleting vxc", slog.String("vxc_uid", vxcUid))

	deleteErr = vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting port", slog.String("port_uid", portUid))

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    portUid,
		DeleteNow: true,
	})
	if deleteErr != nil {
		suite.FailNowf("cannot delete port", "cannot delete port %v", deleteErr)
	}
}

// TestAWSHostedConnectionBuy tests the AWS hosted connection buy process.
func (suite *VXCIntegrationTestSuite) TestAWSHostedConnectionBuy() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	mcrSvc := suite.client.MCRService

	testLocation, locErr := GetRandomLocation(ctx, locSvc, MCR_LOCATION)
	if locErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locErr)
	}

	logger.InfoContext(ctx, "buying mcr (a-end)")
	mcrRes, mcrErr := mcrSvc.BuyMCR(ctx, &BuyMCRRequest{
		Name:             "AWS Hosted Connection Test MCR",
		LocationID:       testLocation.ID,
		Term:             1,
		PortSpeed:        1000,
		MCRAsn:           0,
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
	})
	if mcrErr != nil {
		suite.FailNowf("cannot buy mcr", "cannot buy mcr %v", mcrErr)
	}
	mcrUid := mcrRes.TechnicalServiceUID
	suite.True(IsGuid(mcrUid), "invalid guid for mcr uid")

	logger.InfoContext(ctx, "buying aws hosted connection (b-end)")

	hcRes, hcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:      mcrUid,
		VXCName:      "Hosted Connection AWS Test Connection",
		Term:         1,
		RateLimit:    500,
		Shutdown:     false,
		ResourceTags: testResourceTags,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: GenerateRandomVLAN(),
			PartnerConfig: VXCOrderVrouterPartnerConfig{
				Interfaces: []PartnerConfigInterface{
					{
						IpAddresses: []string{"10.0.0.1/30"},
						IpRoutes: []IpRoute{
							{
								Prefix:      "10.0.0.1/32",
								Description: "Static route 1",
								NextHop:     "10.0.0.2",
							},
						},
						NatIpAddresses: []string{"10.0.0.1"},
						Bfd: BfdConfig{
							TxInterval: 300,
							RxInterval: 300,
							Multiplier: 3,
						},
						BgpConnections: []BgpConnectionConfig{
							{
								PeerAsn:        64512,
								LocalIpAddress: "10.0.0.1",
								PeerIpAddress:  "10.0.0.2",
								Password:       "notARealPAssword",
								Shutdown:       false,
								Description:    "BGP with MED and BFD enabled",
								MedIn:          100,
								MedOut:         100,
								BfdEnabled:     true,
							},
						},
					},
				},
			},
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: "b047870a-adcf-441f-ae34-27a796cdafeb",
			PartnerConfig: VXCPartnerConfigAWS{
				ConnectType:  "AWSHC",
				Type:         "private",
				OwnerAccount: "684021030471",
			},
		},
		WaitForProvision: true,
		WaitForTime:      8 * time.Minute,
	})

	if hcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", hcErr)
	}

	suite.True(IsGuid(hcRes.TechnicalServiceUID), "invalid guid for vxc uid")

	vxcUid := hcRes.TechnicalServiceUID
	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vxcUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	logger.InfoContext(ctx, "deleting vxc", slog.String("vxc_uid", hcRes.TechnicalServiceUID))
	deleteErr := vxcSvc.DeleteVXC(ctx, hcRes.TechnicalServiceUID, &DeleteVXCRequest{DeleteNow: true})
	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting mcr", slog.String("mcr_uid", mcrUid))

	_, deleteErr = mcrSvc.DeleteMCR(ctx, &DeleteMCRRequest{MCRID: mcrUid, DeleteNow: true})
	if deleteErr != nil {
		suite.FailNowf("cannot delete mcr", "cannot delete mcr %v", deleteErr)
	}
}

// TestAWSConnectionBuyDefaults tests the AWS connection buy process with default values.
func (suite *VXCIntegrationTestSuite) TestAWSConnectionBuyDefaults() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	testLocation, locErr := locSvc.GetLocationByName(ctx, TEST_LOCATION_B)
	if locErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locErr)
	}

	logger.InfoContext(ctx, "buying port a end")

	portRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "AWS VIF Test Port",
		Term:                  1,
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}

	portUid := portRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(portUid), "invalid guid for port uid")
	logger.InfoContext(ctx, "buying aws vif connection (b-end)")

	vifRes, vifErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   portUid,
		VXCName:   "Hosted AWS VIF Test Connection",
		RateLimit: 500,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: 0,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
			PartnerConfig: VXCPartnerConfigAWS{
				ConnectType:  "AWS",
				Type:         "private",
				ASN:          65105,
				AmazonASN:    65106,
				OwnerAccount: "684021030471",
			},
		},

		ResourceTags:     testResourceTags,
		WaitForProvision: true,
		WaitForTime:      8 * time.Minute,
	})

	if vifErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vifErr)
	}

	vifUid := vifRes.TechnicalServiceUID
	suite.True(IsGuid(vifUid), "invalid guid for vif uid")

	tags, err := vxcSvc.ListVXCResourceTags(ctx, vifUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vifUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vifUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	logger.InfoContext(ctx, "deleting vxc", slog.String("vxc_uid", vifUid))

	deleteErr := vxcSvc.DeleteVXC(ctx, vifUid, &DeleteVXCRequest{DeleteNow: true})

	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting port", slog.String("port_uid", portUid))

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    portUid,
		DeleteNow: true,
	})

	if deleteErr != nil {
		suite.FailNowf("cannot delete port", "cannot delete port %v", deleteErr)
	}
}

// TestBuyAzureExpressRoute tests the Azure ExpressRoute buy process.
func (suite *VXCIntegrationTestSuite) TestBuyAzureExpressRoute() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	fuzzySearch, locationErr := locSvc.GetLocationByNameFuzzy(ctx, TEST_LOCATION_A)
	if locationErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locationErr)
	}
	testLocation := fuzzySearch[0]

	logger.InfoContext(ctx, "buying azure expressroute port a end")

	aEndPortRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "Azure ExpressRoute Test Port",
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}

	aEndUid := aEndPortRes.TechnicalServiceUIDs[0]

	suite.True(IsGuid(aEndUid), "invalid guid for a end uid")

	serviceKey := "1b2329a5-56dc-45d0-8a0d-87b706297777"

	peerings := []PartnerOrderAzurePeeringConfig{
		{
			Type:            "private",
			PeerASN:         "64555",
			PrimarySubnet:   "10.0.0.0/30",
			SecondarySubnet: "10.0.0.4/30",
			SharedKey:       "SharedKey1",
			VLAN:            100,
		},
		{
			Type:            "microsoft",
			PeerASN:         "64555",
			PrimarySubnet:   "192.88.99.0/30",
			SecondarySubnet: "192.88.99.4/30",
			Prefixes:        "192.88.99.64/26",
			SharedKey:       "SharedKey2",
			VLAN:            200,
		},
	}

	partnerPortRes, partnerPortErr := vxcSvc.LookupPartnerPorts(ctx, &LookupPartnerPortsRequest{
		Key:       serviceKey,
		PortSpeed: 1000,
		Partner:   PARTNER_AZURE,
		ProductID: "",
	})

	if partnerPortErr != nil {
		suite.FailNowf("cannot lookup partner ports", "cannot lookup partner ports %v", partnerPortErr)
	}

	partnerPortId := partnerPortRes.ProductUID

	azurePartnerConfig := VXCPartnerConfigAzure{
		ConnectType: "AzureExpressRoute",
		ServiceKey:  serviceKey,
		Peers:       peerings,
	}

	vxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   aEndUid,
		VXCName:   "Azure ExpressRoute Test VXC",
		RateLimit: 1000,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: 0,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID:    partnerPortId,
			PartnerConfig: azurePartnerConfig,
		},
		ResourceTags:     testResourceTags,
		WaitForProvision: true,
		WaitForTime:      10 * time.Minute,
	})
	if vxcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vxcErr)
	}

	vxcUid := vxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUid), "invalid guid for vxc uid")

	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	err = vxcSvc.UpdateVXCResourceTags(ctx, vxcUid, testUpdatedResourceTags)
	if err != nil {
		suite.FailNowf("cannot update vxc resource tags", "cannot update vxc resource tags %v", err)
	}
	tags, err = vxcSvc.ListVXCResourceTags(ctx, vxcUid)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testUpdatedResourceTags, tags, "updated resource tags are not equal")

	logger.InfoContext(ctx, "deleting vxc")
	deleteErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting port")
	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    aEndUid,
		DeleteNow: true,
	})

	if deleteErr != nil {
		suite.FailNowf("cannot delete port", "cannot delete port %v", deleteErr)
	}
}

// TestBuyGoogleInterconnect tests the Google Interconnect buy process.
func (suite *VXCIntegrationTestSuite) TestBuyGoogleInterconnect() {
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	logger := suite.client.Logger
	locSvc := suite.client.LocationService
	portSvc := suite.client.PortService

	testLocation, locErr := locSvc.GetLocationByName(ctx, TEST_LOCATION_B)
	if locErr != nil {
		suite.FailNowf("cannot find location", "cannot find location %v", locErr)
	}

	logger.InfoContext(ctx, "buying google interconect port a end")

	portRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "Google Interconnect Test Port",
		Term:                  1,
		LocationId:            testLocation.ID,
		PortSpeed:             1000,
		Market:                "AU",
		MarketPlaceVisibility: true,
		WaitForProvision:      true,
		WaitForTime:           5 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("cannot buy port", "cannot buy port %v", portErr)
	}
	portUid := portRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(portUid), "invalid guid for port uid")

	pairingKey := "27325c3a-b640-4b69-a2d5-cdcca797a151/us-west2/1"
	logger.InfoContext(ctx, "buying google interconnect vxc (b-end)")

	partnerPortRes, partnerPortErr := vxcSvc.LookupPartnerPorts(ctx, &LookupPartnerPortsRequest{
		Key:       pairingKey,
		PortSpeed: 1000,
		Partner:   PARTNER_GOOGLE,
		ProductID: "",
	})

	if partnerPortErr != nil {
		suite.FailNowf("cannot lookup partner ports", "cannot lookup partner ports %v", partnerPortErr)
	}

	partnerPortId := partnerPortRes.ProductUID

	suite.True(IsGuid(partnerPortId), "invalid guid for partner port id")

	partnerConfig := VXCPartnerConfigGoogle{
		ConnectType: "GOOGLE",
		PairingKey:  pairingKey,
	}

	vxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   portUid,
		VXCName:   "Test Google Interconnect VXC",
		RateLimit: 1000,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: 0,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID:    partnerPortId,
			PartnerConfig: partnerConfig,
		},
		ResourceTags:     testResourceTags,
		WaitForProvision: true,
		WaitForTime:      10 * time.Minute,
	})

	if vxcErr != nil {
		suite.FailNowf("cannot buy vxc", "cannot buy vxc %v", vxcErr)
	}

	vxcId := vxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcId), "invalid guid for vxc id")
	tags, err := vxcSvc.ListVXCResourceTags(ctx, vxcId)
	if err != nil {
		suite.FailNowf("cannot list vxc resource tags", "cannot list vxc resource tags %v", err)
	}
	suite.EqualValues(testResourceTags, tags, "resource tags are not equal")

	// Attempt to prematurely delete the ports with safe delete enabled. This should fail.
	_, err = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:     portUid,
		DeleteNow:  false,
		SafeDelete: true,
	})
	suite.Error(err, "expected error when deleting port with safe delete enabled")

	logger.InfoContext(ctx, "deleting vxc", slog.String("vxc_uid", vxcRes.TechnicalServiceUID))

	deleteErr := vxcSvc.DeleteVXC(ctx, vxcRes.TechnicalServiceUID, &DeleteVXCRequest{DeleteNow: true})

	if deleteErr != nil {
		suite.FailNowf("cannot delete vxc", "cannot delete vxc %v", deleteErr)
	}

	logger.InfoContext(ctx, "deleting port", slog.String("port_uid", portUid))

	_, deleteErr = portSvc.DeletePort(ctx, &DeletePortRequest{
		PortID:    portUid,
		DeleteNow: true,
	})

	if deleteErr != nil {
		suite.FailNowf("cannot delete port", "cannot delete port %v", deleteErr)
	}
}

// TestMVEtoMVEVXC tests the MVE VXC buy and update process. This test will test an MVE to MVE VXC.
func (suite *VXCIntegrationTestSuite) TestMVEtoMVEVXC() {
	vxcSvc := suite.client.VXCService
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	logger := suite.client.Logger

	logger.InfoContext(ctx, "Starting MVE to MVE VXC test")

	// Setup base Aruba config
	mveConfig := &ArubaConfig{
		Vendor:      "aruba",
		ProductSize: "MEDIUM",
		ImageID:     23,
		AccountName: "test",
		AccountKey:  "test",
		SystemTag:   "test",
	}

	// Create 4 MVEs - we'll connect 1->2 initially, then move to 3->4
	var mveUids []string
	mveNames := []string{"MVE A-End Initial", "MVE B-End Initial", "MVE A-End New", "MVE B-End New"}

	logger.InfoContext(ctx, "Creating 4 MVEs for VXC test")

	for i, mveName := range mveNames {
		buyMVERes, err := mveSvc.BuyMVE(ctx, &BuyMVERequest{
			LocationID:   VXC_MVE_TEST_LOCATION,
			Name:         mveName,
			Term:         12,
			VendorConfig: mveConfig,
			Vnics: []MVENetworkInterface{
				{
					Description: "Test VNIC Index 0",
				},
				{
					Description: "Test VNIC Index 1",
				},
				{
					Description: "Test VNIC Index 2",
				},
			},
			WaitForProvision: true,
			WaitForTime:      5 * time.Minute,
			DiversityZone:    "red",
			ResourceTags:     testResourceTags,
		})
		if err != nil {
			// Clean up any MVEs already created before failing
			for _, uid := range mveUids {
				_, cleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
				if cleanupErr != nil {
					logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", cleanupErr.Error()))
				}
			}
			suite.FailNowf("error buying MVE", "error buying MVE %d (%s): %v", i, mveName, err)
		}

		mveUid := buyMVERes.TechnicalServiceUID
		if !IsGuid(mveUid) {
			// Clean up any MVEs already created before failing
			for _, uid := range mveUids {
				_, cleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
				if cleanupErr != nil {
					logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", cleanupErr.Error()))
				}
			}
			suite.FailNowf("invalid MVE uid", "invalid MVE uid for %s: %s", mveName, mveUid)
		}

		mveUids = append(mveUids, mveUid)
		logger.InfoContext(ctx, "MVE created", slog.String("name", mveName), slog.String("uid", mveUid))
	}

	// Now create a VXC between the first two MVEs
	logger.InfoContext(ctx, "Creating VXC between first two MVEs",
		slog.String("a_end", mveUids[0]),
		slog.String("b_end", mveUids[1]))

	initialVXCName := "MVE to MVE Test VXC"
	initialRateLimit := 200

	buyVxcRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   mveUids[0],
		VXCName:   initialVXCName,
		RateLimit: initialRateLimit,
		Term:      12,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VXCOrderMVEConfig: &VXCOrderMVEConfig{
				NetworkInterfaceIndex: 1,
			},
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: mveUids[1],
			VXCOrderMVEConfig: &VXCOrderMVEConfig{
				NetworkInterfaceIndex: 1,
			},
		},
		WaitForProvision: true,
		WaitForTime:      5 * time.Minute,
		ResourceTags:     testResourceTags,
	})

	if vxcErr != nil {
		// Clean up the MVEs before failing
		for _, uid := range mveUids {
			_, cleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if cleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", cleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot buy VXC", "cannot buy VXC between MVEs: %v", vxcErr)
	}

	vxcUid := buyVxcRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUid), "invalid guid for VXC uid")
	logger.InfoContext(ctx, "VXC created", slog.String("uid", vxcUid))

	// Verify initial VXC configuration
	vxcInfo, getErr := vxcSvc.GetVXC(ctx, vxcUid)
	if getErr != nil {
		// Clean up resources before failing
		cleanupErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
		if cleanupErr != nil {
			logger.ErrorContext(ctx, "Error cleaning up VXC", slog.String("uid", vxcUid), slog.String("error", cleanupErr.Error()))
		}
		for _, uid := range mveUids {
			_, mveCleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if mveCleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", mveCleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot get VXC", "cannot get VXC: %v", getErr)
	}

	suite.Equal(initialVXCName, vxcInfo.Name, "VXC name mismatch")
	suite.Equal(initialRateLimit, vxcInfo.RateLimit, "VXC rate limit mismatch")
	suite.Equal(mveUids[0], vxcInfo.AEndConfiguration.UID, "VXC A-End UID mismatch")
	suite.Equal(mveUids[1], vxcInfo.BEndConfiguration.UID, "VXC B-End UID mismatch")

	// Update VXC fields
	updatedVXCName := "Updated MVE to MVE Test VXC"
	updatedRateLimit := 500

	logger.InfoContext(ctx, "Updating VXC fields",
		slog.String("name", updatedVXCName),
		slog.Int("rate_limit", updatedRateLimit))

	updateRes, updateErr := vxcSvc.UpdateVXC(ctx, vxcUid, &UpdateVXCRequest{
		Name:          PtrTo(updatedVXCName),
		RateLimit:     PtrTo(updatedRateLimit),
		WaitForUpdate: true,
		WaitForTime:   5 * time.Minute,
	})

	if updateErr != nil {
		// Clean up resources before failing
		cleanupErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
		if cleanupErr != nil {
			logger.ErrorContext(ctx, "Error cleaning up VXC", slog.String("uid", vxcUid), slog.String("error", cleanupErr.Error()))
		}
		for _, uid := range mveUids {
			_, mveCleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if mveCleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", mveCleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot update VXC", "cannot update VXC: %v", updateErr)
	}

	suite.Equal(updatedVXCName, updateRes.Name, "Updated VXC name mismatch")

	// Verify updated fields
	vxcInfo, getErr = vxcSvc.GetVXC(ctx, vxcUid)
	if getErr != nil {
		// Clean up resources before failing
		cleanupErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
		if cleanupErr != nil {
			logger.ErrorContext(ctx, "Error cleaning up VXC", slog.String("uid", vxcUid), slog.String("error", cleanupErr.Error()))
		}
		for _, uid := range mveUids {
			_, mveCleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if mveCleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", mveCleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot get VXC", "cannot get VXC after update: %v", getErr)
	}

	suite.Equal(updatedVXCName, vxcInfo.Name, "Updated VXC name mismatch")
	suite.Equal(updatedRateLimit, vxcInfo.RateLimit, "Updated VXC rate limit mismatch")

	// Move the VXC to connect MVE 3 and 4
	logger.InfoContext(ctx, "Moving VXC to connect different MVEs",
		slog.String("new_a_end", mveUids[2]),
		slog.String("new_b_end", mveUids[3]))

	updateRes, updateErr = vxcSvc.UpdateVXC(ctx, vxcUid, &UpdateVXCRequest{
		AEndProductUID: &mveUids[2], // Use MVE 3 for A-End
		BEndProductUID: &mveUids[3], // Use MVE 4 for B-End
		AVnicIndex:     PtrTo(2),    // Use VNIC index 2 for A-End
		BVnicIndex:     PtrTo(2),    // Use VNIC index 2 for B-End
		WaitForUpdate:  true,
		WaitForTime:    10 * time.Minute,
	})

	if updateErr != nil {
		// Clean up resources before failing
		cleanupErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
		if cleanupErr != nil {
			logger.ErrorContext(ctx, "Error cleaning up VXC", slog.String("uid", vxcUid), slog.String("error", cleanupErr.Error()))
		}
		for _, uid := range mveUids {
			_, mveCleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if mveCleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", mveCleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot move VXC", "cannot move VXC: %v", updateErr)
	}

	// Verify that VXC is now connected to MVE 3 and 4
	suite.Equal(mveUids[2], updateRes.AEndConfiguration.UID, "Moved VXC A-End UID mismatch")
	suite.Equal(mveUids[3], updateRes.BEndConfiguration.UID, "Moved VXC B-End UID mismatch")

	// Double check the move by getting the VXC again
	vxcInfo, getErr = vxcSvc.GetVXC(ctx, vxcUid)
	if getErr != nil {
		// Clean up resources before failing
		cleanupErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
		if cleanupErr != nil {
			logger.ErrorContext(ctx, "Error cleaning up VXC", slog.String("uid", vxcUid), slog.String("error", cleanupErr.Error()))
		}
		for _, uid := range mveUids {
			_, mveCleanupErr := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{MVEID: uid})
			if mveCleanupErr != nil {
				logger.ErrorContext(ctx, "Error cleaning up MVE", slog.String("uid", uid), slog.String("error", mveCleanupErr.Error()))
			}
		}
		suite.FailNowf("cannot get VXC", "cannot get VXC after move: %v", getErr)
	}

	suite.Equal(mveUids[2], vxcInfo.AEndConfiguration.UID, "Moved VXC GET A-End UID mismatch")
	suite.Equal(mveUids[3], vxcInfo.BEndConfiguration.UID, "Moved VXC GET B-End UID mismatch")

	// Attempt to delete the MVEs 3 and 4 with safe delete enabled. This should fail.
	_, err := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{
		MVEID:      mveUids[2],
		SafeDelete: true,
	})
	suite.Error(err, "expected error when deleting MVE with safe delete enabled")
	_, err = mveSvc.DeleteMVE(ctx, &DeleteMVERequest{
		MVEID:      mveUids[3],
		SafeDelete: true,
	})
	suite.Error(err, "expected error when deleting MVE with safe delete enabled")

	// Clean up resources - first delete the VXC
	logger.InfoContext(ctx, "Test complete, cleaning up resources")

	deleteErr := vxcSvc.DeleteVXC(ctx, vxcUid, &DeleteVXCRequest{DeleteNow: true})
	if deleteErr != nil {
		logger.ErrorContext(ctx, "Error deleting VXC", slog.String("error", deleteErr.Error()))
	}

	// Then delete all MVEs
	for i, uid := range mveUids {
		logger.InfoContext(ctx, "Deleting MVE",
			slog.String("name", mveNames[i]),
			slog.String("uid", uid))

		deleteRes, err := mveSvc.DeleteMVE(ctx, &DeleteMVERequest{
			MVEID: uid,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Error deleting MVE",
				slog.String("uid", uid),
				slog.String("error", err.Error()))
		} else if !deleteRes.IsDeleted {
			logger.WarnContext(ctx, "MVE delete response indicates not deleted", slog.String("uid", uid))
		}
	}

	logger.InfoContext(ctx, "MVE to MVE VXC test completed successfully")
}
