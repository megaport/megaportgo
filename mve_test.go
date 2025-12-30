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
		Name:          "test-mve",
		Term:          12,
		LocationID:    1,
		DiversityZone: "blue",
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
		suite.Equal(wantOrder.Config, gotOrder.Config)
	})
	got, err := mveSvc.BuyMVE(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
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
		ProductID:             req.MVEID,
		ProductType:           PRODUCT_MVE,
		Name:                  req.Name,
		CostCentre:            "",
		MarketplaceVisibility: PtrTo(false),
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
