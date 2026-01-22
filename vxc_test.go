package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestCreateVXCOrder_PortUIDAutoPopulation tests that PortUID is automatically
// populated from AEndConfiguration.ProductUID when PortUID is not set.
func TestCreateVXCOrder_PortUIDAutoPopulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		req            *BuyVXCRequest
		expectedPortID string
	}{
		{
			name: "PortUID is set explicitly - should use PortUID",
			req: &BuyVXCRequest{
				PortUID: "explicit-port-uid",
				VXCName: "test-vxc",
				AEndConfiguration: VXCOrderEndpointConfiguration{
					ProductUID: "a-end-product-uid",
				},
			},
			expectedPortID: "explicit-port-uid",
		},
		{
			name: "PortUID is empty but AEndConfiguration.ProductUID is set - should use AEndConfiguration.ProductUID",
			req: &BuyVXCRequest{
				PortUID: "",
				VXCName: "test-vxc",
				AEndConfiguration: VXCOrderEndpointConfiguration{
					ProductUID: "a-end-product-uid",
				},
			},
			expectedPortID: "a-end-product-uid",
		},
		{
			name: "Both PortUID and AEndConfiguration.ProductUID are empty - should remain empty",
			req: &BuyVXCRequest{
				PortUID: "",
				VXCName: "test-vxc",
				AEndConfiguration: VXCOrderEndpointConfiguration{
					ProductUID: "",
				},
			},
			expectedPortID: "",
		},
		{
			name: "PortUID is set and AEndConfiguration.ProductUID is empty - should use PortUID",
			req: &BuyVXCRequest{
				PortUID: "explicit-port-uid",
				VXCName: "test-vxc",
				AEndConfiguration: VXCOrderEndpointConfiguration{
					ProductUID: "",
				},
			},
			expectedPortID: "explicit-port-uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders := createVXCOrder(tt.req)

			assert.Len(t, orders, 1, "Expected exactly one VXC order")
			assert.Equal(t, tt.expectedPortID, orders[0].PortID, "PortID mismatch")
		})
	}
}

// VXCClientTestSuite tests the VXC service.
type VXCClientTestSuite struct {
	ClientTestSuite
}

func TestVXCClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(VXCClientTestSuite))
}

