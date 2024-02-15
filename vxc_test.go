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

type VXCClientTestSuite struct {
	ClientTestSuite
}

func TestVXCClientTestSuite(t *testing.T) {
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

func (suite *VXCClientTestSuite) TestBuyVXC() {
	portProductUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	vxcProductUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &BuyVXCRequest{
		PortUID: portProductUid,
		VXCName: "test-vxc",
		RateLimit: 50,
		AEndConfiguration: VXCOrderAEndConfiguration{
			VLAN: 0,
		},
		BEndConfiguration: VXCOrderBEndConfiguration{
			ProductUID: vxcProductUid,
			VLAN: 0,
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
				"costCentre": null,
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
					"contractTermMonths": 1,
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
		ID: 1,
		UID: vxcUid,
		Name: "test-vxc",
		Type: "VXC",
		RateLimit: 50,
		DistanceBand: "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm: "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy: companyUid,
		CreateDate: startDate,
		ContractStartDate: startDate,
		ContractEndDate: endDate,
		Resources: VXCResources{
		CspConnection: map[string]interface{}{
			"bandwidth": float64(50),
			"connectType": "",
			"resource_name": "b_csp_connection",
			"resource_type": "csp_connection",
			"name": "test-vxc",
			"ownerAccount": "test-owner-account-id",
			"bandwidths": []interface{}{float64(50)},
			"connectionId": "test-connection-id",
		},
		VLL: VLLConfig{
			AEndVLAN: 0,
			BEndVLAN: 0,
			RateLimitMBPS: 50,
			ResourceName: "vll",
			ResourceType: "vll",
		},
	},
	VXCApproval: VXCApproval{
		Status: "",
		Message: "",
		UID: "",
		Type: "",
		NewSpeed: 0,
	},
	ContractTermMonths: 1,
	CompanyUID: companyUid,
	CompanyName: "Test Company",
	AttributeTags: map[string]string{},
	Cancelable: true,
	AEndConfiguration: VXCEndConfiguration{
		OwnerUID: companyUid,
		UID: portUid,
		Name: "test-port",
		LocationID: 1,
		Location: "Test Location",
		VLAN: 0,
	},
	BEndConfiguration: VXCEndConfiguration{
		OwnerUID: companyUid,
		UID: bEndUid,
		Name: "Test Product",
		LocationID: 1,
		Location: "Test Location",
		VLAN: 0,
	},
	}

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
				"csp_connection": {
					"bandwidth": 50,
					"connectType": "",
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

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}

	updateReq := &UpdateVXCRequest{
		Name: &updateName,
		AEndVLAN: &aEndVlan,
		BEndVlan: &bEndVlan,
		RateLimit: &rateLimit,
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
					"bandwidth": 100,
					"connectType": "",
					"resource_name": "b_csp_connection",
					"resource_type": "csp_connection",
					"name": "test-vxc",
					"ownerAccount": "test-owner-account-id",
					"bandwidths": [
						100
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
		ID: 1,
		UID: vxcUid,
		Name: updateName,
		Type: "VXC",
		RateLimit: 100,
		DistanceBand: "ZONE",
		ProvisioningStatus: "LIVE",
		UsageAlgorithm: "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		CreatedBy: companyUid,
		CreateDate: startDate,
		ContractStartDate: startDate,
		ContractEndDate: endDate,
		Resources: VXCResources{
		CspConnection: map[string]interface{}{
			"bandwidth": float64(100),
			"connectType": "",
			"resource_name": "b_csp_connection",
			"resource_type": "csp_connection",
			"name": "test-vxc",
			"ownerAccount": "test-owner-account-id",
			"bandwidths": []interface{}{float64(100)},
			"connectionId": "test-connection-id",
		},
		VLL: VLLConfig{
			AEndVLAN: 0,
			BEndVLAN: 0,
			RateLimitMBPS: 100,
			ResourceName: "vll",
			ResourceType: "vll",
		},
	},
	VXCApproval: VXCApproval{
		Status: "",
		Message: "",
		UID: "",
		Type: "",
		NewSpeed: 0,
	},
	ContractTermMonths: 1,
	CompanyUID: companyUid,
	CompanyName: "Test Company",
	AttributeTags: map[string]string{},
	Cancelable: true,
	AEndConfiguration: VXCEndConfiguration{
		OwnerUID: companyUid,
		UID: portUid,
		Name: "test-port",
		LocationID: 1,
		Location: "Test Location",
		VLAN: 1,
	},
	BEndConfiguration: VXCEndConfiguration{
		OwnerUID: companyUid,
		UID: bEndUid,
		Name: "Test Product",
		LocationID: 1,
		Location: "Test Location",
		VLAN: 1,
	},
	}
	update := &VXCUpdate{
		Name:      updateReq.Name,
		RateLimit: updateReq.RateLimit,
		AEndVLAN:  updateReq.AEndVLAN,
		BEndVLAN:  updateReq.BEndVlan,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_VXC, vxcUid)
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