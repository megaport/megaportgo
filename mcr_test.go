package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]MCROrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
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
		MaxVXCSpeed:        1000,
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
		MaxVXCSpeed:        1000,
		CompanyUID:         companyUid,
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_MCR2",
		AttributeTags:      map[string]string{},
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location",
			City:    "Atlanta",
			Metro:   "Atlanta",
			Country: "USA",
		},
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
