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

// TestGetProductLOA tests the GetProductLOA method
func (suite *ProductClientTestSuite) TestGetProductLOA() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\n1 0 obj\n<</Type/Catalog/Pages 2 0 R>>\nendobj\n2 0 obj\n<</Type/Pages/Count 1/Kids[3 0 R]>>\nendobj\n3 0 obj\n<</Type/Page/MediaBox[0 0 612 792]/Resources<<>>>>\nendobj\nxref\n0 4\n0000000000 65535 f\n0000000010 00000 n\n0000000053 00000 n\n0000000102 00000 n\ntrailer\n<</Size 4/Root 1 0 R>>\nstartxref\n149\n%%EOF")

	path := fmt.Sprintf("/v2/product/%s/loa", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		// Set appropriate headers for PDF content
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=\"LOA_"+productUID+".pdf\"")
		w.Write(mockPDFContent)
	})

	// Test the GetProductLOA function
	loaContent, err := productSvc.GetProductLOA(ctx, productUID)
	suite.NoError(err)
	suite.Equal(mockPDFContent, loaContent)
}

// TestSaveProductLOAToFile tests the SaveProductLOAToFile method
func (suite *ProductClientTestSuite) TestSaveProductLOAToFile() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\nTest LOA Content")

	path := fmt.Sprintf("/v2/product/%s/loa", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDFContent)
	})

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "loa_test_*.pdf")
	suite.NoError(err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test saving the LOA to a file
	err = productSvc.SaveProductLOAToFile(ctx, productUID, tempFile.Name())
	suite.NoError(err)

	// Verify the file content matches the mock PDF
	savedContent, err := os.ReadFile(tempFile.Name())
	suite.NoError(err)
	suite.Equal(mockPDFContent, savedContent)
}

// TestGetProductLOABase64 tests the GetProductLOABase64 method
func (suite *ProductClientTestSuite) TestGetProductLOABase64() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	// Mock PDF content for testing
	mockPDFContent := []byte("%PDF-1.5\nTest LOA Content")
	expectedBase64 := base64.StdEncoding.EncodeToString(mockPDFContent)

	path := fmt.Sprintf("/v2/product/%s/loa", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDFContent)
	})

	// Test the base64 encoding function
	base64Content, err := productSvc.GetProductLOABase64(ctx, productUID)
	suite.NoError(err)
	suite.Equal(expectedBase64, base64Content)

	// Verify the decoded content matches the original
	decodedContent, err := base64.StdEncoding.DecodeString(base64Content)
	suite.NoError(err)
	suite.Equal(mockPDFContent, decodedContent)
}

// TestGetProductLOAError tests error handling in the GetProductLOA method
func (suite *ProductClientTestSuite) TestGetProductLOAError() {
	ctx := context.Background()
	productSvc := suite.client.ProductService
	productUID := "non-existent-product-uid"

	path := fmt.Sprintf("/v2/product/%s/loa", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Product not found"}`)
	})

	// Test error handling
	_, err := productSvc.GetProductLOA(ctx, productUID)
	suite.Error(err)
	suite.Contains(err.Error(), "404")
	suite.Contains(err.Error(), "Product not found")
}
