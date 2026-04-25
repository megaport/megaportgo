package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// MCRClientTestSuite tests the MCR Service.
type MCRClientTestSuite struct {
	ClientTestSuite
}

func TestMCRClientTestSuite(t *testing.T) {
	suite.Run(t, new(MCRClientTestSuite))
}

func (suite *MCRClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *MCRClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestBuyMCR tests the BuyMCR method
func (suite *MCRClientTestSuite) TestBuyMCR() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	want := &BuyMCRResponse{TechnicalServiceUID: productUid}
	req := &BuyMCRRequest{
		LocationID:    1,
		Name:          "test-mcr",
		Term:          1,
		PortSpeed:     1000,
		MCRAsn:        0,
		DiversityZone: "red",
	}
	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`

	mcrOrder := []MCROrder{
		{
			LocationID: req.LocationID,
			Name:       req.Name,
			Term:       1,
			Type:       "MCR2",
			PortSpeed:  1000,
			Config: MCROrderConfig{
				DiversityZone: "red",
				ASN:           0,
			},
		},
	}
	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		var wrapper struct {
			NetworkDesign []MCROrder `json:"networkDesign"`
			DiscountCodes []string   `json:"discountCodes"`
		}
		err := json.NewDecoder(r.Body).Decode(&wrapper)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := wrapper.NetworkDesign
		wantOrder := mcrOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.PortSpeed, gotOrder.PortSpeed)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.Type, gotOrder.Type)
		suite.Equal(wantOrder.Config, gotOrder.Config)
		suite.Equal([]string{}, wrapper.DiscountCodes)
	})
	got, err := mcrSvc.BuyMCR(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetMCR tests the GetMCR method.
func (suite *MCRClientTestSuite) TestGetMCR() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}
	want := &MCR{
		ID:                 1,
		LocationID:         1,
		UID:                productUid,
		Name:               "test-mcr",
		Type:               "MCR2",
		ProvisioningStatus: "LIVE",
		CreateDate:         startDate,
		CreatedBy:          companyUid,
		Market:             "US",
		PortSpeed:          1000,
		CompanyName:        "Test Company",
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location",
			City:    "Atlanta",
			Metro:   "Atlanta",
			Country: "USA",
		},
		ContractTermMonths: 12,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		LiveDate:           startDate,
		TerminateDate:      endDate,
		Cancelable:         true,
		VXCPermitted:       true,
		Virtual:            true,
		CompanyUID:         companyUid,
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_MCR2",
		AttributeTags:      map[string]string{},
		AssociatedVXCs:     []*VXC{},
		AssociatedIXs:      []*IX{},
	}
	jblob := `{
			"message": "test-message",
			"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
			"data": {
		"productId": 1,
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productName": "test-mcr",
		"secondaryName": null,
		"productType": "MCR2",
		"provisioningStatus": "LIVE",
		"locationDetail":{"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},
		"portSpeed": 1000,
		"maxVxcSpeed": 1000,
		"locationId": 1,
		"lagPrimary": false,
		"lagId": null,
		"aggregationId": null,
		"createDate": 1706104800000,
		"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"terminateDate": 1737727200000,
		"liveDate": 1706104800000,
		"contractStartDate": 1706104800000,
		"contractEndDate": 1737727200000,
		"contractTermMonths": 12,
		"market": "US",
		"usageAlgorithm": "POST_PAID_HOURLY_SPEED_MCR2",
		"marketplaceVisibility": false,
		"vxcPermitted": true,
		"vxcAutoApproval": false,
		"virtual": true,
		"buyoutPort": false,
		"locked": false,
		"adminLocked": false,
		"connectType": "VROUTER",
		"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"companyName": "Test Company",
		"associatedVxcs": [],
		"associatedIxs": [],
		"attributeTags": {},
		"up": true,
		"cancelable": true
	}
	}`
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := mcrSvc.GetMCR(ctx, productUid)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestCreatePrefixFilterList tests the CreatePrefixFilterList method.
func (suite *MCRClientTestSuite) TestCreatePrefixFilterList() {
	mcrId := "36b3f68e-2f54-4331-bf94-f8984449365f"
	mcrSvc := suite.client.MCRService
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
		Description:   "new-list",
		AddressFamily: "IPv4",
		Entries:       prefixFilterEntries,
	}
	jblob := `{
    "message": "Created a new prefix list for MCR",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": {
        "id": 2819,
        "description": "new-list",
        "entries": [
            {
                "prefix": "10.0.1.0/24",
                "action": "permit"
            },
            {
                "prefix": "10.0.2.0/24",
                "action": "deny"
            }
        ],
        "addressFamily": "IPv4"
    }
}`
	url := "/v2/product/mcr2/" + mcrId + "/prefixList"
	suite.mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})
	_, prefixErr := mcrSvc.CreatePrefixFilterList(ctx, &CreateMCRPrefixFilterListRequest{
		MCRID:            mcrId,
		PrefixFilterList: validatedPrefixFilterList,
	})
	suite.NoError(prefixErr)
}

// TestListMCRs tests the ListMCRs method
func (suite *MCRClientTestSuite) TestListMCRs() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService

	// Define test data
	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	// Mock API response
	jblob := `{
        "message": "Products retrieved successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "productId": 1,
                "productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
                "productName": "test-mcr-1",
                "productType": "MCR2",
                "provisioningStatus": "LIVE",
                "locationId": 1,
                "portSpeed": 1000,
                "createDate": 1706104800000,
                "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "terminateDate": 1737727200000,
                "contractStartDate": 1706104800000,
                "contractEndDate": 1737727200000,
                "contractTermMonths": 12,
                "market": "US",
                "companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "companyName": "Test Company",
                "locationDetail": {"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},
                "vxcPermitted": true,
                "virtual": true,
                "attributeTags": {}
            },
            {
                "productId": 2,
                "productUid": "46c3f68e-3f54-5331-cf94-g9984449365g",
                "productName": "test-mcr-2",
                "productType": "MCR2",
                "provisioningStatus": "DECOMMISSIONED",
                "locationId": 2,
                "portSpeed": 500,
                "createDate": 1706104800000,
                "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "terminateDate": 1737727200000,
                "contractStartDate": 1706104800000,
                "contractEndDate": 1737727200000,
                "contractTermMonths": 12,
                "market": "EU",
                "companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "companyName": "Test Company",
                "locationDetail": {"name":"Test Location 2","city":"London","metro":"London","country":"UK"},
                "vxcPermitted": true,
                "virtual": true,
                "attributeTags": {}
            },
            {
                "productId": 3,
                "productUid": "56d3f68e-4f54-6331-df94-h9984449365h",
                "productName": "test-port",
                "productType": "MEGAPORT"
            }
        ]
    }`

	// Expected MCRs after filtering
	want := []*MCR{
		{
			ID:                 1,
			UID:                "36b3f68e-2f54-4331-bf94-f8984449365f",
			Name:               "test-mcr-1",
			Type:               "MCR2",
			ProvisioningStatus: "LIVE",
			LocationID:         1,
			PortSpeed:          1000,
			CreateDate:         startDate,
			CreatedBy:          "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			Market:             "US",
			CompanyUID:         "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			CompanyName:        "Test Company",
			ContractStartDate:  startDate,
			ContractEndDate:    endDate,
			TerminateDate:      endDate,
			ContractTermMonths: 12,
			VXCPermitted:       true,
			Virtual:            true,
			AttributeTags:      map[string]string{},
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "Atlanta",
				Metro:   "Atlanta",
				Country: "USA",
			},
		},
	}

	// Set up handler for the /v2/products endpoint
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	// Test with default behavior (exclude inactive)
	req := &ListMCRsRequest{
		IncludeInactive: false,
	}

	got, err := mcrSvc.ListMCRs(ctx, req)
	suite.NoError(err)
	suite.Equal(1, len(got), "Should only return 1 active MCR")
	suite.Equal(want, got)

	// Test with includeInactive=true
	reqWithInactive := &ListMCRsRequest{
		IncludeInactive: true,
	}

	// Update expectations for including inactive MCRs
	wantWithInactive := append(want, &MCR{
		ID:                 2,
		UID:                "46c3f68e-3f54-5331-cf94-g9984449365g",
		Name:               "test-mcr-2",
		Type:               "MCR2",
		ProvisioningStatus: "DECOMMISSIONED",
		LocationID:         2,
		PortSpeed:          500,
		CreateDate:         startDate,
		CreatedBy:          "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		Market:             "EU",
		CompanyUID:         "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		CompanyName:        "Test Company",
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		TerminateDate:      endDate,
		ContractTermMonths: 12,
		VXCPermitted:       true,
		Virtual:            true,
		AttributeTags:      map[string]string{},
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location 2",
			City:    "London",
			Metro:   "London",
			Country: "UK",
		},
	})

	gotWithInactive, err := mcrSvc.ListMCRs(ctx, reqWithInactive)
	suite.NoError(err)
	suite.Equal(2, len(gotWithInactive), "Should return both active and inactive MCRs")
	suite.Equal(wantWithInactive, gotWithInactive)
}

// TestModifyMCR tests the ModifyMCR method.
func (suite *MCRClientTestSuite) TestModifyMCR() {
	ctx := context.Background()

	mcrSvc := suite.client.MCRService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}
	req := &ModifyMCRRequest{
		MCRID:                 productUid,
		Name:                  "test-mcr-updated",
		CostCentre:            "US",
		MarketplaceVisibility: PtrTo(false),
	}
	jblobGet := `{
    "message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": {
		"productId": 1,
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productName": "test-mcr",
		"secondaryName": null,
		"productType": "MCR2",
		"provisioningStatus": "LIVE",
		"portSpeed": 1000,
		"maxVxcSpeed": 1000,
		"locationId": 1,
		"lagPrimary": false,
		"lagId": null,
		"aggregationId": null,
		"createDate": 1706104800000,
		"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"terminateDate": 1737727200000,
		"liveDate": 1706104800000,
		"contractStartDate": 1706104800000,
		"contractEndDate": 1737727200000,
		"contractTermMonths": 12,
		"market": "US",
		"usageAlgorithm": "POST_PAID_HOURLY_SPEED_MCR2",
		"marketplaceVisibility": false,
		"vxcPermitted": true,
		"vxcAutoApproval": false,
		"virtual": true,
		"locationDetail":{"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},
		"buyoutPort": false,
		"locked": false,
		"adminLocked": false,
		"connectType": "VROUTER",
		"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"companyName": "Test Company",
		"associatedVxcs": [],
		"associatedIxs": [],
		"attributeTags": {},
		"up": true,
		"cancelable": true
	}
	}`
	jblob := `{
    "message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": {
		"productId": 1,
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productName": "test-mcr-updated",
		"secondaryName": null,
		"productType": "MCR2",
		"provisioningStatus": "LIVE",
		"portSpeed": 1000,
		"maxVxcSpeed": 1000,
		"locationId": 1,
		"lagPrimary": false,
		"lagId": null,
		"aggregationId": null,
		"createDate": 1706104800000,
		"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"terminateDate": 1737727200000,
		"liveDate": 1706104800000,
		"contractStartDate": 1706104800000,
		"contractEndDate": 1738504800000,
		"contractTermMonths": 12,
		"market": "US",
		"usageAlgorithm": "POST_PAID_HOURLY_SPEED_MCR2",
		"marketplaceVisibility": false,
		"vxcPermitted": true,
		"vxcAutoApproval": false,
		"virtual": true,
		"buyoutPort": false,
		"locked": false,
		"adminLocked": false,
		"connectType": "VROUTER",
		"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"companyName": "Test Company",
		"associatedVxcs": [],
		"associatedIxs": [],
		"attributeTags": {},
		"up": true,
		"cancelable": true
	}
	}`
	wantReq := &ModifyProductRequest{
		ProductID:             req.MCRID,
		ProductType:           PRODUCT_MCR,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MCR, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(ModifyProductRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %V", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblobGet)
	})
	want := &MCR{
		ID:                 1,
		LocationID:         1,
		UID:                productUid,
		Name:               "test-mcr",
		Type:               "MCR2",
		ProvisioningStatus: "LIVE",
		CreateDate:         startDate,
		CreatedBy:          companyUid,
		Market:             "US",
		PortSpeed:          1000,
		CompanyName:        "Test Company",
		ContractTermMonths: 12,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		TerminateDate:      endDate,
		LiveDate:           startDate,
		Cancelable:         true,
		VXCPermitted:       true,
		Virtual:            true,
		CompanyUID:         companyUid,
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_MCR2",
		AttributeTags:      map[string]string{},
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location",
			City:    "Atlanta",
			Metro:   "Atlanta",
			Country: "USA",
		},
		AssociatedVXCs: []*VXC{},
		AssociatedIXs:  []*IX{},
	}
	got, err := mcrSvc.GetMCR(ctx, productUid)
	suite.NoError(err)

	suite.Equal(want, got)

	wantModify := &ModifyMCRResponse{
		IsUpdated: true,
	}
	gotModify, err := mcrSvc.ModifyMCR(ctx, req)
	suite.NoError(err)
	suite.Equal(wantModify, gotModify)
}

// TestDeleteMCR tests the DeleteMCR method.
func (suite *MCRClientTestSuite) TestDeleteMCR() {
	ctx := context.Background()

	mcrSvc := suite.client.MCRService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &DeleteMCRRequest{
		MCRID:     productUid,
		DeleteNow: true,
	}
	jblob := `{
    "message": "Action [CANCEL_NOW Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`
	path := "/v3/product/" + req.MCRID + "/action/CANCEL_NOW"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &DeleteMCRResponse{
		IsDeleting: true,
	}

	got, err := mcrSvc.DeleteMCR(ctx, req)

	suite.NoError(err)
	suite.Equal(want, got)
}

// TestDeleteMCRCancelLaterNotAllowed verifies that DeleteMCR rejects DeleteNow=false.
func (suite *MCRClientTestSuite) TestDeleteMCRCancelLaterNotAllowed() {
	ctx := context.Background()
	req := &DeleteMCRRequest{
		MCRID:     "36b3f68e-2f54-4331-bf94-f8984449365f",
		DeleteNow: false,
	}
	_, err := suite.client.MCRService.DeleteMCR(ctx, req)
	suite.ErrorIs(err, ErrMCRCancelLaterNotAllowed)
}

// TestDeleteMCRNilRequest verifies that DeleteMCR rejects a nil request.
func (suite *MCRClientTestSuite) TestDeleteMCRNilRequest() {
	ctx := context.Background()
	_, err := suite.client.MCRService.DeleteMCR(ctx, nil)
	suite.ErrorIs(err, ErrDeleteMCRRequestNil)
}

// TestRestoreMCR tests the RestoreMCR method.
func (suite *MCRClientTestSuite) TestRestoreMCR() {
	ctx := context.Background()

	mcrSvc := suite.client.MCRService

	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
	"message": "Action [UN_CANCEL Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`
	path := "/v3/product/" + productUid + "/action/UN_CANCEL"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &RestoreMCRResponse{
		IsRestored: true,
	}

	got, err := mcrSvc.RestoreMCR(ctx, productUid)

	suite.NoError(err)
	suite.Equal(want, got)
}

// TestValidateIPsecAddOn tests the validation of IPsec add-on configurations
func (suite *MCRClientTestSuite) TestValidateIPsecAddOn() {
	// Test valid configurations with all valid tunnel counts
	for _, count := range []int{10, 20, 30} {
		validAddOn := &MCRAddOnIPsecConfig{
			TunnelCount: count,
		}
		err := validateMCRAddOn(validAddOn)
		suite.NoError(err, "tunnel count %d should be valid", count)
	}

	// Test valid configuration with zero tunnel count (will default to 10)
	validAddOnZeroTunnels := &MCRAddOnIPsecConfig{
		TunnelCount: 0,
	}
	err := validateMCRAddOn(validAddOnZeroTunnels)
	suite.NoError(err)

	// Test GetAddOnType returns correct type
	ipsecAddOn := &MCRAddOnIPsecConfig{}
	suite.Equal(AddOnTypeIPsec, ipsecAddOn.GetAddOnType())

	// Test invalid: tunnel count not in valid set
	invalidTunnelCount := &MCRAddOnIPsecConfig{
		TunnelCount: 5,
	}
	err = validateMCRAddOn(invalidTunnelCount)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)

	// Test invalid: tunnel count exceeds maximum
	invalidTunnelCountHigh := &MCRAddOnIPsecConfig{
		TunnelCount: 40,
	}
	err = validateMCRAddOn(invalidTunnelCountHigh)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)

	// Test invalid: negative tunnel count
	negativeTunnelCount := &MCRAddOnIPsecConfig{
		TunnelCount: -1,
	}
	err = validateMCRAddOn(negativeTunnelCount)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)
}

