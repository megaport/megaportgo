package megaport

import (
	"context"
	"encoding/json"
	"fmt"
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
		Name:       "test-mve",
		Term:       12,
		LocationID: 1,
		VendorConfig: PaloAltoConfig{
			ImageID:           32,
			ProductSize:       "SMALL",
			Vendor:            "palo alto",
			AdminSSHPublicKey: "test-key",
			AdminPasswordHash: "test-hash",
		},
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
		LocationID:  req.LocationID,
		Name:        req.Name,
		Term:        req.Term,
		ProductType: strings.ToUpper(PRODUCT_MVE),
		VendorConfig: &PaloAltoConfig{
			ImageID:           32,
			ProductSize:       "SMALL",
			Vendor:            "palo alto",
			AdminSSHPublicKey: "test-key",
			AdminPasswordHash: "test-hash",
		},
		NetworkInterfaces: []MVENetworkInterface{{Description: "Data Plane", VLAN: 0}},
	}}
	want := &BuyMVEResponse{TechnicalServiceUID: productUid}

	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]MVEOrderConfig)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}

		orders := *v
		wantOrder := mveOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.ProductType, gotOrder.ProductType)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.NetworkInterfaces, gotOrder.NetworkInterfaces)
	})
	got, err := mveSvc.BuyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
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
		AttributeTags:         map[string]string{},
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
		ProductID:             req.MVEID,
		ProductType:           PRODUCT_MVE,
		Name:                  req.Name,
		CostCentre:            "",
		MarketplaceVisibility: false,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MVE, productUid)
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
	wantRes := &ModifyMVEResponse{
		MVEUpdated: true,
	}
	gotRes, err := mveSvc.ModifyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
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
