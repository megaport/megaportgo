package megaport

import (
	"context"
	"log/slog"
	"os"
	"sync"
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
		MarketplaceVisibility: &mcr.MarketplaceVisibility,
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
	logger.InfoContext(ctx, "Deleting MCR now.")

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

	logger.InfoContext(ctx, "Buying MCR Port.")
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
		ResourceTags:     testResourceTags,
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

	mcrDetails, err := mcrSvc.GetMCR(ctx, mcrId)
	if err != nil {
		suite.FailNowf("could not get mcr", "could not get mcr %v", err)
	}
	suite.EqualValues(mcrDetails.ResourceTags, testResourceTags)

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

// TestCreatePrefixFilterList tests the creation of a prefix filter list for an MCR.
func (suite *MCRIntegrationTestSuite) TestMegaportPrefixFilterList() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	mcrSvc := suite.client.MCRService
	logger := suite.client.Logger

	logger.InfoContext(ctx, "Buying MCR Port.")
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

	prefixFilterEntries1 := []*MCRPrefixListEntry{
		{
			Action: "permit",
			Prefix: "10.0.1.0/24",
			Ge:     25,
			Le:     32,
		},
		{
			Action: "deny",
			Prefix: "10.0.2.0/24",
			Ge:     24,
			Le:     25,
		},
	}

	prefixFilterEntries2 := []*MCRPrefixListEntry{
		{
			Action: "permit",
			Prefix: "10.0.1.0/24",
			Ge:     26,
			Le:     32,
		},
		{
			Action: "deny",
			Prefix: "10.0.2.0/24",
			Ge:     25,
			Le:     27,
		},
	}

	validatedPrefixFilterList1 := MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 1",
		AddressFamily: "IPv4",
		Entries:       prefixFilterEntries1,
	}

	validatedPrefixFilterList2 := MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 2",
		AddressFamily: "IPv4",
		Entries:       prefixFilterEntries2,
	}

	want1 := &MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 1",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     25,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     0,
				Le:     25,
			},
		},
	}
	want2 := &MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 2",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     26,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     25,
				Le:     27,
			},
		},
	}
	prefixRes1, prefixErr := mcrSvc.CreatePrefixFilterList(ctx, &CreateMCRPrefixFilterListRequest{
		MCRID:            mcrId,
		PrefixFilterList: validatedPrefixFilterList1,
	})
	if prefixErr != nil {
		suite.FailNowf("could not create prefix filter list", "could not create prefix filter list %v", prefixErr)
	}
	want1.ID = prefixRes1.PrefixFilterListID

	prefixRes2, prefixErr := mcrSvc.CreatePrefixFilterList(ctx, &CreateMCRPrefixFilterListRequest{
		MCRID:            mcrId,
		PrefixFilterList: validatedPrefixFilterList2,
	})
	if prefixErr != nil {
		suite.FailNowf("could not create prefix filter list", "could not create prefix filter list %v", prefixErr)
	}
	want2.ID = prefixRes2.PrefixFilterListID

	ids := []int{prefixRes1.PrefixFilterListID, prefixRes2.PrefixFilterListID}

	wg := sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			getPrefixFilterListRes, getPrefixFilterListErr := mcrSvc.GetMCRPrefixFilterList(ctx, mcrId, id)
			if getPrefixFilterListErr != nil {
				suite.FailNowf("could not get prefix filter list", "could not get prefix filter list %v", getPrefixFilterListErr)
			}
			switch id {
			case want1.ID:
				suite.EqualValues(want1, getPrefixFilterListRes)
			case want2.ID:
				suite.EqualValues(want2, getPrefixFilterListRes)
			}
		}(id)
	}
	wg.Wait()

	update1 := &MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 1 Updated",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     26,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     24,
				Le:     26,
			},
		},
	}
	update2 := &MCRPrefixFilterList{
		Description:   "Test Prefix Filter List 2 Updated",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     27,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     25,
				Le:     27,
			},
		},
	}
	wantUpdate1 := &MCRPrefixFilterList{
		ID:            want1.ID,
		Description:   "Test Prefix Filter List 1 Updated",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     26,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     0,
				Le:     26,
			},
		},
	}
	wantUpdate2 := &MCRPrefixFilterList{
		ID:            want2.ID,
		Description:   "Test Prefix Filter List 2 Updated",
		AddressFamily: "IPv4",
		Entries: []*MCRPrefixListEntry{
			{
				Action: "permit",
				Prefix: "10.0.1.0/24",
				Ge:     27,
				Le:     32,
			},
			{
				Action: "deny",
				Prefix: "10.0.2.0/24",
				Ge:     25,
				Le:     27,
			},
		},
	}
	_, updateErr := mcrSvc.ModifyMCRPrefixFilterList(ctx, mcrId, want1.ID, update1)
	if updateErr != nil {
		suite.FailNowf("could not update prefix filter list", "could not update prefix filter list %v", updateErr)
	}
	_, updateErr = mcrSvc.ModifyMCRPrefixFilterList(ctx, mcrId, want2.ID, update2)
	if updateErr != nil {
		suite.FailNowf("could not update prefix filter list", "could not update prefix filter list %v", updateErr)
	}
	wantUpdate1.ID = want1.ID
	wantUpdate2.ID = want2.ID

	// Check for Updated MCR Prefix Filter Lists
	wg2 := sync.WaitGroup{}
	for _, id := range ids {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			getPrefixFilterListRes, getPrefixFilterListErr := mcrSvc.GetMCRPrefixFilterList(ctx, mcrId, id)
			if getPrefixFilterListErr != nil {
				suite.FailNowf("could not get prefix filter list", "could not get prefix filter list %v", getPrefixFilterListErr)
			}
			switch id {
			case wantUpdate1.ID:
				suite.EqualValues(wantUpdate1, getPrefixFilterListRes)
			case wantUpdate2.ID:
				suite.EqualValues(wantUpdate2, getPrefixFilterListRes)
			}
		}(id)
	}
	wg2.Wait()

	// Delete MCR Prefix Filter List
	_, deleteErr := mcrSvc.DeleteMCRPrefixFilterList(ctx, mcrId, want2.ID)
	if deleteErr != nil {
		suite.FailNowf("could not delete prefix filter list", "could not delete prefix filter list %v", deleteErr)
	}

	// Check for Deleted MCR Prefix Filter List
	listRes, listErr := mcrSvc.ListMCRPrefixFilterLists(ctx, mcrId)
	if listErr != nil {
		suite.FailNowf("could not list prefix filter lists", "could not list prefix filter lists %v", listErr)
	}
	suite.Len(listRes, 1)

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