func (suite *VXCClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *VXCClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestBuyVXC tests the BuyVXC method.
func (suite *VXCClientTestSuite) TestBuyVXC() {
	portProductUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	vxcProductUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &BuyVXCRequest{
		PortUID:   portProductUid,
		VXCName:   "test-vxc",
		RateLimit: 50,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			VLAN: 0,
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: vxcProductUid,
			VLAN:       0,
		},
	}
	vxcOrder := []VXCOrder{{
		PortID: req.PortUID,
		AssociatedVXCs: []VXCOrderConfiguration{
			{
				Name:      req.VXCName,
				RateLimit: req.RateLimit,
				AEnd:      req.AEndConfiguration,
				BEnd:      req.BEndConfiguration,
				Term:      req.Term,
				Shutdown:  req.Shutdown,
			},
		},
	}}
	vxcSvc := suite.client.VXCService
	ctx := context.Background()
	jblob := `{
		"message": "VXC [f36b3f68e-2f54-4331-bf94-f8984449365f] created.",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"createDate": 1706104800000,
				"vxcOrderId": 1,
				"payerMegaPortId": 1,
				"nonPayerMegaPortId": 1,
				"payerMegaPortName": "test-port",
				"nonPayerMegaPortName": null,
				"payerCompanyId": 1153,
				"nonPayerCompanyId": 117,
				"payerLocationId": 226,
				"nonPayerLocationId": 75,
				"salesId": null,
				"payerCompanyName": "Test Company",
				"nonPayerCompanyName": "AWS",
				"payerMegaPortNsId": 1,
				"nonPayerMegaPortNsId": 1,
				"payerVlanId": 0,
				"nonPayerVlanId": 0,
				"payerInnerVlanId": null,
				"nonPayerInnerVlanId": null,
				"payerApproverName": "test user",
				"payerApproverId": 1,
				"nonPayerApproverName": "test user",
				"nonPayerApproverId": 1,
				"payerApproval": 1,
				"nonPayerApproval": 1,
				"fixedTerm": true,
				"duration": 1,
				"rollover": true,
				"serviceName": "test-vxc",
				"payerStatus": "APPROVED",
				"nonPayerStatus": "APPROVED",
				"speed": 50,
				"distanceBand": "ZONE",
				"intercapPath": "",
				"awsId": null,
				"promoCode": null,
				"dealUid": null,
				"rateType": "MONTHLY",
				"vxcJTechnicalServiceId": 1,
				"vxcJTechnicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
				"provisionDate": 1,
				"orderType": "NEW",
				"monthlyDiscountAmount": null,
				"discountMonths": null,
				"amazonDirectConnectConfigDto": null,
				"amsixConnectConfigDto": null,
				"sdrcProvItem": null,
				"rate": null,
				"setup": null,
				"asn": null,
				"bgpPassword": null,
				"usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
				"costCentre": "test-cost-centre",
				"azureServiceKey": null,
				"oracleVirtualCircuitId": null,
				"serviceKey": null,
				"vxc": {
					"serviceName": "test-vxc",
					"name": "test-vxc",
					"secondaryName": null,
					"technicalServiceId": 1,
					"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
					"requestedDate": 1,
					"configuredDate": null,
					"currentEstimatedDelivery": null,
					"companyName": "Test Company",
					"companyId": 1,
					"billingContactName": null,
					"billingContactId": null,
					"adminContactName": null,
					"adminContactId": null,
					"technicalContactName": null,
					"technicalContactId": null,
					"salesName": null,
					"salesId": null,
					"billableId": 1,
					"billableUsageAlgorithm": null,
					"productType": "VXC",
					"provisioningStatus": "DEPLOYABLE",
					"failedReason": null,
					"inAdvanceBillingStatus": null,
					"provisioningItems": [],
					"tags": [],
					"vxcDistanceBand": "ZONE",
					"intercapPath": "",
					"marketplaceVisibility": true,
					"vxcPermitted": true,
					"vxcAutoApproval": false,
					"createDate": 1706104800000,
					"terminationDate": null,
					"contractStartDate": 1706104800000,
					"contractEndDate": 1737727200000,
					"contractTermMonths": 12,
					"rateType": "MONTHLY",
					"trialAgreement": false,
					"payerCompanyId": null,
					"nonPayerCompanyId": null,
					"minimumSpeed": null,
					"maximumSpeed": null,
					"rateLimit": 50,
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
					"aMetro": null,
					"aCountry": null,
					"aLocationId": null,
					"bLocationId": null,
					"bMetro": null,
					"bCountry": null,
					"attributeTags": {},
					"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
					"buyoutPort": null,
					"virtual": false,
					"locked": false,
					"adminLocked": false,
					"bgpShutdownDefault": false,
					"originDomain": null
				},
				"connectType": null,
				"payerConfig": {},
				"nonPayerConfig": {},
				"attributeTags": {},
				"serviceLicense": null,
				"originDomain": null,
				"fullyApproved": true
			}
		]
	}`
	path := "/v3/networkdesign/buy"
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new([]VXCOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := vxcOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.PortID, gotOrder.PortID)
		suite.Equal(&wantOrder.AssociatedVXCs, &gotOrder.AssociatedVXCs)
	})
	want := &BuyVXCResponse{
		TechnicalServiceUID: vxcProductUid,
	}
	got, err := vxcSvc.BuyVXC(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetVXCs tests the GetVXC method.
func (suite *VXCClientTestSuite) TestGetVXC() {
	ctx := context.Background()
	vxcSvc := suite.client.VXCService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	vxcUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	portUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	bEndUid := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	wantVxc := &VXC{
		ID:                 1,
		UID:                vxcUid,
		Name:               "test-vxc",
		Type:               "VXC",
		RateLimit:          50,
		DistanceBand:       "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy:          companyUid,
		CreateDate:         startDate,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		Resources: &VXCResources{
			CSPConnection: &CSPConnection{
				CSPConnection: []CSPConnectionConfig{
					CSPConnectionAWSHC{
						Bandwidth:    50,
						ConnectType:  "AWSHC",
						ResourceName: "b_csp_connection",
						ResourceType: "csp_connection",
						Name:         "test-vxc",
						OwnerAccount: "test-owner-account-id",
						Bandwidths:   []int{50},
						ConnectionID: "test-connection-id",
					},
				},
			},
			VLL: &VLLConfig{
				AEndVLAN:      0,
				BEndVLAN:      0,
				RateLimitMBPS: 50,
				ResourceName:  "vll",
				ResourceType:  "vll",
				Shutdown:      false,
			},
		},
		VXCApproval: &VXCApproval{
			Status:   "",
			Message:  "",
			UID:      "",
			Type:     "",
			NewSpeed: 0,
		},
		ContractTermMonths: 1,
		CompanyUID:         companyUid,
		CompanyName:        "Test Company",
		AttributeTags:      map[string]string{},
		Cancelable:         true,
		Shutdown:           false,
		AEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        portUid,
			Name:       "test-port",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "Atlanta",
				Metro:   "Atlanta",
				Country: "USA",
			},
		},
		BEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        bEndUid,
			Name:       "Test Product",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "New York",
				Metro:   "New York",
				Country: "USA",
			},
		}}

	jblob := `{
		"message": "Found Product 6b3f68e-2f54-4331-bf94-f8984449365f",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productId": 1,
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"productName": "test-vxc",
			"productType": "VXC",
			"rateLimit": 50,
			"distanceBand": "ZONE",
			"provisioningStatus": "LIVE",
			"usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
			"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"createDate": 1706104800000,
			"resources": {
				"csp_connection": [{
					"bandwidth": 50,
					"connectType": "AWSHC",
					"resource_name": "b_csp_connection",
					"resource_type": "csp_connection",
					"name": "test-vxc",
					"ownerAccount": "test-owner-account-id",
					"bandwidths": [
						50
					],
					"connectionId": "test-connection-id"
				}],
				"vll": {
					"a_vlan": 0,
					"b_vlan": 0,
					"rate_limit_mbps": 50,
					"resource_name": "vll",
					"resource_type": "vll",
					"up": 0,
					"shutdown": false
				}
			},
			"vxcApproval": {
				"status": null,
				"message": null,
				"uid": null,
				"type": null,
				"newSpeed": null
			},
			"contractStartDate": 1706104800000,
			"contractEndDate": 1737727200000,
			"contractTermMonths": 1,
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"companyName": "Test Company",
			"locked": false,
			"adminLocked": false,
			"attributeTags": {},
			"up": false,
			"shutdown": false,
			"cancelable": true,
			"aEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "9b1c46c7-1e8d-4035-bf38-1bc60d346d57",
				"productName": "test-port",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "Atlanta",
					"metro": "Atlanta",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			},
			"bEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "91ededc2-473f-4a30-ad24-0703c7f35e50",
				"productName": "Test Product",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "New York",
					"metro": "New York",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			}
		}
	}`
	path := "/v2/product/" + vxcUid
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	gotVxc, err := vxcSvc.GetVXC(ctx, vxcUid)
	suite.NoError(err)
	suite.Equal(wantVxc, gotVxc)
}