// TestBuyMCRWithIPsecValidation tests that BuyMCR validates IPsec add-ons
func (suite *MCRClientTestSuite) TestBuyMCRWithIPsecValidation() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService

	// Test with invalid tunnel count
	req := &BuyMCRRequest{
		LocationID:    1,
		Name:          "test-mcr",
		Term:          1,
		PortSpeed:     1000,
		MCRAsn:        0,
		DiversityZone: "red",
		AddOns: []MCRAddOn{
			&MCRAddOnIPsecConfig{
				TunnelCount: 5, // Invalid - must be 10, 20, or 30
			},
		},
	}

	_, err := mcrSvc.BuyMCR(ctx, req)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)

	// Test with valid tunnel count (10)
	reqValid := &BuyMCRRequest{
		LocationID:    1,
		Name:          "test-mcr",
		Term:          1,
		PortSpeed:     1000,
		MCRAsn:        0,
		DiversityZone: "red",
		AddOns: []MCRAddOn{
			&MCRAddOnIPsecConfig{
				TunnelCount: 10,
			},
		},
	}

	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jblob := `{
		"message": "test-message",
		"terms": "test-terms",
		"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
		]
	}`

	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var wrapper struct {
			NetworkDesign json.RawMessage `json:"networkDesign"`
			DiscountCodes []string        `json:"discountCodes"`
		}
		err := json.NewDecoder(r.Body).Decode(&wrapper)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		suite.NotNil(wrapper.NetworkDesign)
		suite.Equal([]string{}, wrapper.DiscountCodes)

		var orders []map[string]interface{}
		suite.NoError(json.Unmarshal(wrapper.NetworkDesign, &orders))
		suite.Len(orders, 1)
		suite.Equal("test-mcr", orders[0]["productName"])
		suite.Equal(float64(1), orders[0]["term"])
		suite.Equal(float64(1000), orders[0]["portSpeed"])
		suite.Equal(float64(1), orders[0]["locationId"])
		fmt.Fprint(w, jblob)
	})

	got, err := mcrSvc.BuyMCR(ctx, reqValid)
	suite.NoError(err)
	suite.Equal(productUid, got.TechnicalServiceUID)
}

