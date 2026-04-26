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

	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		var wrapper struct {
			NetworkDesign []PortOrder `json:"networkDesign"`
			DiscountCodes []string    `json:"discountCodes"`
		}
		err := json.NewDecoder(r.Body).Decode(&wrapper)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := wrapper.NetworkDesign
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
		suite.Equal([]string{}, wrapper.DiscountCodes)
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

func pricebookJSONResponse(productType string) string {
	return `{
		"message": "OK",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productType": "` + productType + `",
			"currency": "USD",
			"monthlyRate": 100.0,
			"monthlyRackRate": 150.0,
			"prices": [{"chargeReason": "CORE", "frequency": "MONTHLY", "amount": 100.0}],
			"discounts": []
		}
	}`
}

// TestGetProductPricingVXC tests pricing for a VXC and verifies productType is injected.
func (suite *ProductClientTestSuite) TestGetProductPricingVXC() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("VXC", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("VXC"))
	})

	got, err := productSvc.GetProductPricing(ctx, &VXCPriceBookRequest{
		ALocationID: 1,
		BLocationID: 2,
		Speed:       1000,
	})
	suite.NoError(err)
	suite.NotNil(got)
	suite.Equal("USD", got.Currency)
	suite.Equal(100.0, got.MonthlyRate)
	suite.Equal(150.0, got.MonthlyRackRate)
	suite.Len(got.Prices, 1)
	suite.Equal(PriceBookChargeReasonCore, got.Prices[0].ChargeReason)
	suite.Equal(PricingFrequencyMonthly, got.Prices[0].Frequency)
}

// TestGetProductPricingMCR tests that MCR pricing sends MCR2 as productType.
func (suite *ProductClientTestSuite) TestGetProductPricingMCR() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("MCR2", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("MCR2"))
	})

	got, err := productSvc.GetProductPricing(ctx, &MCRPriceBookRequest{
		LocationID: 1,
		Speed:      1000,
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingPort tests that Port pricing sends MEGAPORT as productType.
func (suite *ProductClientTestSuite) TestGetProductPricingPort() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("MEGAPORT", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("MEGAPORT"))
	})

	got, err := productSvc.GetProductPricing(ctx, &MegaportPriceBookRequest{
		LocationID: 1,
		Speed:      1000,
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingMVE tests MVE pricing.
func (suite *ProductClientTestSuite) TestGetProductPricingMVE() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("MVE", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("MVE"))
	})

	got, err := productSvc.GetProductPricing(ctx, &MVEPriceBookRequest{
		LocationID: 1,
		MVELabel:   "MVE_2_8",
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingIX tests IX pricing.
func (suite *ProductClientTestSuite) TestGetProductPricingIX() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("IX", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("IX"))
	})

	got, err := productSvc.GetProductPricing(ctx, &IXPriceBookRequest{
		PortLocationID: 1,
		IXType:         "Brisbane IX",
		Speed:          1000,
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingNATGateway tests NAT Gateway pricing.
func (suite *ProductClientTestSuite) TestGetProductPricingNATGateway() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("NAT_GATEWAY", body["productType"])
		fmt.Fprint(w, pricebookJSONResponse("NAT_GATEWAY"))
	})

	got, err := productSvc.GetProductPricing(ctx, &NATGatewayPriceBookRequest{
		LocationID:   1,
		Speed:        1000,
		SessionCount: 100,
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingIPAddress tests IP Address pricing.
func (suite *ProductClientTestSuite) TestGetProductPricingIPAddress() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		var body map[string]interface{}
		suite.NoError(json.NewDecoder(r.Body).Decode(&body))
		suite.Equal("IP_ADDRESS", body["productType"])
		suite.Equal("/24", body["ipBlock"])
		fmt.Fprint(w, pricebookJSONResponse("IP_ADDRESS"))
	})

	got, err := productSvc.GetProductPricing(ctx, &IPAddressPriceBookRequest{
		LocationID: 1,
		IPBlock:    "/24",
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingWithDiscount tests that discount details are decoded correctly.
func (suite *ProductClientTestSuite) TestGetProductPricingWithDiscount() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	jblob := `{
		"message": "OK",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"productType": "VXC",
			"currency": "USD",
			"monthlyRate": 90.0,
			"monthlyRackRate": 100.0,
			"prices": [{"chargeReason": "CORE", "frequency": "MONTHLY", "amount": 100.0}],
			"discounts": [
				{
					"discountReason": "TERM",
					"amount": 10.0,
					"discountDetails": {
						"uid": "ba79a52a-c900-4b82-9897-a283d51840b6",
						"code": "TERM12",
						"description": "12-month term discount",
						"percentageAmount": 10.0,
						"shared": false
					}
				}
			]
		}
	}`

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	got, err := productSvc.GetProductPricing(ctx, &VXCPriceBookRequest{
		ALocationID: 1,
		BLocationID: 2,
		Speed:       1000,
	})
	suite.NoError(err)
	suite.NotNil(got)
	suite.Len(got.Discounts, 1)
	suite.Equal(DiscountReasonTerm, got.Discounts[0].DiscountReason)
	suite.Equal(10.0, got.Discounts[0].Amount)
	suite.NotNil(got.Discounts[0].DiscountDetails)
	suite.Equal("TERM12", got.Discounts[0].DiscountDetails.Code)
	suite.Equal("12-month term discount", got.Discounts[0].DiscountDetails.Description)
	suite.Equal(10.0, got.Discounts[0].DiscountDetails.PercentageAmount)
}

// TestGetProductPricingForCompany tests that companyId is passed as a query param.
func (suite *ProductClientTestSuite) TestGetProductPricingForCompany() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	suite.mux.HandleFunc("/v4/pricebook/product", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		suite.Equal("42", r.URL.Query().Get("companyId"))
		fmt.Fprint(w, pricebookJSONResponse("VXC"))
	})

	got, err := productSvc.GetProductPricingForCompany(ctx, &GetProductPricingRequest{
		Req: &VXCPriceBookRequest{
			ALocationID: 1,
			BLocationID: 2,
			Speed:       1000,
		},
		CompanyID: 42,
	})
	suite.NoError(err)
	suite.NotNil(got)
}

