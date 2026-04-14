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

// BillingMarketClientTestSuite tests the BillingMarket service.
type BillingMarketClientTestSuite struct {
	ClientTestSuite
}

func TestBillingMarketClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(BillingMarketClientTestSuite))
}

func (suite *BillingMarketClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url

	// Manually initialize the BillingMarketService for testing
	suite.client.BillingMarketService = NewBillingMarketService(suite.client)
}

func (suite *BillingMarketClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestGetBillingMarkets tests the GetBillingMarkets method.
func (suite *BillingMarketClientTestSuite) TestGetBillingMarkets() {
	ctx := context.Background()
	billingMarketSvc := suite.client.BillingMarketService

	// Mock response data
	wantBillingMarket := &BillingMarket{
		ID:                          5982,
		WellKnownSupplier:           "MEGAPORT_US",
		SupplierName:                "Megaport (USA) Inc",
		CurrencyEnum:                "USD",
		Language:                    "en",
		BillingContactName:          "Best Demo",
		BillingContactEmail:         "bestdemo@megaport.com",
		BillingContactPhone:         "+61 7 12341234",
		Address1:                    "53 Stone ln",
		Postcode:                    "12345",
		Country:                     "US",
		City:                        "Bedrock",
		State:                       "FLINT",
		InvoiceTemplate:             "invoice_US",
		TaxRate:                     0.1,
		EstimateInvoice:             1012.74,
		FirstPartyID:                1558,
		SecondPartyID:               41950,
		AttachInvoiceToEmail:        true,
		Region:                      "US",
		PaymentTermInDays:           30,
		StripeAccountPublishableKey: "pk_test_8rImVXnpU0Uh7LDyJLpYtYJz",
		StripeSupportedBankCurrencies: []string{
			"USD",
		},
		VATExempt:   false,
		Active:      true,
		SecuredHash: "1aa368c4ea6637aee56060efb1be1751",
	}

	// Create the mock response JSON
	jblob := `{
        "message": "Markets for company 41950",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "id": 5982,
                "wellKnownSupplier": "MEGAPORT_US",
                "supplierName": "Megaport (USA) Inc",
                "currencyEnum": "USD",
                "language": "en",
                "billingContactName": "Best Demo",
                "billingContactEmail": "bestdemo@megaport.com",
                "billingContactPhone": "+61 7 12341234",
                "address1": "53 Stone ln",
                "postcode": "12345",
                "country": "US",
                "city": "Bedrock",
                "state": "FLINT",
                "invoiceTemplate": "invoice_US",
                "taxRate": 0.1,
                "estimateInvoice": 1012.74,
                "firstPartyId": 1558,
                "secondPartyId": 41950,
                "attachInvoiceToEmail": true,
                "region": "US",
                "paymentTermInDays": 30,
                "stripeAccountPublishableKey": "pk_test_8rImVXnpU0Uh7LDyJLpYtYJz",
                "stripeSupportedBankCurrencies": [
                    "USD"
                ],
                "vatExempt": false,
                "active": true,
                "securedHash": "1aa368c4ea6637aee56060efb1be1751"
            }
        ]
    }`

	// Set up the HTTP handler
	path := "/v2/market"
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet) // Verify the HTTP method
		fmt.Fprint(w, jblob)                // Return the mock response
	})

	// Call the method under test
	gotBillingMarkets, err := billingMarketSvc.GetBillingMarkets(ctx)

	// Verify the results
	suite.NoError(err)
	suite.Len(gotBillingMarkets, 1)

	// Comparing the first (and only) billing market
	gotBillingMarket := gotBillingMarkets[0]

	// Verify all fields match
	suite.Equal(wantBillingMarket.ID, gotBillingMarket.ID)
	suite.Equal(wantBillingMarket.WellKnownSupplier, gotBillingMarket.WellKnownSupplier)
	suite.Equal(wantBillingMarket.SupplierName, gotBillingMarket.SupplierName)
	suite.Equal(wantBillingMarket.CurrencyEnum, gotBillingMarket.CurrencyEnum)
	suite.Equal(wantBillingMarket.Language, gotBillingMarket.Language)
	suite.Equal(wantBillingMarket.BillingContactName, gotBillingMarket.BillingContactName)
	suite.Equal(wantBillingMarket.BillingContactEmail, gotBillingMarket.BillingContactEmail)
	suite.Equal(wantBillingMarket.BillingContactPhone, gotBillingMarket.BillingContactPhone)
	suite.Equal(wantBillingMarket.Address1, gotBillingMarket.Address1)
	suite.Equal(wantBillingMarket.Postcode, gotBillingMarket.Postcode)
	suite.Equal(wantBillingMarket.Country, gotBillingMarket.Country)
	suite.Equal(wantBillingMarket.City, gotBillingMarket.City)
	suite.Equal(wantBillingMarket.State, gotBillingMarket.State)
	suite.Equal(wantBillingMarket.InvoiceTemplate, gotBillingMarket.InvoiceTemplate)
	suite.Equal(wantBillingMarket.TaxRate, gotBillingMarket.TaxRate)
	suite.Equal(wantBillingMarket.EstimateInvoice, gotBillingMarket.EstimateInvoice)
	suite.Equal(wantBillingMarket.FirstPartyID, gotBillingMarket.FirstPartyID)
	suite.Equal(wantBillingMarket.SecondPartyID, gotBillingMarket.SecondPartyID)
	suite.Equal(wantBillingMarket.AttachInvoiceToEmail, gotBillingMarket.AttachInvoiceToEmail)
	suite.Equal(wantBillingMarket.Region, gotBillingMarket.Region)
	suite.Equal(wantBillingMarket.PaymentTermInDays, gotBillingMarket.PaymentTermInDays)
	suite.Equal(wantBillingMarket.StripeAccountPublishableKey, gotBillingMarket.StripeAccountPublishableKey)
	suite.Equal(wantBillingMarket.StripeSupportedBankCurrencies, gotBillingMarket.StripeSupportedBankCurrencies)
	suite.Equal(wantBillingMarket.VATExempt, gotBillingMarket.VATExempt)
	suite.Equal(wantBillingMarket.Active, gotBillingMarket.Active)
	suite.Equal(wantBillingMarket.SecuredHash, gotBillingMarket.SecuredHash)
}