// TestUpdateMCRWithAddOn tests that UpdateMCRWithAddOn posts the correct payload without waiting.
func (suite *MCRClientTestSuite) TestUpdateMCRWithAddOn() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	mcrID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	suite.mux.HandleFunc(fmt.Sprintf("/v3/product/%s/addon", mcrID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)

		v := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		suite.Equal(AddOnTypeIPsec, v["addOnType"])
		suite.Equal(float64(10), v["tunnelCount"])

		fmt.Fprint(w, `{"message":"ok"}`)
	})

	err := mcrSvc.UpdateMCRWithAddOn(ctx, mcrID, MCRAddOnRequest{
		AddOn: &MCRAddOnIPsecConfig{
			AddOnType:   AddOnTypeIPsec,
			TunnelCount: 10,
		},
	})
	suite.NoError(err)
}

// TestUpdateMCRWithAddOnWaitTimeout tests that WaitForProvision returns an error on timeout.
func (suite *MCRClientTestSuite) TestUpdateMCRWithAddOnWaitTimeout() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	mcrID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	suite.mux.HandleFunc(fmt.Sprintf("/v3/product/%s/addon", mcrID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, `{"message":"ok"}`)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", mcrID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprintf(w, `{"data":{"productUid":%q,"provisioningStatus":"CONFIGURING"}}`, mcrID)
	})

	err := mcrSvc.UpdateMCRWithAddOn(ctx, mcrID, MCRAddOnRequest{
		AddOn: &MCRAddOnIPsecConfig{
			AddOnType:   AddOnTypeIPsec,
			TunnelCount: 10,
		},
		WaitForProvision: true,
		WaitForTime:      100 * time.Millisecond,
	})
	suite.Error(err)
	suite.Contains(err.Error(), "time expired")
}