// TestGetProductPricingForCompanyNilRequest tests that GetProductPricingForCompany rejects a nil wrapper.
func (suite *ProductClientTestSuite) TestGetProductPricingForCompanyNilRequest() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	got, err := productSvc.GetProductPricingForCompany(ctx, nil)
	suite.ErrorIs(err, ErrPricingRequestNil)
	suite.Nil(got)
}

// TestGetProductPricingForCompanyNilInnerRequest tests that GetProductPricingForCompany rejects
// a wrapper containing a nil PriceBookRequest, including the typed-nil case.
func (suite *ProductClientTestSuite) TestGetProductPricingForCompanyNilInnerRequest() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	got, err := productSvc.GetProductPricingForCompany(ctx, &GetProductPricingRequest{Req: nil, CompanyID: 1})
	suite.ErrorIs(err, ErrPricingRequestNil)
	suite.Nil(got)

	var typedNil *VXCPriceBookRequest
	got, err = productSvc.GetProductPricingForCompany(ctx, &GetProductPricingRequest{Req: typedNil, CompanyID: 1})
	suite.ErrorIs(err, ErrPricingRequestNil)
	suite.Nil(got)
}

// TestGetProductPricingForCompanyMutuallyExclusiveCompanyIdentifiers tests that setting both
// CompanyID and CompanyUID returns an error rather than sending both query parameters.
func (suite *ProductClientTestSuite) TestGetProductPricingForCompanyMutuallyExclusiveCompanyIdentifiers() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	got, err := productSvc.GetProductPricingForCompany(ctx, &GetProductPricingRequest{
		Req: &VXCPriceBookRequest{
			ALocationID: 1,
			BLocationID: 2,
			Speed:       1000,
		},
		CompanyID:  42,
		CompanyUID: "abc-123",
	})
	suite.ErrorIs(err, ErrPricingCompanyIDAndUIDSet)
	suite.Nil(got)
}