// TestGetAzureVXC tests the GetVXC method for an Azure VXC.
func (suite *VXCClientTestSuite) TestGetAzureVXC() {
	ctx := context.Background()
	vxcSvc := suite.client.VXCService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	vxcUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	portUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	bEndUid := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	wantVxc := &VXC{
		ID:                 1,
		UID:                vxcUid,
		Name:               "test-vxc",
		Type:               "VXC",
		RateLimit:          50,
		DistanceBand:       "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy:          companyUid,
		CreateDate:         startDate,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		Resources: &VXCResources{
			CSPConnection: &CSPConnection{
				CSPConnection: []CSPConnectionConfig{
					CSPConnectionAzure{
						Bandwidth:    50,
						ConnectType:  "AZURE",
						ResourceName: "b_csp_connection",
						ResourceType: "csp_connection",
						Managed:      true,
						VLAN:         0,
						Megaports: []CSPConnectionAzureMegaport{
							{
								Port: 0,
								Type: "MEGAPORT",
								VXC:  1,
							},
						},
						Ports: []CSPConnectionAzurePort{{
							ServiceID:     1,
							Type:          "PORT",
							VXCServiceIDs: []int{1},
						},
						},
						ServiceKey: "test-service-key",
					},
				},
			},
			VLL: &VLLConfig{
				AEndVLAN:      0,
				BEndVLAN:      0,
				RateLimitMBPS: 50,
				ResourceName:  "vll",
				ResourceType:  "vll",
			},
		},
		VXCApproval: &VXCApproval{
			Status:   "",
			Message:  "",
			UID:      "",
			Type:     "",
			NewSpeed: 0,
		},
		ContractTermMonths: 1,
		CompanyUID:         companyUid,
		CompanyName:        "Test Company",
		AttributeTags:      map[string]string{},
		Cancelable:         true,
		AEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        portUid,
			Name:       "test-port",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "Atlanta",
				Metro:   "Atlanta",
				Country: "USA",
			},
		},
		BEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        bEndUid,
			Name:       "Test Product",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "New York",
				Metro:   "New York",
				Country: "USA",
			},
		}}

	jblob := `{
		"message": "Found Product 6b3f68e-2f54-4331-bf94-f8984449365f",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productId": 1,
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"productName": "test-vxc",
			"productType": "VXC",
			"rateLimit": 50,
			"distanceBand": "ZONE",
			"provisioningStatus": "LIVE",
			"usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
			"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"createDate": 1706104800000,
			"resources": {
				"csp_connection": [{
					"bandwidth": 50,
					"connectType": "AZURE",
					"resource_name": "b_csp_connection",
					"resource_type": "csp_connection",
					"vlan": 0,
					"managed": true,
					"megaports": [{
						"port": 0,
						"type": "MEGAPORT",
						"vxc": 1
					}],
					"ports": [{
						"service_id": 1,
						"type": "PORT",
						"vxc_service_ids": [1]
					}],
					"service_key": "test-service-key"
				}],
				"vll": {
					"a_vlan": 0,
					"b_vlan": 0,
					"rate_limit_mbps": 50,
					"resource_name": "vll",
					"resource_type": "vll",
					"up": 0,
					"shutdown": false
				}
			},
			"vxcApproval": {
				"status": null,
				"message": null,
				"uid": null,
				"type": null,
				"newSpeed": null
			},
			"contractStartDate": 1706104800000,
			"contractEndDate": 1737727200000,
			"contractTermMonths": 1,
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"companyName": "Test Company",
			"locked": false,
			"adminLocked": false,
			"attributeTags": {},
			"up": false,
			"shutdown": false,
			"cancelable": true,
			"aEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "9b1c46c7-1e8d-4035-bf38-1bc60d346d57",
				"productName": "test-port",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "Atlanta",
					"metro": "Atlanta",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			},
			"bEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "91ededc2-473f-4a30-ad24-0703c7f35e50",
				"productName": "Test Product",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "New York",
					"metro": "New York",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			}
		}
	}`
	path := "/v2/product/" + vxcUid
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	gotVxc, err := vxcSvc.GetVXC(ctx, vxcUid)
	suite.NoError(err)
	suite.Equal(wantVxc, gotVxc)
}