// TestUpdateMCRIPsecAddOn tests the UpdateMCRIPsecAddOn method
func (suite *MCRClientTestSuite) TestUpdateMCRIPsecAddOn() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	mcrID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	addOnUID := "addon-12345"

	// Test with valid tunnel count (10)
	jblob := `{
		"message": "IPsec add-on updated successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	suite.mux.HandleFunc(fmt.Sprintf("/v3/product/%s/addon/%s", mcrID, addOnUID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPut)

		// Verify the payload
		v := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}

		suite.Equal(AddOnTypeIPsec, v["addOnType"])
		suite.Equal(float64(10), v["tunnelCount"]) // JSON numbers decode as float64

		fmt.Fprint(w, jblob)
	})

	err := mcrSvc.UpdateMCRIPsecAddOn(ctx, mcrID, addOnUID, 10)
	suite.NoError(err)
}

// TestUpdateMCRIPsecAddOnDisable tests disabling IPsec by setting tunnel count to 0
func (suite *MCRClientTestSuite) TestUpdateMCRIPsecAddOnDisable() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	mcrID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	addOnUID := "addon-12345"

	jblob := `{
		"message": "IPsec add-on disabled successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	suite.mux.HandleFunc(fmt.Sprintf("/v3/product/%s/addon/%s", mcrID, addOnUID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPut)

		// Verify the payload
		v := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}

		suite.Equal(AddOnTypeIPsec, v["addOnType"])
		suite.Equal(float64(0), v["tunnelCount"]) // 0 to disable

		fmt.Fprint(w, jblob)
	})

	err := mcrSvc.UpdateMCRIPsecAddOn(ctx, mcrID, addOnUID, 0)
	suite.NoError(err)
}