// TestGetProductPricingNilRequest tests that GetProductPricing returns an error for a nil request.
func (suite *ProductClientTestSuite) TestGetProductPricingNilRequest() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	got, err := productSvc.GetProductPricing(ctx, nil)
	suite.ErrorIs(err, ErrPricingRequestNil)
	suite.Nil(got)
}

// TestGetProductPricingTypedNilRequest tests that GetProductPricing catches a typed-nil request,
// which passes the req == nil interface check but wraps a nil concrete pointer.
func (suite *ProductClientTestSuite) TestGetProductPricingTypedNilRequest() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	var typedNil *VXCPriceBookRequest
	got, err := productSvc.GetProductPricing(ctx, typedNil)
	suite.ErrorIs(err, ErrPricingRequestNil)
	suite.Nil(got)
}

// TestGetProductPricingValidation tests required-field validation per request type.
func (suite *ProductClientTestSuite) TestGetProductPricingValidation() {
	ctx := context.Background()
	productSvc := suite.client.ProductService

	// VXC missing locations
	_, err := productSvc.GetProductPricing(ctx, &VXCPriceBookRequest{Speed: 1000})
	suite.ErrorIs(err, ErrPricingVXCLocationRequired)

	// VXC missing speed
	_, err = productSvc.GetProductPricing(ctx, &VXCPriceBookRequest{ALocationID: 1, BLocationID: 2})
	suite.ErrorIs(err, ErrPricingVXCSpeedRequired)

	// MCR missing location
	_, err = productSvc.GetProductPricing(ctx, &MCRPriceBookRequest{Speed: 1000})
	suite.ErrorIs(err, ErrPricingLocationRequired)

	// MCR missing speed
	_, err = productSvc.GetProductPricing(ctx, &MCRPriceBookRequest{LocationID: 1})
	suite.ErrorIs(err, ErrPricingSpeedRequired)

	// IX missing location
	_, err = productSvc.GetProductPricing(ctx, &IXPriceBookRequest{IXType: "Brisbane IX", Speed: 1000})
	suite.ErrorIs(err, ErrPricingIXLocationRequired)

	// IX missing type
	_, err = productSvc.GetProductPricing(ctx, &IXPriceBookRequest{PortLocationID: 1, Speed: 1000})
	suite.ErrorIs(err, ErrPricingIXTypeRequired)

	// NAT Gateway missing session count
	_, err = productSvc.GetProductPricing(ctx, &NATGatewayPriceBookRequest{LocationID: 1, Speed: 1000})
	suite.ErrorIs(err, ErrPricingNATSessionRequired)

	// IP Address missing block
	_, err = productSvc.GetProductPricing(ctx, &IPAddressPriceBookRequest{LocationID: 1})
	suite.ErrorIs(err, ErrPricingIPBlockRequired)

	// IP Address missing location
	_, err = productSvc.GetProductPricing(ctx, &IPAddressPriceBookRequest{IPBlock: "/24"})
	suite.ErrorIs(err, ErrPricingIPLocationRequired)

	// MVE missing location
	_, err = productSvc.GetProductPricing(ctx, &MVEPriceBookRequest{MVELabel: "MVE_2_8"})
	suite.ErrorIs(err, ErrPricingMVELocationRequired)
}

func (suite *ProductClientTestSuite) TestListProductResourceTags() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	suite.mux.HandleFunc(fmt.Sprintf("/v2/product/%s/tags", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, resourceTagJSONBlob)
	})
	res, err := productSvc.ListProductResourceTags(ctx, productUid)
	suite.NoError(err)
	suite.EqualValues(testProductResourceTags, res)
}