// TestGetGoogleVXC tests the GetVXC method for a Google VXC.
func (suite *VXCClientTestSuite) TestGetGoogleVXC() {
	ctx := context.Background()
	vxcSvc := suite.client.VXCService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	vxcUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	portUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	bEndUid := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	wantVxc := &VXC{
		ID:                 1,
		UID:                vxcUid,
		Name:               "test-vxc",
		Type:               "VXC",
		RateLimit:          50,
		DistanceBand:       "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy:          companyUid,
		CreateDate:         startDate,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		Resources: &VXCResources{
			CSPConnection: &CSPConnection{
				CSPConnection: []CSPConnectionConfig{
					CSPConnectionGoogle{
						Bandwidth:    50,
						Bandwidths:   []int{50},
						ConnectType:  "GOOGLE",
						ResourceName: "b_csp_connection",
						ResourceType: "csp_connection",
						CSPName:      "GOOGLE",
						Megaports: []CSPConnectionGoogleMegaport{
							{
								Port: 0,
								VXC:  1,
							},
						},
						Ports: []CSPConnectionGooglePort{{
							ServiceID:     1,
							VXCServiceIDs: []int{1},
						},
						},
						PairingKey: "test-pairing-key",
					},
				},
			},
			VLL: &VLLConfig{
				AEndVLAN:      0,
				BEndVLAN:      0,
				RateLimitMBPS: 50,
				ResourceName:  "vll",
				ResourceType:  "vll",
			},
		},
		VXCApproval: &VXCApproval{
			Status:   "",
			Message:  "",
			UID:      "",
			Type:     "",
			NewSpeed: 0,
		},
		ContractTermMonths: 1,
		CompanyUID:         companyUid,
		CompanyName:        "Test Company",
		AttributeTags:      map[string]string{},
		Cancelable:         true,
		CostCentre:         "",
		AEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        portUid,
			Name:       "test-port",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "Atlanta",
				Metro:   "Atlanta",
				Country: "USA",
			},
		},
		BEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        bEndUid,
			Name:       "Test Product",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       0,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "New York",
				Metro:   "New York",
				Country: "USA",
			},
		}}

	jblob := `{
		"message": "Found Product 6b3f68e-2f54-4331-bf94-f8984449365f",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productId": 1,
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"productName": "test-vxc",
			"productType": "VXC",
			"rateLimit": 50,
			"distanceBand": "ZONE",
			"provisioningStatus": "LIVE",
			"usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
			"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"createDate": 1706104800000,
			"resources": {
				"csp_connection": [{
					"bandwidth": 50,
					"bandwidths": [50],
					"connectType": "GOOGLE",
					"csp_name": "GOOGLE",
					"resource_name": "b_csp_connection",
					"resource_type": "csp_connection",
					"megaports": [{
						"port": 0,
						"vxc": 1
					}],
					"ports": [{
						"service_id": 1,
						"vxc_service_ids": [1]
					}],
					"pairingKey": "test-pairing-key"
				}],
				"vll": {
					"a_vlan": 0,
					"b_vlan": 0,
					"rate_limit_mbps": 50,
					"resource_name": "vll",
					"resource_type": "vll",
					"up": 0,
					"shutdown": false
				}
			},
			"vxcApproval": {
				"status": null,
				"message": null,
				"uid": null,
				"type": null,
				"newSpeed": null
			},
			"contractStartDate": 1706104800000,
			"contractEndDate": 1737727200000,
			"contractTermMonths": 1,
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"companyName": "Test Company",
			"costCentre": "",
			"locked": false,
			"adminLocked": false,
			"attributeTags": {},
			"up": false,
			"shutdown": false,
			"cancelable": true,
			"aEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "9b1c46c7-1e8d-4035-bf38-1bc60d346d57",
				"productName": "test-port",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "Atlanta",
					"metro": "Atlanta",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			},
			"bEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "91ededc2-473f-4a30-ad24-0703c7f35e50",
				"productName": "Test Product",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "New York",
					"metro": "New York",
					"country": "USA"
				},
				"vlan": 0,
				"innerVlan": null,
				"secondaryName": null
			}
		}
	}`
	path := "/v2/product/" + vxcUid
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	gotVxc, err := vxcSvc.GetVXC(ctx, vxcUid)
	suite.NoError(err)
	suite.Equal(wantVxc, gotVxc)
}

