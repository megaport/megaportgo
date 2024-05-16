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

// ProductClientTestSuite tests the product client
type ProductClientTestSuite struct {
	ClientTestSuite
}

func TestProductClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ProductClientTestSuite))
}

func (suite *ProductClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *ProductClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestExecuteOrder tests the ExecuteOrder method
func (suite *ProductClientTestSuite) TestExecuteOrder() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
		]
	}`

	portOrder := []PortOrder{
		{
			Name:                  "test-port",
			Term:                  12,
			PortSpeed:             10000,
			LocationID:            226,
			Virtual:               false,
			Market:                "US",
			MarketplaceVisibility: false,
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
	wantRes := PtrTo([]byte(jblob))
	gotRes, err := productSvc.ExecuteOrder(ctx, portOrder)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}

// TestModifyProduct tests the ModifyProduct method
func (suite *ProductClientTestSuite) TestModifyProduct() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
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
	productType := PRODUCT_MEGAPORT
	wantReq := &ModifyProductRequest{
		ProductID:             productUid,
		ProductType:           productType,
		Name:                  "updated-test-product",
		CostCentre:            "US",
		MarketplaceVisibility: PtrTo(false),
	}
	path := fmt.Sprintf("/v2/product/%s/%s", productType, productUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(ModifyProductRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could  not decode json %v", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(wantReq, v)
	})
	wantRes := &ModifyProductResponse{
		IsUpdated: true,
	}
	gotRes, err := productSvc.ModifyProduct(ctx, wantReq)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}

// TestDeleteProduct tests the DeleteProduct method
func (suite *ProductClientTestSuite) TestDeleteProduct() {
	ctx := context.Background()

	productSvc := suite.client.ProductService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
    "message": "Action [CANCEL_NOW Service 36b3f68e-2f54-4331-bf94-f8984449365f] has been done.",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy"
	}`

	req := &DeleteProductRequest{
		ProductID: productUid,
		DeleteNow: true,
	}

	path := "/v3/product/" + req.ProductID + "/action/CANCEL_NOW"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	wantRes := &DeleteProductResponse{}

	gotRes, err := productSvc.DeleteProduct(ctx, req)

	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}

// TestRestoreProduct tests the RestoreProduct method
func (suite *ProductClientTestSuite) TestRestoreProduct() {
	ctx := context.Background()

	productSvc := suite.client.ProductService
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

	wantRes := &RestoreProductResponse{}

	gotRes, err := productSvc.RestoreProduct(ctx, productUid)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}

// TestManageProductLuck tests the ManageProductLock method
func (suite *ProductClientTestSuite) TestManageProductLuck() {
	ctx := context.Background()

	productSvc := suite.client.ProductService
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

	req := &ManageProductLockRequest{
		ProductID:  productUid,
		ShouldLock: true,
	}

	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s/lock", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	wantRes := &ManageProductLockResponse{}

	gotRes, err := productSvc.ManageProductLock(ctx, req)
	suite.NoError(err)
	suite.Equal(wantRes, gotRes)
}
