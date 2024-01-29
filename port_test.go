package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/suite"
)

type PortClientTestSuite struct {
	ClientTestSuite
}

func TestPortClientTestSuite(t *testing.T) {
	suite.Run(t, new(PortClientTestSuite))
}

func (suite *PortClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *PortClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *PortClientTestSuite) TestBuyPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuyPortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsLag:      false,
		LagCount:   0,
		IsPrivate:  true,
	}

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`
	portOrder := []types.PortOrder{
		{
			Name:                  req.Name,
			Term:                  req.Term,
			PortSpeed:             req.PortSpeed,
			LocationID:            req.LocationId,
			Virtual:               false,
			Market:                req.Market,
			MarketplaceVisibility: !req.IsPrivate,
		},
	}
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]types.PortOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := portOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.PortSpeed, gotOrder.PortSpeed)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.Virtual, gotOrder.Virtual)
		suite.Equal(wantOrder.MarketplaceVisibility, gotOrder.MarketplaceVisibility)
	})
	got, err := portSvc.BuyPort(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestBuySinglePort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuySinglePortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsPrivate:  true,
	}

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`
	portOrder := []types.PortOrder{
		{
			Name:                  req.Name,
			Term:                  req.Term,
			PortSpeed:             req.PortSpeed,
			LocationID:            req.LocationId,
			Virtual:               false,
			Market:                req.Market,
			MarketplaceVisibility: !req.IsPrivate,
		},
	}
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]types.PortOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := portOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.PortSpeed, gotOrder.PortSpeed)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.Virtual, gotOrder.Virtual)
		suite.Equal(wantOrder.MarketplaceVisibility, gotOrder.MarketplaceVisibility)
	})
	got, err := portSvc.BuySinglePort(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestBuyLAGPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuyLAGPortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsPrivate:  true,
		LagCount:   2,
	}

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`

	portOrder := []types.PortOrder{
		{
			Name:                  req.Name,
			Term:                  req.Term,
			PortSpeed:             req.PortSpeed,
			LocationID:            req.LocationId,
			Virtual:               false,
			Market:                req.Market,
			MarketplaceVisibility: !req.IsPrivate,
			LagPortCount:          req.LagCount,
		},
	}
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]types.PortOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := portOrder[0]
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
		suite.Equal(wantOrder.Name, gotOrder.Name)
		suite.Equal(wantOrder.Term, gotOrder.Term)
		suite.Equal(wantOrder.PortSpeed, gotOrder.PortSpeed)
		suite.Equal(wantOrder.LocationID, gotOrder.LocationID)
		suite.Equal(wantOrder.Virtual, gotOrder.Virtual)
		suite.Equal(wantOrder.MarketplaceVisibility, gotOrder.MarketplaceVisibility)
		suite.Equal(wantOrder.LagPortCount, gotOrder.LagPortCount)
	})
	got, err := portSvc.BuyLAGPort(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestBuyPortInvalidTerm() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	req := &BuyPortRequest{
		Name:       "test-port-bad-term",
		Term:       37,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsLag:      false,
		LagCount:   0,
		IsPrivate:  true,
	}
	_, err := portSvc.BuyPort(ctx, req)
	suite.Equal(errors.New(mega_err.ERR_TERM_NOT_VALID), err)
}

func (suite *ClientTestSuite) TestListPorts() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	productUid2 := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	productUid3 := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	want1 := &types.Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		Type:                  "MEGAPORT",
		SecondaryName:         "test-secondary-name",
		ProvisioningStatus:    "CONFIGURED",
		PortSpeed:             10000,
		LocationID:            226,
		LAGPrimary:            false,
		Market:                "US",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		Virtual:               false,
		BuyoutPort:            false,
		CompanyName:           "test-company",
		CompanyUID:            companyUid,
		ContractStartDate:     1706104800000,
		ContractEndDate:       1737727200000,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
	}
	want2 := &types.Port{
		ID:                    999998,
		UID:                   productUid2,
		Name:                  "test-port2",
		Type:                  "MEGAPORT",
		SecondaryName:         "test-secondary-name2",
		ProvisioningStatus:    "CONFIGURED",
		PortSpeed:             10000,
		LocationID:            226,
		LAGPrimary:            false,
		Market:                "US",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		Virtual:               false,
		BuyoutPort:            false,
		CompanyName:           "test-company",
		CompanyUID:            companyUid,
		ContractStartDate:     1706104800000,
		ContractEndDate:       1737727200000,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
	}
	want3 := &types.Port{
		ID:                    999997,
		UID:                   productUid3,
		Name:                  "test-port3",
		SecondaryName:         "test-secondary-name3",
		Type:                  "MEGAPORT",
		ProvisioningStatus:    "CONFIGURED",
		PortSpeed:             10000,
		LocationID:            226,
		LAGPrimary:            false,
		Market:                "US",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		Virtual:               false,
		BuyoutPort:            false,
		CompanyName:           "test-company",
		CompanyUID:            companyUid,
		ContractStartDate:     1706104800000,
		ContractEndDate:       1737727200000,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
	}
	wantPorts := []*types.Port{want1, want2, want3}
	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [{
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999998,"productUid":"9b1c46c7-1e8d-4035-bf38-1bc60d346d57","productName":"test-port2","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name2","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999997,"productUid":"91ededc2-473f-4a30-ad24-0703c7f35e50","productName":"test-port3","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name3","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}]
	}`
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := portSvc.ListPorts(ctx)
	suite.NoError(err)
	suite.Equal(wantPorts, got)
}