// TestUpdateVXC tests the UpdateVXC method.
func (suite *VXCClientTestSuite) TestUpdateVXC() {
	ctx := context.Background()
	vxcSvc := suite.client.VXCService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	vxcUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	portUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	bEndUid := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	updateName := "test-vxc-updated"
	aEndVlan := 1
	bEndVlan := 1
	rateLimit := 100
	costCentre := "test-cost-centre"
	shutdown := false
	updatedTerms := 12

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	updateReq := &UpdateVXCRequest{
		Name:       &updateName,
		AEndVLAN:   &aEndVlan,
		BEndVLAN:   &bEndVlan,
		RateLimit:  &rateLimit,
		Term:       &updatedTerms,
		Shutdown:   &shutdown,
		CostCentre: &costCentre,
	}

	jblob := `{
		"message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productId": 1,
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"productName": "test-vxc-updated",
			"productType": "VXC",
			"rateLimit": 100,
			"distanceBand": "ZONE",
			"provisioningStatus": "LIVE",
			"usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
			"createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"createDate": 1706104800000,
			"resources": {
				"csp_connection": {
					"bandwidth": 50,
					"connectType": "AWSHC",
					"resource_name": "b_csp_connection",
					"resource_type": "csp_connection",
					"name": "test-vxc",
					"ownerAccount": "test-owner-account-id",
					"bandwidths": [
						50
					],
					"connectionId": "test-connection-id"
				},
				"vll": {
					"a_vlan": 0,
					"b_vlan": 0,
					"rate_limit_mbps": 100,
					"resource_name": "vll",
					"resource_type": "vll",
					"up": 0,
					"shutdown": false
				}
			},
			"vxcApproval": {
				"status": null,
				"message": null,
				"uid": null,
				"type": null,
				"newSpeed": null
			},
			"contractStartDate": 1706104800000,
			"contractEndDate": 1737727200000,
			"contractTermMonths": 12,
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"costCentre": "test-cost-centre",
			"companyName": "Test Company",
			"locked": false,
			"adminLocked": false,
			"attributeTags": {},
			"up": false,
			"shutdown": false,
			"cancelable": true,
			"aEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "9b1c46c7-1e8d-4035-bf38-1bc60d346d57",
				"productName": "test-port",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "Atlanta",
					"metro": "Atlanta",
					"country": "USA"
				},
				"vlan": 1,
				"innerVlan": null,
				"secondaryName": null
			},
			"bEnd": {
				"ownerUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
				"productUid": "91ededc2-473f-4a30-ad24-0703c7f35e50",
				"productName": "Test Product",
				"locationId": 1,
				"location": "Test Location",
				"locationDetail": {
					"name": "Test Location",
					"city": "New York",
					"metro": "New York",
					"country": "USA"
				},
				"vlan": 1,
				"innerVlan": null,
				"secondaryName": null
			}
		}
	}`

	wantVxc := &VXC{
		ID:                 1,
		UID:                vxcUid,
		Name:               updateName,
		Type:               "VXC",
		RateLimit:          100,
		DistanceBand:       "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy:          companyUid,
		CreateDate:         startDate,
		ContractStartDate:  startDate,
		ContractEndDate:    endDate,
		CostCentre:         costCentre,
		Resources: &VXCResources{
			CSPConnection: &CSPConnection{
				CSPConnection: []CSPConnectionConfig{
					CSPConnectionAWSHC{
						Bandwidth:    50,
						ConnectType:  "AWSHC",
						ResourceName: "b_csp_connection",
						ResourceType: "csp_connection",
						Name:         "test-vxc",
						OwnerAccount: "test-owner-account-id",
						Bandwidths:   []int{50},
						ConnectionID: "test-connection-id",
					},
				},
			},
			VLL: &VLLConfig{
				AEndVLAN:      0,
				BEndVLAN:      0,
				RateLimitMBPS: 100,
				ResourceName:  "vll",
				ResourceType:  "vll",
				Shutdown:      false,
			},
		},
		VXCApproval: &VXCApproval{
			Status:   "",
			Message:  "",
			UID:      "",
			Type:     "",
			NewSpeed: 0,
		},
		ContractTermMonths: 12,
		CompanyUID:         companyUid,
		CompanyName:        "Test Company",
		AttributeTags:      map[string]string{},
		Cancelable:         true,
		AEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        portUid,
			Name:       "test-port",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       1,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "Atlanta",
				Metro:   "Atlanta",
				Country: "USA",
			},
		},
		BEndConfiguration: VXCEndConfiguration{
			OwnerUID:   companyUid,
			UID:        bEndUid,
			Name:       "Test Product",
			LocationID: 1,
			Location:   "Test Location",
			VLAN:       1,
			LocationDetails: &ProductLocationDetails{
				Name:    "Test Location",
				City:    "New York",
				Metro:   "New York",
				Country: "USA",
			},
		},
	}
	update := &VXCUpdate{
		Name:       *updateReq.Name,
		RateLimit:  updateReq.RateLimit,
		AEndVLAN:   updateReq.AEndVLAN,
		BEndVLAN:   updateReq.BEndVLAN,
		Shutdown:   updateReq.Shutdown,
		CostCentre: updateReq.CostCentre,
		Term:       updateReq.Term,
	}
	path := fmt.Sprintf("/v3/product/%s/%s", PRODUCT_VXC, vxcUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(VXCUpdate)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %V", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(update, v)
	})
	gotVxc, err := vxcSvc.UpdateVXC(ctx, vxcUid, updateReq)
	suite.NoError(err)
	suite.Equal(wantVxc, gotVxc)
}

