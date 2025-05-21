package megaport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// PortClientTestSuite tests the port service client
type PortClientTestSuite struct {
	ClientTestSuite
}

func TestPortClientTestSuite(t *testing.T) {
	t.Parallel()
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

// TestBuyPort tests the BuyPort method
func (suite *PortClientTestSuite) TestBuyPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	want := &BuyPortResponse{TechnicalServiceUIDs: []string{"36b3f68e-2f54-4331-bf94-f8984449365f"}}

	req := &BuyPortRequest{
		Name:                  "test-port",
		Term:                  12,
		PortSpeed:             10000,
		LocationId:            226,
		Market:                "US",
		LagCount:              0,
		MarketPlaceVisibility: true,
		DiversityZone:         "red",
	}

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`
	portOrder := []PortOrder{
		{
			Name:       req.Name,
			Term:       req.Term,
			PortSpeed:  req.PortSpeed,
			LocationID: req.LocationId,
			Virtual:    false,
			Market:     req.Market,
			Config: PortOrderConfig{
				DiversityZone: req.DiversityZone,
			},
			MarketplaceVisibility: req.MarketPlaceVisibility,
		},
	}
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]PortOrder)
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

// TestBuyPortInvalidTerm tests the BuyPort method with an invalid term
func (suite *PortClientTestSuite) TestBuyPortInvalidTerm() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	req := &BuyPortRequest{
		Name:                  "test-port-bad-term",
		Term:                  37,
		PortSpeed:             10000,
		LocationId:            226,
		Market:                "US",
		LagCount:              0,
		MarketPlaceVisibility: true,
	}
	_, err := portSvc.BuyPort(ctx, req)
	suite.Equal(ErrInvalidTerm, err)
}

// TestListPorts tests the ListPorts method
func (suite *PortClientTestSuite) TestListPorts() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	productUid2 := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	productUid3 := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}
	liveDate := &Time{GetTime(1737728200000)}

	want1 := &Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		Type:                  PRODUCT_MEGAPORT,
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
		CreateDate:            startDate,
		CompanyUID:            companyUid,
		ContractStartDate:     startDate,
		ContractEndDate:       endDate,
		TerminateDate:         endDate,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
		LiveDate:              liveDate,
	}
	want2 := &Port{
		ID:                    999998,
		UID:                   productUid2,
		Name:                  "test-port2",
		Type:                  PRODUCT_MEGAPORT,
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
		CreateDate:            startDate,
		ContractStartDate:     startDate,
		ContractEndDate:       endDate,
		TerminateDate:         endDate,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
		LiveDate:              liveDate,
	}
	want3 := &Port{
		ID:                    999997,
		UID:                   productUid3,
		Name:                  "test-port3",
		SecondaryName:         "test-secondary-name3",
		Type:                  PRODUCT_MEGAPORT,
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
		ContractStartDate:     startDate,
		ContractEndDate:       endDate,
		CreateDate:            startDate,
		TerminateDate:         endDate,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
		LiveDate:              liveDate,
	}
	wantPorts := []*Port{want1, want2, want3}
	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [{
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999998,"productUid":"9b1c46c7-1e8d-4035-bf38-1bc60d346d57","productName":"test-port2","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name2","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999997,"productUid":"91ededc2-473f-4a30-ad24-0703c7f35e50","productName":"test-port3","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name3","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}]
	}`
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := portSvc.ListPorts(ctx)
	suite.NoError(err)
	suite.Equal(wantPorts, got)
}

// TestGetPort tests the GetPort method
func (suite *PortClientTestSuite) TestGetPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	aggregationId := 12345

	startDate := &Time{GetTime(1706104800000)}
	endDate := &Time{GetTime(1737727200000)}
	liveDate := &Time{GetTime(1737728200000)}

	want := &Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		SecondaryName:         "test-secondary-name",
		ProvisioningStatus:    "CONFIGURED",
		Type:                  PRODUCT_MEGAPORT,
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
		CreateDate:            startDate,
		ContractStartDate:     startDate,
		ContractEndDate:       endDate,
		TerminateDate:         endDate,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
		LiveDate:              liveDate,
		LocationDetails: &ProductLocationDetails{
			Name:    "Test Location",
			City:    "Atlanta",
			Metro:   "Atlanta",
			Country: "USA",
		},
		AggregationID: aggregationId,
		LagPortUIDs:   []string{"36b3f68e-2f54-4331-bf94-f8984449365f", "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"},
		LagCount:      2,
	}

	jblob := `{
            "message": "Found Product 36b3f68e-2f54-4331-bf94-f8984449365f",
            "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
            "data": {
            "productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":12345,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"locationDetail":{"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
            }
            }`

	jblobListPorts := `{
            "message": "test-message",
            "terms": "test-terms",
            "data": [{
            "productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":12345,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999998,"productUid":"9b1c46c7-1e8d-4035-bf38-1bc60d346d57","productName":"test-port2","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":1706104800000,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":1737728200000,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name2","lagPrimary":false,"lagId":0,"aggregationId":12345,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}]
    }`

	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblobListPorts)
	})

	got, err := portSvc.GetPort(ctx, productUid)
	suite.NoError(err)
	suite.Equal(want, got)
}

