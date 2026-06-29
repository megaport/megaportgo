package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// MVEClientTestSuite tests the MVE Service Client.
type MVEClientTestSuite struct {
	ClientTestSuite
}

func TestMVEClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MVEClientTestSuite))
}

func (suite *MVEClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *MVEClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestBuyMVE tests the BuyMVE method
func (suite *MVEClientTestSuite) TestBuyMVE() {
	ctx := context.Background()
	mveSvc := suite.client.MVEService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &BuyMVERequest{
		Name:                  "test-mve",
		Term:                  12,
		LocationID:            1,
		DiversityZone:         "blue",
		MarketplaceVisibility: PtrTo(false),
	}
	jblob := `{
    "message": "MVE [36b3f68e-2f54-4331-bf94-f8984449365f] created.",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [
        {
            "serviceName": "test-mve",
            "name": "test-mve",
            "secondaryName": null,
            "technicalServiceId": 1,
            "technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
            "requestedDate": 1707237079607,
            "configuredDate": null,
            "currentEstimatedDelivery": null,
            "companyName": "test-company",
            "companyId": 1,
            "billingContactName": null,
            "billingContactId": null,
            "adminContactName": null,
            "adminContactId": null,
            "technicalContactName": null,
            "technicalContactId": null,
            "salesName": null,
            "salesId": 1,
            "billableId": 1,
            "billableUsageAlgorithm": null,
            "productType": "MVE",
            "provisioningStatus": "DEPLOYABLE",
            "failedReason": null,
            "inAdvanceBillingStatus": null,
            "provisioningItems": [],
            "tags": [],
            "vxcDistanceBand": null,
            "intercapPath": null,
            "marketplaceVisibility": false,
            "vxcPermitted": true,
            "vxcAutoApproval": false,
            "createDate": 1706104800000,
            "terminationDate": null,
            "contractStartDate": 1706104800000,
            "contractTermMonths": 12,
            "rateType": "MONTHLY",
            "trialAgreement": false,
            "payerCompanyId": null,
            "nonPayerCompanyId": null,
            "minimumSpeed": null,
            "maximumSpeed": null,
            "rateLimit": null,
            "errorMessage": null,
            "lagId": null,
            "aggregationId": null,
            "lagPrimary": null,
            "market": "USA",
            "accountManager": null,
            "promptUid": null,
            "components": [],
            "attributes": [],
            "aLocation": null,
            "bLocation": null,
            "aMetro": "Atlanta",
            "aCountry": "USA",
            "aLocationId": 1,
            "bLocationId": null,
            "bMetro": null,
            "bCountry": null,
            "attributeTags": {},
            "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
            "buyoutPort": false,
            "virtual": false,
            "locked": false,
            "adminLocked": false,
            "bgpShutdownDefault": false,
            "vendor": "PALO_ALTO",
            "mveSize": "SMALL",
            "originDomain": null
        }
    ]
}`
	mveOrder := []MVEOrderConfig{{
		LocationID:        req.LocationID,
		Name:              req.Name,
		Term:              req.Term,
		ProductType:       strings.ToUpper(PRODUCT_MVE),
		NetworkInterfaces: []MVENetworkInterface{{Description: "Data Plane", VLAN: 0}},
		Config: MVEConfig{
			DiversityZone: req.DiversityZone,
		},
		MarketplaceVisibility: PtrTo(false),
	}}
	want := &BuyMVEResponse{TechnicalServiceUID: productUid}

	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		var wrapper struct {
			NetworkDesign []MVEOrderConfig `json:"networkDesign"`
			DiscountCodes []string         `json:"discountCodes"`
		}
		err := json.NewDecoder(r.Body).Decode(&wrapper)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}

		orders := wrapper.NetworkDesign
		wantOrder := mveOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.ProductType, gotOrder.ProductType)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.NetworkInterfaces, gotOrder.NetworkInterfaces)
		suite.Equal(wantOrder.Config, gotOrder.Config)
		suite.Equal(wantOrder.MarketplaceVisibility, gotOrder.MarketplaceVisibility)
		suite.Equal([]string{}, wrapper.DiscountCodes)
	})
	got, err := mveSvc.BuyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestCreateMVEOrderMarketplaceVisibility asserts the marshaled MVEOrderConfig