// TestDeleteVXC tests the DeleteVXC method.
func (suite *VXCClientTestSuite) TestDeleteVXC() {
	ctx := context.Background()

	vxcSvc := suite.client.VXCService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	req := &DeleteVXCRequest{
		DeleteNow: true,
	}

	jblob := `{
		"message": "Action [CANCEL_NOW Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	path := "/v3/product/" + productUid + "/action/CANCEL_NOW"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	err := vxcSvc.DeleteVXC(ctx, productUid, req)

	suite.NoError(err)
}

// TestDeleteVXC tests to see if the custom unmarshalling works for decommed VXCs.
func (suite *VXCClientTestSuite) TestDecomissionedVXCMarshal() {
	ctx := context.Background()

	vxcSvc := suite.client.VXCService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
		"message": "Found Product 6b3f68e-2f54-4331-bf94-f8984449365f",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productId": 1,
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"productName": "test-vxc",
			"resources": {
				"vll": []
			}
		}
	}`

	path := "/v2/product/" + productUid
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	gotVxc, err := vxcSvc.GetVXC(ctx, productUid)

	wantVxc := &VXC{
		ID:   1,
		UID:  productUid,
		Name: "test-vxc",
		Resources: &VXCResources{
			VLL: nil,
		},
	}

	suite.Equal(wantVxc, gotVxc)
	suite.NoError(err)
}