// TestSetBillingMarket tests the SetBillingMarket method.
func (suite *BillingMarketClientTestSuite) TestSetBillingMarket() {
	ctx := context.Background()
	billingMarketSvc := suite.client.BillingMarketService

	// Create a test request
	address2 := "Suite 123"
	req := &SetBillingMarketRequest{
		CurrencyEnum:        "USD",
		Language:            "en",
		BillingContactName:  "Test Contact",
		BillingContactPhone: "+1 555-123-4567",
		BillingContactEmail: "test@example.com",
		Address1:            "123 Main St",
		Address2:            &address2,
		City:                "Anytown",
		State:               "CA",
		Postcode:            "12345",
		Country:             "US",
		YourPONumber:        "PO-12345",
		TaxNumber:           "Tax-67890",
		FirstPartyID:        FIRST_PARTY_ID_US,
	}

	// Expected response
	supplyID := 12345

	// Mock response JSON
	jblob := fmt.Sprintf(`{
        "message": "Market has been created for company",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
            "supplyId": %d
        }
    }`, supplyID)

	// Set up HTTP handler
	path := "/v2/market"
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost) // Verify the HTTP method

		// Decode the request body to verify it
		var reqBody SetBillingMarketRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			suite.FailNowf("Failed to decode request body", "Error: %v", err)
		}

		// Verify request fields
		suite.Equal(req.CurrencyEnum, reqBody.CurrencyEnum)
		suite.Equal(req.Language, reqBody.Language)
		suite.Equal(req.BillingContactName, reqBody.BillingContactName)
		suite.Equal(req.BillingContactPhone, reqBody.BillingContactPhone)
		suite.Equal(req.BillingContactEmail, reqBody.BillingContactEmail)
		suite.Equal(req.Address1, reqBody.Address1)
		suite.Equal(*req.Address2, *reqBody.Address2)
		suite.Equal(req.City, reqBody.City)
		suite.Equal(req.State, reqBody.State)
		suite.Equal(req.Postcode, reqBody.Postcode)
		suite.Equal(req.Country, reqBody.Country)
		suite.Equal(req.YourPONumber, reqBody.YourPONumber)
		suite.Equal(req.TaxNumber, reqBody.TaxNumber)
		suite.Equal(req.FirstPartyID, reqBody.FirstPartyID)

		fmt.Fprint(w, jblob) // Return the mock response
	})

	// Call the method under test
	resp, err := billingMarketSvc.SetBillingMarket(ctx, req)

	// Verify results
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(supplyID, resp.SupplyID)
}

// TestSetBillingMarketError tests error handling in the SetBillingMarket method.
func (suite *BillingMarketClientTestSuite) TestSetBillingMarketError() {
	ctx := context.Background()
	billingMarketSvc := suite.client.BillingMarketService

	// Create a test request
	req := &SetBillingMarketRequest{
		CurrencyEnum:        "USD",
		Language:            "en",
		BillingContactName:  "Test Contact",
		BillingContactPhone: "+1 555-123-4567",
		BillingContactEmail: "test@example.com",
		Address1:            "123 Main St",
		City:                "Anytown",
		State:               "CA",
		Postcode:            "12345",
		Country:             "US",
		FirstPartyID:        FIRST_PARTY_ID_US,
	}

	// Mock an error response
	errorResponse := `{
        "message": "Invalid request",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "error": "Missing required fields"
    }`

	// Set up HTTP handler
	path := "/v2/market"
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost) // Verify the HTTP method
		w.WriteHeader(http.StatusBadRequest) // Return 400 Bad Request
		fmt.Fprint(w, errorResponse)
	})

	// Call the method under test
	resp, err := billingMarketSvc.SetBillingMarket(ctx, req)

	// Verify results
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "failed to perform HTTP request")
}