// TestUpdateMCRIPsecAddOnInvalidTunnelCount tests validation of tunnel count
func (suite *MCRClientTestSuite) TestUpdateMCRIPsecAddOnInvalidTunnelCount() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	mcrID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	addOnUID := "addon-12345"

	// Test invalid tunnel count (5)
	err := mcrSvc.UpdateMCRIPsecAddOn(ctx, mcrID, addOnUID, 5)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)

	// Test invalid tunnel count (11 - exceeds max)
	err = mcrSvc.UpdateMCRIPsecAddOn(ctx, mcrID, addOnUID, 11)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)

	// Test invalid negative tunnel count
	err = mcrSvc.UpdateMCRIPsecAddOn(ctx, mcrID, addOnUID, -1)
	suite.Error(err)
	suite.Equal(ErrInvalidIPsecTunnelCount, err)
}

func (suite *MCRClientTestSuite) TestGetMCRTelemetry() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": [
			{
				"type": "BITS",
				"subtype": "IN",
				"samples": [[1608516536000, 125.5], [1608517536000, 130.2]],
				"unit": {"name": "Mbps", "fullName": "Megabits per second"}
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("7", r.URL.Query().Get("days"))
		suite.Equal([]string{"BITS"}, r.URL.Query()["type"])
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	resp, err := mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
	suite.Equal("BITS", resp.Type)
	suite.Equal(int64(1608516536000), resp.TimeFrame.From)
	suite.Equal(int64(1608603936000), resp.TimeFrame.To)
	suite.Len(resp.Data, 1)
	suite.Equal("BITS", resp.Data[0].Type)
	suite.Equal("IN", resp.Data[0].Subtype)
	suite.Len(resp.Data[0].Samples, 2)
	suite.Equal(int64(1608516536000), resp.Data[0].Samples[0].Timestamp)
	suite.Equal(125.5, resp.Data[0].Samples[0].Value)
	suite.Equal("Mbps", resp.Data[0].Unit.Name)
	suite.Equal("Megabits per second", resp.Data[0].Unit.FullName)
}