// TestListVXCs tests the ListVXCs method with various filters
func (suite *VXCClientTestSuite) TestListVXCs() {
	// Define mock responses for products list API
	productsResponse := `{
        "message": "Found 3 Products",
        "terms": "This data is subject to the Acceptable Use Policy",
        "data": [
            {
                "productId": 1001,
                "productUid": "port-test-aaa-111",
                "productName": "Test Source Port",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedVxcs": [
                    {
                        "productId": 2001,
                        "productUid": "vxc-test-123",
                        "productName": "Test VXC 1",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 500,
                        "aEnd": {
                            "productUid": "port-test-aaa-111",
                            "productName": "Test Source Port",
                            "vlan": 100
                        },
                        "bEnd": {
                            "productUid": "port-test-bbb-222",
                            "productName": "Test Destination Port",
                            "vlan": 200
                        }
                    },
                    {
                        "productId": 2002,
                        "productUid": "vxc-test-456",
                        "productName": "Test EXAMPLE VXC",
                        "productType": "VXC",
                        "provisioningStatus": "CONFIGURED",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "port-test-aaa-111",
                            "productName": "Test Source Port",
                            "vlan": 101
                        },
                        "bEnd": {
                            "productUid": "mcr-test-ccc-333",
                            "productName": "Test MCR",
                            "vlan": 301
                        }
                    }
                ]
            },
            {
                "productId": 1002,
                "productUid": "mcr-test-ccc-333",
                "productName": "Test MCR",
                "productType": "MCR2",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedVxcs": [
                    {
                        "productId": 2002,
                        "productUid": "vxc-test-456",
                        "productName": "Test EXAMPLE VXC",
                        "productType": "VXC",
                        "provisioningStatus": "CONFIGURED",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "port-test-aaa-111",
                            "productName": "Test Source Port",
                            "vlan": 101
                        },
                        "bEnd": {
                            "productUid": "mcr-test-ccc-333",
                            "productName": "Test MCR",
                            "vlan": 301
                        }
                    }
                ]
            },
            {
                "productId": 1003,
                "productUid": "port-test-ddd-444",
                "productName": "Test Other Port",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedVxcs": [
                    {
                        "productId": 2003,
                        "productUid": "vxc-test-789",
                        "productName": "Inactive Test VXC",
                        "productType": "VXC",
                        "provisioningStatus": "CANCELLED",
                        "rateLimit": 500,
                        "aEnd": {
                            "productUid": "port-test-ddd-444",
                            "productName": "Test Other Port",
                            "vlan": 102
                        },
                        "bEnd": {
                            "productUid": "port-test-eee-555",
                            "productName": "Test Another Port",
                            "vlan": 401
                        }
                    }
                ]
            }
        ]
    }`

	// Set up mock handler for products list API
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(productsResponse))
		if err != nil {
			suite.FailNowf("could not write response", "could not write response %v", err)
		}
	})

	// Test cases for ListVXCs with different filters
	tests := []struct {
		name            string
		request         *ListVXCsRequest
		expectedCount   int
		expectedVXCUIDs []string
	}{
		{
			name:            "List all active VXCs (default behavior)",
			request:         &ListVXCsRequest{},
			expectedCount:   2,
			expectedVXCUIDs: []string{"vxc-test-123", "vxc-test-456"},
		},
		{
			name:            "List all VXCs including inactive",
			request:         &ListVXCsRequest{IncludeInactive: true},
			expectedCount:   3,
			expectedVXCUIDs: []string{"vxc-test-123", "vxc-test-456", "vxc-test-789"},
		},
		{
			name:            "Filter by exact name",
			request:         &ListVXCsRequest{Name: "Test VXC 1"},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-123"},
		},
		{
			name:            "Filter by name contains",
			request:         &ListVXCsRequest{NameContains: "EXAMPLE"},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-456"},
		},
		{
			name:            "Filter by status",
			request:         &ListVXCsRequest{Status: []string{"CONFIGURED"}},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-456"},
		},
		{
			name:            "Filter by A-End product UID",
			request:         &ListVXCsRequest{AEndProductUID: "port-test-aaa-111"},
			expectedCount:   2,
			expectedVXCUIDs: []string{"vxc-test-123", "vxc-test-456"},
		},
		{
			name:            "Filter by B-End product UID",
			request:         &ListVXCsRequest{BEndProductUID: "mcr-test-ccc-333"},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-456"},
		},
		{
			name:            "Filter by rate limit",
			request:         &ListVXCsRequest{RateLimit: 1000},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-456"},
		},
		{
			name:            "Multiple filters (A-End and rate limit)",
			request:         &ListVXCsRequest{AEndProductUID: "port-test-aaa-111", RateLimit: 500},
			expectedCount:   1,
			expectedVXCUIDs: []string{"vxc-test-123"},
		},
		{
			name:            "Filter with no matches",
			request:         &ListVXCsRequest{Name: "Non-existent VXC"},
			expectedCount:   0,
			expectedVXCUIDs: []string{},
		},
	}

	// Run all test cases
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Execute the test
			result, err := suite.client.VXCService.ListVXCs(context.Background(), tc.request)

			// Assert success
			suite.NoError(err, "Expected no error")
			suite.Equal(tc.expectedCount, len(result), "Expected %d VXCs, got %d", tc.expectedCount, len(result))

			// Verify the correct VXCs were returned
			var actualUIDs []string
			for _, vxc := range result {
				actualUIDs = append(actualUIDs, vxc.UID)
			}

			// Verify each expected UID is in the result
			for _, expectedUID := range tc.expectedVXCUIDs {
				suite.Contains(actualUIDs, expectedUID, "Expected VXC with UID %s in results", expectedUID)
			}
		})
	}
}