// carries marketplaceVisibility when set, and omits it (preserving the API
// default) when the request leaves it nil.
func (suite *MVEClientTestSuite) TestCreateMVEOrderMarketplaceVisibility() {
	visible := suite.marshalMVEOrder(&BuyMVERequest{MarketplaceVisibility: PtrTo(true)})
	suite.Contains(visible, `"marketplaceVisibility":true`)

	hidden := suite.marshalMVEOrder(&BuyMVERequest{MarketplaceVisibility: PtrTo(false)})
	suite.Contains(hidden, `"marketplaceVisibility":false`)

	unset := suite.marshalMVEOrder(&BuyMVERequest{})
	suite.NotContains(unset, "marketplaceVisibility")
}

func (suite *MVEClientTestSuite) marshalMVEOrder(req *BuyMVERequest) string {
	orders := createMVEOrder(req)
	suite.Require().Len(orders, 1)
	b, err := json.Marshal(orders[0])
	suite.Require().NoError(err)
	return string(b)
}

// TestListMVEs tests the ListMVEs method which lists provisioned MVE products
func (suite *MVEClientTestSuite) TestListMVEs() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()

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
                "productName": "test-mve-1",
                "productType": "MVE",
                "provisioningStatus": "LIVE",
                "locationId": 1,
                "createDate": 1706104800000,
                "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "terminateDate": 1737727200000,
                "contractStartDate": 1706104800000,
                "contractEndDate": 1737727200000,
                "contractTermMonths": 12,
                "market": "US",
                "companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "companyName": "test-company",
                "vendor": "PALO_ALTO",
                "mveSize": "SMALL"
            },
            {
                "productId": 2,
                "productUid": "46c3f68e-3f54-5331-cf94-g9984449365g",
                "productName": "test-mve-2",
                "productType": "MVE",
                "provisioningStatus": "DECOMMISSIONED",
                "locationId": 2,
                "createDate": 1706104800000,
                "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "terminateDate": 1737727200000,
                "contractStartDate": 1706104800000,
                "contractEndDate": 1737727200000,
                "contractTermMonths": 12,
                "market": "EU",
                "companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
                "companyName": "test-company",
                "vendor": "CISCO",
                "mveSize": "MEDIUM"
            },
            {
                "productId": 3,
                "productUid": "56d3f68e-4f54-6331-df94-h9984449365h",
                "productName": "test-port",
                "productType": "MEGAPORT"
            }
        ]
    }`

	// Expected MVEs after filtering
	want := []*MVE{
		{
			ID:                 1,
			UID:                "36b3f68e-2f54-4331-bf94-f8984449365f",
			Name:               "test-mve-1",
			Type:               "MVE",
			ProvisioningStatus: "LIVE",
			CreateDate:         startDate,
			CreatedBy:          "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			Market:             "US",
			LocationID:         1,
			CompanyUID:         "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			CompanyName:        "test-company",
			ContractStartDate:  startDate,
			ContractEndDate:    endDate,
			TerminateDate:      endDate,
			ContractTermMonths: 12,
			Vendor:             "PALO_ALTO",
			Size:               "SMALL",
		},
	}

	// Set up handler for the /v2/products endpoint
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	// Test with default behavior (exclude inactive)
	req := &ListMVEsRequest{
		IncludeInactive: false,
	}

	got, err := mveSvc.ListMVEs(ctx, req)
	suite.NoError(err)
	suite.Equal(1, len(got), "Should only return 1 active MVE")
	suite.Equal(want, got)

	// Test with includeInactive=true
	reqWithInactive := &ListMVEsRequest{
		IncludeInactive: true,
	}

	// Update expectations for including inactive MVEs
	wantWithInactive := append(want, &MVE{
		ID:                 2,
		UID:                "46c3f68e-3f54-5331-cf94-g9984449365g",
		Name:               "test-mve-2",
		Type:               "MVE",
		ProvisioningStatus: "DECOMMISSIONED",
		CreateDate:         startDate,
		CreatedBy:          "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		Market:             "EU",
		LocationID:         2,
		CompanyUID:         "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		CompanyName:        "test-company",
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		TerminateDate:      endDate,
		ContractTermMonths: 12,
		Vendor:             "CISCO",
		Size:               "MEDIUM",
	})

	gotWithInactive, err := mveSvc.ListMVEs(ctx, reqWithInactive)
	suite.NoError(err)
	suite.Equal(2, len(gotWithInactive), "Should return both active and inactive MVEs")
	suite.Equal(wantWithInactive, gotWithInactive)
}

// TestGetMVE tests the GetMVE method
func (suite *MVEClientTestSuite) TestGetMVE() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}
	jblob := `{
	"message": "Found Product 36b3f68e-2f54-4331-bf94-f8984449365f",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
		"productId": 1,
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productName": "test-mve",
		"secondaryName": null,
		"productType": "MVE",
		"provisioningStatus": "LIVE",
		"portSpeed": null,
		"maxVxcSpeed": 10000,
		"locationId": 1,
		"lagPrimary": false,
		"lagId": null,
		"locationDetail":{"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},
		"aggregationId": null,
		"createDate": 1706104800000,
		"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"terminateDate": 1737727200000,
		"liveDate": null,
		"contractStartDate": 1706104800000,
		"contractEndDate": 1737727200000,
		"contractTermMonths": 12,
		"market": "US",
		"usageAlgorithm": "POST_PAID_FIXED",
		"marketplaceVisibility": false,
		"vxcPermitted": true,
		"vxcAutoApproval": false,
		"virtual": false,
		"buyoutPort": false,
		"locked": false,
		"adminLocked": false,
		"vendor": "PALO_ALTO",
		"mveSize": "SMALL",
		"mveLabel": "MVE 2/8",
		"connectType": "MVE",
		"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
        "companyName": "test-company",
		"associatedVxcs": [],
		"associatedIxs": [],
		"attributeTags": {},
		"vnics": [
			{
				"vlan": 0,
				"description": "Management"
			},
			{
				"vlan": 1,
				"description": "Data Plane"
			}
		],
		"resources": {
			"interface": {
				"demarcation": "",
				"loa_template": "megaport",
				"media": "LR4",
				"port_speed": 40000,
				"resource_name": "interface",
				"resource_type": "interface",
				"up": 1
			},
			"virtual_machine": [
				{
					"cpu_count": 2,
					"id": 0,
					"image": {
						"id": 0,
						"vendor": "palo alto",
						"product": "test product",
						"version": "1.0"
					},
					"resource_type": "virtual_machine",
					"up": true,
					"vnics": [
						{
							"vlan": 0,
							"description": "Management"
						},
						{
							"vlan": 1,
							"description": "Data Plane"
						}
					]
				}
			]
		},
		"diversityZone": "blue",
		"up": false,
		"cancelable": true
	}
}`
	path := "/v2/product/" + productUid
	wantMVE := &MVE{
		ID:                    1,
		UID:                   productUid,
		Name:                  "test-mve",
		Type:                  "MVE",
		ProvisioningStatus:    "LIVE",
		CreateDate:            startDate,
		CreatedBy:             companyUid,
		Market:                "US",
		LocationID:            1,
		UsageAlgorithm:        "POST_PAID_FIXED",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		CompanyUID:            companyUid,
		CompanyName:           "test-company",
		ContractStartDate:     startDate,
		ContractEndDate:       endDate,
		TerminateDate:         endDate,
		ContractTermMonths:    12,
		Virtual:               false,
		BuyoutPort:            false,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location",
			City:    "Atlanta",
			Metro:   "Atlanta",
			Country: "USA",
		},
		DiversityZone: "blue",
		AttributeTags: map[string]string{},
		Resources: &MVEResources{
			Interface: &PortInterface{
				Demarcation:  "",
				LOATemplate:  "megaport",
				Media:        "LR4",
				PortSpeed:    40000,
				ResourceName: "interface",
				ResourceType: "interface",
				Up:           1,
			},
			VirtualMachines: []*MVEVirtualMachine{
				{
					CpuCount: 2,
					ID:       0,
					Image: &MVEVirtualMachineImage{
						ID:      0,
						Vendor:  "palo alto",
						Product: "test product",
						Version: "1.0",
					},
					ResourceType: "virtual_machine",
					Up:           true,
					Vnics: []*MVENetworkInterface{
						{VLAN: 0, Description: "Management"},
						{VLAN: 1, Description: "Data Plane"},
					},
				},
			},
		},
		Vendor: "PALO_ALTO",
		Size:   "SMALL",
		NetworkInterfaces: []*MVENetworkInterface{
			{
				VLAN:        0,
				Description: "Management",
			},
			{
				VLAN:        1,
				Description: "Data Plane",
			},
		},
		AssociatedVXCs: []*VXC{},
		AssociatedIXs:  []*IX{},
	}
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := mveSvc.GetMVE(ctx, productUid)
	suite.NoError(err)
	suite.Equal(wantMVE, got)
}

// TestModifyMVE tests the ModifyMVE method
func (suite *MVEClientTestSuite) TestModifyMVE() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &ModifyMVERequest{
		MVEID: productUid,
		Name:  "test-mve-updated",
	}
	jblob := `{
	"message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
		"productId": 1,
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productName": "test-mve-updated",
		"secondaryName": null,
		"productType": "MVE",
		"provisioningStatus": "LIVE",
		"portSpeed": null,
		"maxVxcSpeed": 10000,
		"locationId": 1,
		"lagPrimary": false,
		"lagId": null,
		"aggregationId": null,
		"createDate": 1706104800000,
		"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
		"terminateDate": 1737727200000,
		"liveDate": null,
		"contractStartDate": 1706104800000,
		"contractEndDate": 1737727200000,
		"contractTermMonths": 12,
		"market": "US",
		"usageAlgorithm": "POST_PAID_FIXED",
		"marketplaceVisibility": false,
		"vxcPermitted": true,
		"vxcAutoApproval": false,
		"virtual": false,
		"buyoutPort": false,
		"locked": false,
		"adminLocked": false,
		"vendor": "PALO_ALTO",
		"mveSize": "SMALL",
		"mveLabel": "MVE 2/8",
		"connectType": "MVE",
		"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
        "companyName": "test-company",
		"associatedIxs": [],
		"attributeTags": {},
		"vnics": [
			{
				"vlan": 0,
				"description": "Management"
			},
			{
				"vlan": 1,
				"description": "Data Plane"
			}
		],
		"resources": {},
		"diversityZone": "blue",
		"up": false,
		"cancelable": true
	}
}`
	wantReq := &ModifyProductRequest{
		ProductID:   req.MVEID,
		ProductType: PRODUCT_MVE,
		Name:        req.Name,
		CostCentre:  "",
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MVE, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			suite.FailNowf("could not read body", "%v", err)
		}
		v := new(ModifyProductRequest)
		if err := json.Unmarshal(body, v); err != nil {
			suite.FailNowf("could not decode json", "%v", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
		// When the caller doesn't set MarketplaceVisibility, it must not
		// appear on the wire — otherwise the backend would flip visibility.
		suite.NotContains(string(body), "marketplaceVisibility")
	})
	wantRes := &ModifyMVEResponse{
		MVEUpdated: true,
	}
	gotRes, err := mveSvc.ModifyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}

// TestModifyMVEMarketplaceVisibility verifies the request only carries
// marketplaceVisibility when the caller sets it explicitly.
func (suite *MVEClientTestSuite) TestModifyMVEMarketplaceVisibility() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &ModifyMVERequest{
		MVEID:                 productUid,
		MarketplaceVisibility: PtrTo(true),
	}
	jblob := `{"message":"updated","data":{"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productType":"MVE"}}`
	wantReq := &ModifyProductRequest{
		ProductID:             req.MVEID,
		ProductType:           PRODUCT_MVE,
		MarketplaceVisibility: PtrTo(true),
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MVE, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			suite.FailNowf("could not read body", "%v", err)
		}
		v := new(ModifyProductRequest)
		if err := json.Unmarshal(body, v); err != nil {
			suite.FailNowf("could not decode json", "%v", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
		suite.Contains(string(body), `"marketplaceVisibility":true`)
	})
	gotRes, err := mveSvc.ModifyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(&ModifyMVEResponse{MVEUpdated: true}, gotRes)
}

// TestModifyMVEWithVnics verifies vNIC descriptions are forwarded
// through to the PUT /v2/product/mve/{uid} body as the API expects.
func (suite *MVEClientTestSuite) TestModifyMVEWithVnics() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &ModifyMVERequest{
		MVEID: productUid,
		Vnics: []MVEVnicUpdate{
			{Description: "Management"},
			{Description: "Data Plane"},
		},
	}
	jblob := `{
	"message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
		"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
		"productType": "MVE"
	}
}`
	wantReq := &ModifyProductRequest{
		ProductID:   req.MVEID,
		ProductType: PRODUCT_MVE,
		Vnics: []MVEVnicUpdate{
			{Description: "Management"},
			{Description: "Data Plane"},
		},
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MVE, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			suite.FailNowf("could not read body", "%v", err)
		}
		v := new(ModifyProductRequest)
		if err := json.Unmarshal(body, v); err != nil {
			suite.FailNowf("could not decode json", "%v", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
		// Confirm the wire format matches the public OpenAPI contract:
		// ServiceUpdateRequest.vnics is an array of {description}.
		suite.Contains(string(body), `"vnics":[{"description":"Management"},{"description":"Data Plane"}]`)
	})
	gotRes, err := mveSvc.ModifyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(&ModifyMVEResponse{MVEUpdated: true}, gotRes)
}

// TestModifyMVENilRequest verifies ModifyMVE rejects a nil request up front
// instead of panicking on field access.
func (suite *MVEClientTestSuite) TestModifyMVENilRequest() {
	ctx := context.Background()
	gotRes, err := suite.client.MVEService.ModifyMVE(ctx, nil)
	suite.ErrorIs(err, ErrModifyMVERequestNil)
	suite.Nil(gotRes)
}

// TestModifyMVECostCentreTooLong verifies ModifyMVE rejects a CostCentre
// over 255 characters before dispatching the HTTP request.
func (suite *MVEClientTestSuite) TestModifyMVECostCentreTooLong() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MVE, "36b3f68e-2f54-4331-bf94-f8984449365f")
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.FailNow("ModifyMVE must not call the API when CostCentre is too long")
	})
	req := &ModifyMVERequest{
		MVEID:      "36b3f68e-2f54-4331-bf94-f8984449365f",
		CostCentre: strings.Repeat("x", 256),
	}
	gotRes, err := mveSvc.ModifyMVE(ctx, req)
	suite.ErrorIs(err, ErrCostCentreTooLong)
	suite.Nil(gotRes)
}

func (suite *MVEClientTestSuite) TestListMVEImages() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	// v4 API response format - nested structure with multiple product groups
	// This tests that the flattening logic correctly denormalizes Product/Vendor
	// from parent level to each individual image
	jblob := `{
		"message": "Current supported MVE images",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
		  "mveImages": [
			{
			  "product": "FortiGate-VM",
			  "vendor": "Fortinet",
			  "vendorProductId": "fortinet_fortigate-vm",
			  "images": [
				{
				  "id": 56,
				  "version": "6.4.15",
				  "productCode": "fortigate",
				  "vendorDescription": null,
				  "releaseImage": true,
				  "availableSizes": ["MVE 2/8", "MVE 4/16", "MVE 8/32"]
				},
				{
				  "id": 57,
				  "version": "7.0.14",
				  "productCode": "fortigate",
				  "vendorDescription": null,
				  "releaseImage": true,
				  "availableSizes": ["MVE 2/8", "MVE 4/16", "MVE 8/32", "MVE 12/48"]
				}
			  ]
			},
			{
			  "product": "C8000",
			  "vendor": "Cisco",
			  "vendorProductId": "cisco_c8000",
			  "images": [
				{
				  "id": 92,
				  "version": "17.16.01a",
				  "productCode": "c8000",
				  "vendorDescription": "Cisco Catalyst 8000V Edge Software",
				  "releaseImage": true,
				  "availableSizes": ["MVE 2/8", "MVE 4/16", "MVE 8/32", "MVE 12/48"]
				}
			  ]
			},
			{
			  "product": "vMX",
			  "vendor": "Meraki",
			  "vendorProductId": "meraki_vmx",
			  "images": [
				{
				  "id": 97,
				  "version": "Meraki Classic 19.2",
				  "productCode": "meraki-vmx",
				  "vendorDescription": null,
				  "releaseImage": true,
				  "availableSizes": ["MVE 2/8"]
				}
			  ]
			}
		  ]
		}
	}`
	// After flattening, each image should have Product/Vendor denormalized from parent
	want := []*MVEImage{
		{
			ID:                56,
			Version:           "6.4.15",
			Product:           "FortiGate-VM",
			Vendor:            "Fortinet",
			VendorDescription: "",
			ReleaseImage:      true,
			ProductCode:       "fortigate",
			AvailableSizes:    []string{"MVE 2/8", "MVE 4/16", "MVE 8/32"},
		},
		{
			ID:                57,
			Version:           "7.0.14",
			Product:           "FortiGate-VM",
			Vendor:            "Fortinet",
			VendorDescription: "",
			ReleaseImage:      true,
			ProductCode:       "fortigate",
			AvailableSizes:    []string{"MVE 2/8", "MVE 4/16", "MVE 8/32", "MVE 12/48"},
		},
		{
			ID:                92,
			Version:           "17.16.01a",
			Product:           "C8000",
			Vendor:            "Cisco",
			VendorDescription: "Cisco Catalyst 8000V Edge Software",
			ReleaseImage:      true,
			ProductCode:       "c8000",
			AvailableSizes:    []string{"MVE 2/8", "MVE 4/16", "MVE 8/32", "MVE 12/48"},
		},
		{
			ID:                97,
			Version:           "Meraki Classic 19.2",
			Product:           "vMX",
			Vendor:            "Meraki",
			VendorDescription: "",
			ReleaseImage:      true,
			ProductCode:       "meraki-vmx",
			AvailableSizes:    []string{"MVE 2/8"},
		},
	}
	suite.mux.HandleFunc("/v4/product/mve/images", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := mveSvc.ListMVEImages(ctx)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *MVEClientTestSuite) TestListAvailableMVESizes() {
	mveSvc := suite.client.MVEService
	ctx := context.Background()
	jblob := `{
		"message": "Current supported MVE sizes",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
		  {
			"size": "SMALL",
			"label": "MVE 2/8",
			"cpuCoreCount": 2,
			"ramGB": 8
		  },
		  {
			"size": "MEDIUM",
			"label": "MVE 4/16",
			"cpuCoreCount": 4,
			"ramGB": 16
		  }]}`
	want := []*MVESize{
		{
			Size:         "SMALL",
			Label:        "MVE 2/8",
			CPUCoreCount: 2,
			RamGB:        8,
		},
		{
			Size:         "MEDIUM",
			Label:        "MVE 4/16",
			CPUCoreCount: 4,
			RamGB:        16,
		},
	}
	suite.mux.HandleFunc("/v3/product/mve/variants", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := mveSvc.ListAvailableMVESizes(ctx)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestDeleteMVE tests the DeleteMVE method.
func (suite *MVEClientTestSuite) TestDeleteMVE() {
	ctx := context.Background()

	mveSvc := suite.client.MVEService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
    "message": "Action [CANCEL_NOW Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	req := &DeleteMVERequest{
		MVEID: productUid,
	}

	path := "/v3/product/" + req.MVEID + "/action/CANCEL_NOW"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &DeleteMVEResponse{
		IsDeleted: true,
	}

	got, err := mveSvc.DeleteMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestCiscoConfigAdminPasswordMarshalling verifies that CiscoConfig.AdminPassword
// round-trips through JSON serialisation under the wire-format key "adminPassword".
// The Megaport API requires this field for Cisco FTDv MVE buy orders.
func (suite *MVEClientTestSuite) TestCiscoConfigAdminPasswordMarshalling() {
	cfg := CiscoConfig{
		Vendor:        "cisco",
		ImageID:       1,
		ProductSize:   "SMALL",
		AdminPassword: "s3cret-plaintext",
	}

	raw, err := json.Marshal(cfg)
	suite.NoError(err)
	suite.Contains(string(raw), `"adminPassword":"s3cret-plaintext"`)

	var decoded CiscoConfig
	suite.NoError(json.Unmarshal(raw, &decoded))
	suite.Equal("s3cret-plaintext", decoded.AdminPassword)

	// Empty value must be omitted from the wire payload.
	empty := CiscoConfig{Vendor: "cisco", ImageID: 1, ProductSize: "SMALL"}
	emptyRaw, err := json.Marshal(empty)
	suite.NoError(err)
	suite.NotContains(string(emptyRaw), "adminPassword")
}

// TestPaloAltoConfigAdminPasswordMarshalling verifies that PaloAltoConfig.AdminPassword
// round-trips through JSON serialisation under the wire-format key "adminPassword".
// The Megaport API requires this field for Palo Alto MVE buy orders.
func (suite *MVEClientTestSuite) TestPaloAltoConfigAdminPasswordMarshalling() {
	cfg := PaloAltoConfig{
		Vendor:        "palo_alto",
		ImageID:       1,
		ProductSize:   "SMALL",
		AdminPassword: "s3cret-plaintext",
	}

	raw, err := json.Marshal(cfg)
	suite.NoError(err)
	suite.Contains(string(raw), `"adminPassword":"s3cret-plaintext"`)

	var decoded PaloAltoConfig
	suite.NoError(json.Unmarshal(raw, &decoded))
	suite.Equal("s3cret-plaintext", decoded.AdminPassword)

	// Empty value must be omitted from the wire payload.
	empty := PaloAltoConfig{Vendor: "palo_alto", ImageID: 1, ProductSize: "SMALL"}
	emptyRaw, err := json.Marshal(empty)
	suite.NoError(err)
	suite.NotContains(string(emptyRaw), "adminPassword")
}

// TestMVENilRequestGuards verifies that the required-request MVE methods reject
// a nil request with a sentinel error instead of panicking.
func (suite *MVEClientTestSuite) TestMVENilRequestGuards() {
	ctx := context.Background()
	tests := []struct {
		name string
		call func() error
		want error
	}{
		{"BuyMVE", func() error { _, err := suite.client.MVEService.BuyMVE(ctx, nil); return err }, ErrBuyMVERequestNil},
		{"ValidateMVEOrder", func() error { return suite.client.MVEService.ValidateMVEOrder(ctx, nil) }, ErrBuyMVERequestNil},
		{"ModifyMVE", func() error { _, err := suite.client.MVEService.ModifyMVE(ctx, nil); return err }, ErrModifyMVERequestNil},
		{"DeleteMVE", func() error { _, err := suite.client.MVEService.DeleteMVE(ctx, nil); return err }, ErrDeleteMVERequestNil},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.ErrorIs(tt.call(), tt.want)
		})
	}
}

// TestListMVEsNilRequest verifies that ListMVEs treats a nil request as "no filter".
func (suite *MVEClientTestSuite) TestListMVEsNilRequest() {
	ctx := context.Background()
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, `{"message":"OK","data":[]}`)
	})
	mves, err := suite.client.MVEService.ListMVEs(ctx, nil)
	suite.NoError(err)
	suite.NotNil(mves)
}