func (suite *PortClientTestSuite) TestModifyPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	req := &ModifyPortRequest{
		PortID:                productUid,
		Name:                  "updated-test-product",
		CostCentre:            "US",
		MarketplaceVisibility: PtrTo(false),
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
        "productType": "megaport",
        "provisioningStatus": "DEPLOYABLE",
        "failedReason": null,
		"locationDetail":{"name":"Test Location","city":"Atlanta","metro":"Atlanta","country":"USA"},
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
	wantReq := &ModifyProductRequest{
		ProductID:             req.PortID,
		ProductType:           PRODUCT_MEGAPORT,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_MEGAPORT, productUid)
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
	want := &ModifyPortResponse{
		IsUpdated: true,
	}
	got, err := portSvc.ModifyPort(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestDeletePort tests the DeletePort method
func (suite *PortClientTestSuite) TestDeletePort() {
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

// TestRestorePort tests the RestorePort method
func (suite *PortClientTestSuite) TestRestorePort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
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

	want := &RestorePortResponse{
		IsRestored: true,
	}

	got, err := portSvc.RestorePort(ctx, productUid)

	suite.NoError(err)
	suite.Equal(want, got)
}

// TestLockPort tests the LockPort method
func (suite *PortClientTestSuite) TestLockPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
	"message": "Service locked",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":true,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0
            	}
        	}
    	}
	}`

	path := fmt.Sprintf("/v2/product/%s/lock", productUid)

	jblobGet := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":1737727200000,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
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

	got, err := portSvc.LockPort(ctx, productUid)

	suite.NoError(err)
	suite.Equal(want, got)
}

// TestUnlockPort tests the UnlockPort method
func (suite *PortClientTestSuite) TestUnlockPort() {
	ctx := context.Background()

	portSvc := suite.client.PortService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblobGet := `{
	"message": "Found Product 36b3f68e-2f54-4331-bf94-f8984449365f",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":true,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0
            	}
        	}
    	}
	}`

	path := fmt.Sprintf("/v2/product/%s/lock", productUid)

	jblobUnlock := `{
			"message": "Service unlocked",
			"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"megaport","provisioningStatus":"LIVE","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
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

	got, err := portSvc.UnlockPort(ctx, productUid)

	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetPortLOA tests the GetPortLOA method
func (suite *PortClientTestSuite) TestGetPortLOA() {
	ctx := context.Background()
	portSvc := suite.client.PortService
	portID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\nMegaport Port LOA Test Content")

	path := fmt.Sprintf("/v2/product/%s/loa", portID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=\"LOA_"+portID+".pdf\"")
		w.Write(mockPDFContent)
	})

	// Test the GetPortLOA function
	loaContent, err := portSvc.GetPortLOA(ctx, portID)
	suite.NoError(err)
	suite.Equal(mockPDFContent, loaContent)
}

// TestSavePortLOAToFile tests the SavePortLOAToFile method
func (suite *PortClientTestSuite) TestSavePortLOAToFile() {
	ctx := context.Background()
	portSvc := suite.client.PortService
	portID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\nMegaport Port LOA Test Content")

	path := fmt.Sprintf("/v2/product/%s/loa", portID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDFContent)
	})

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "port_loa_test_*.pdf")
	suite.NoError(err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test saving the LOA to a file
	err = portSvc.SavePortLOAToFile(ctx, portID, tempFile.Name())
	suite.NoError(err)

	// Verify the file content matches the mock PDF
	savedContent, err := os.ReadFile(tempFile.Name())
	suite.NoError(err)
	suite.Equal(mockPDFContent, savedContent)
}

// TestGetPortLOABase64 tests the GetPortLOABase64 method
func (suite *PortClientTestSuite) TestGetPortLOABase64() {
	ctx := context.Background()
	portSvc := suite.client.PortService
	portID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\nMegaport Port LOA Test Content")
	expectedBase64 := base64.StdEncoding.EncodeToString(mockPDFContent)

	path := fmt.Sprintf("/v2/product/%s/loa", portID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDFContent)
	})

	// Test the base64 encoding function
	base64Content, err := portSvc.GetPortLOABase64(ctx, portID)
	suite.NoError(err)
	suite.Equal(expectedBase64, base64Content)

	// Verify the decoded content matches the original
	decodedContent, err := base64.StdEncoding.DecodeString(base64Content)
	suite.NoError(err)
	suite.Equal(mockPDFContent, decodedContent)
}