// TestListVXCsDeduplication tests that duplicate VXCs are properly deduplicated
func (suite *VXCClientTestSuite) TestListVXCsDeduplication() {
	// Define mock response with duplicated VXC
	productsResponse := `{
        "message": "Found 3 Products",
        "terms": "This data is subject to the Acceptable Use Policy",
        "data": [
            {
                "productId": 3001,
                "productUid": "port-test-aaa-999",
                "productName": "A-End Port",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 4001,
                        "productUid": "vxc-test-duplicate",
                        "productName": "Duplicated VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "port-test-aaa-999",
                            "productName": "A-End Port",
                            "vlan": 100
                        },
                        "bEnd": {
                            "productUid": "port-test-bbb-999",
                            "productName": "B-End Port",
                            "vlan": 200
                        }
                    }
                ]
            },
            {
                "productId": 3002,
                "productUid": "port-test-bbb-999",
                "productName": "B-End Port",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 4001,
                        "productUid": "vxc-test-duplicate",
                        "productName": "Duplicated VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "port-test-aaa-999",
                            "productName": "A-End Port",
                            "vlan": 100
                        },
                        "bEnd": {
                            "productUid": "port-test-bbb-999",
                            "productName": "B-End Port",
                            "vlan": 200
                        }
                    }
                ]
            },
            {
                "productId": 3003,
                "productUid": "mve-test-ccc-999",
                "productName": "MVE Product",
                "productType": "MVE",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 4001,
                        "productUid": "vxc-test-duplicate",
                        "productName": "Duplicated VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "port-test-aaa-999",
                            "productName": "A-End Port",
                            "vlan": 100
                        },
                        "bEnd": {
                            "productUid": "port-test-bbb-999",
                            "productName": "B-End Port",
                            "vlan": 200
                        }
                    }
                ]
            }
        ]
    }`

	// Set up mock handler
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(productsResponse))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	// Execute the test
	result, err := suite.client.VXCService.ListVXCs(context.Background(), nil)

	// Assert success and deduplication
	suite.NoError(err, "Expected no error")
	suite.Equal(1, len(result), "Expected exactly 1 VXC after deduplication")

	if len(result) == 1 {
		suite.Equal("vxc-test-duplicate", result[0].UID, "Expected the duplicated VXC")
		suite.Equal("Duplicated VXC", result[0].Name, "Expected correct VXC name")
	}
}

// TestListVXCsWithDifferentProductTypes tests that VXCs are properly extracted from different product types
func (suite *VXCClientTestSuite) TestListVXCsWithDifferentProductTypes() {
	// Define mock response with VXCs associated with different product types
	productsResponse := `{
        "message": "Found 6 Products",
        "terms": "This data is subject to the Acceptable Use Policy",
        "data": [
            {
                "productId": 5001,
                "productUid": "port-test-src-001",
                "productName": "Source Port",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 6001,
                        "productUid": "vxc-port-123",
                        "productName": "Port VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 500,
                        "aEnd": {
                            "productUid": "port-test-src-001",
                            "productName": "Source Port"
                        },
                        "bEnd": {
                            "productUid": "port-test-dst-001",
                            "productName": "Destination Port"
                        }
                    }
                ]
            },
            {
                "productId": 5003,
                "productUid": "mcr-test-src-001",
                "productName": "Source MCR",
                "productType": "MCR2",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 6002,
                        "productUid": "vxc-mcr-456",
                        "productName": "MCR VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "aEnd": {
                            "productUid": "mcr-test-src-001",
                            "productName": "Source MCR"
                        },
                        "bEnd": {
                            "productUid": "port-test-dst-002",
                            "productName": "Destination Port 2"
                        }
                    }
                ]
            },
            {
                "productId": 5005,
                "productUid": "mve-test-src-001",
                "productName": "Source MVE",
                "productType": "MVE",
                "provisioningStatus": "LIVE",
                "associatedVxcs": [
                    {
                        "productId": 6003,
                        "productUid": "vxc-mve-789",
                        "productName": "MVE VXC",
                        "productType": "VXC",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 250,
                        "aEnd": {
                            "productUid": "mve-test-src-001",
                            "productName": "Source MVE"
                        },
                        "bEnd": {
                            "productUid": "port-test-dst-003",
                            "productName": "Destination Port 3"
                        }
                    }
                ]
            }
        ]
    }`

	// Set up mock handler
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(productsResponse))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	// Execute the test
	result, err := suite.client.VXCService.ListVXCs(context.Background(), nil)

	// Assert success
	suite.NoError(err, "Expected no error")
	suite.Equal(3, len(result), "Expected 3 VXCs from different product types")

	// Check for expected VXC UIDs
	expectedUIDs := map[string]bool{
		"vxc-port-123": false,
		"vxc-mcr-456":  false,
		"vxc-mve-789":  false,
	}

	for _, vxc := range result {
		expectedUIDs[vxc.UID] = true
	}

	// Verify all VXCs were found
	for uid, found := range expectedUIDs {
		suite.True(found, "VXC with UID %s was not found", uid)
	}
}