func (suite *ClientTestSuite) TestGetPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := &types.Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		SecondaryName:         "test-secondary-name",
		ProvisioningStatus:    "CONFIGURED",
		Type:                  "MEGAPORT",
		PortSpeed:             10000,
		LocationID:            226,
		LAGPrimary:            false,
		Market:                "US",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		Virtual:               false,
		BuyoutPort:            false,
		CompanyName:           "test-company",
		CompanyUID:            companyUid,
		ContractStartDate:     1706104800000,
		ContractEndDate:       1737727200000,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
	}

	jblob := `{
			"message": "Found Product 36b3f68e-2f54-4331-bf94-f8984449365f",
			"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
			}
			}`
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := portSvc.GetPort(ctx, &GetPortRequest{
		PortID: productUid,
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestModifyPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &ModifyPortRequest{
		PortID:                productUid,
		Name:                  "updated-test-product",
		CostCentre:            "US",
		MarketplaceVisibility: false,
	}
	jblob := `{
    "message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": {
        "serviceName": "updated-test-product",
        "name": "updated-test-product",
        "secondaryName": null,
        "technicalServiceId": 185927,
        "technicalServiceUid": "ef60d544-00e1-4ccc-bcff-3e2050bface5",
        "requestedDate": 1706202200307,
        "configuredDate": null,
        "currentEstimatedDelivery": null,
        "companyName": "test-company",
        "companyId": 1153,
        "billingContactName": null,
        "billingContactId": null,
        "adminContactName": null,
        "adminContactId": null,
        "technicalContactName": null,
        "technicalContactId": null,
        "salesName": null,
        "salesId": null,
        "billableId": 177726,
        "billableUsageAlgorithm": null,
        "productType": "MEGAPORT",
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
        "createDate": 1706202200307,
        "terminationDate": null,
        "contractStartDate": null,
        "contractTermMonths": 1,
        "rateType": "MONTHLY",
        "trialAgreement": false,
        "payerCompanyId": null,
        "nonPayerCompanyId": null,
        "minimumSpeed": null,
        "maximumSpeed": null,
        "rateLimit": 10000,
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
        "buyoutPort": false,
        "virtual": false,
        "locked": false,
        "adminLocked": false,
        "bgpShutdownDefault": false,
        "originDomain": null
    	}
	}`
	wantReq := &types.ProductUpdate{
		Name:                 req.Name,
		CostCentre:           req.CostCentre,
		MarketplaceVisbility: req.MarketplaceVisibility,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", types.PRODUCT_MEGAPORT, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(types.ProductUpdate)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %V", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
	})
	want := &ModifyPortResponse{
		IsUpdated: true,
	}
	got, err := portSvc.ModifyPort(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestDeletePort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
    "message": "Action [CANCEL_NOW Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	req := &DeletePortRequest{
		PortID:    productUid,
		DeleteNow: true,
	}

	path := "/v3/product/" + req.PortID + "/action/CANCEL_NOW"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &DeletePortResponse{
		IsDeleting: true,
	}

	got, err := portSvc.DeletePort(ctx, req)

	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestRestorePort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
	"message": "Action [UN_CANCEL Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	req := &RestorePortRequest{
		PortID: productUid,
	}

	path := "/v3/product/" + req.PortID + "/action/UN_CANCEL"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &RestorePortResponse{
		IsRestoring: true,
	}

	got, err := portSvc.RestorePort(ctx, req)

	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestLockPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
	"message": "Service locked",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":true,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0
            	}
        	}
    	}
	}`

	req := &LockPortRequest{
		PortID: productUid,
	}

	path := fmt.Sprintf("/v2/product/%s/lock", req.PortID)

	jblobGet := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
			}
			}`
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblobGet)
	})

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	want := &LockPortResponse{
		IsLocking: true,
	}

	got, err := portSvc.LockPort(ctx, req)

	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *ClientTestSuite) TestUnlockPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblobGet := `{
	"message": "Found Product 36b3f68e-2f54-4331-bf94-f8984449365f",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":true,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0
            	}
        	}
    	}
	}`

	req := &UnlockPortRequest{
		PortID: productUid,
	}

	path := fmt.Sprintf("/v2/product/%s/lock", req.PortID)

	jblobUnlock := `{
			"message": "Service unlocked",
			"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
			}
			}`
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblobGet)
	})

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodDelete)
		fmt.Fprint(w, jblobUnlock)
	})

	want := &UnlockPortResponse{
		IsUnlocking: true,
	}

	got, err := portSvc.UnlockPort(ctx, req)

	suite.NoError(err)
	suite.Equal(want, got)
}