func (suite *MCRClientTestSuite) TestGetMCRTelemetryWithFromTo() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": []
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("1608516536000", r.URL.Query().Get("from"))
		suite.Equal("1608603936000", r.URL.Query().Get("to"))
		suite.Equal([]string{"BITS", "PACKETS"}, r.URL.Query()["type"])
		suite.Empty(r.URL.Query().Get("days"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	fromTime := time.UnixMilli(1608516536000)
	toTime := time.UnixMilli(1608603936000)
	resp, err := mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS", "PACKETS"},
		From:       &fromTime,
		To:         &toTime,
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
}

func (suite *MCRClientTestSuite) TestGetMCRTelemetryValidation() {
	ctx := context.Background()
	mcrSvc := suite.client.MCRService

	// Missing ProductUID
	_, err := mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		Types: []string{"BITS"},
		Days:  PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrMCRTelemetryProductUIDRequired)

	// Missing Types
	_, err = mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: "some-uid",
		Days:       PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrMCRTelemetryTypesRequired)

	// Days and From/To mutually exclusive
	fromTime := time.UnixMilli(1608516536000)
	_, err = mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
		From:       &fromTime,
	})
	suite.ErrorIs(err, ErrMCRTelemetryTimeExclusive)

	// Days out of range
	_, err = mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](0),
	})
	suite.ErrorIs(err, ErrMCRTelemetryDaysOutOfRange)

	// From without To
	_, err = mcrSvc.GetMCRTelemetry(ctx, &GetMCRTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		From:       &fromTime,
	})
	suite.ErrorIs(err, ErrMCRTelemetryFromToIncomplete)
}
