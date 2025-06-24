package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// BillingMarketService is an interface for interfacing with the Billing Market endpoints of the Megaport API
type BillingMarketService interface {
	// SetBillingMarket configures the billing market (and currency) and the billing contact details.
	SetBillingMarket(ctx context.Context, req *SetBillingMarketRequest) (*SetBillingMarketResponse, error)
	// GetBillingMarkets retrieves the billing markets and contact details for the account.
	GetBillingMarkets(ctx context.Context) ([]*BillingMarket, error)
}

// BillingMarketServiceOp handles communication with the Billing Market related methods of the Megaport API.
type BillingMarketServiceOp struct {
	client *Client
}

// NewBillingMarketService returns a BillingMarketService
func NewBillingMarketService(c *Client) BillingMarketService {
	return &BillingMarketServiceOp{client: c}
}

// BillingMarket represents a billing market as returned by the Megaport API.
type BillingMarket struct {
	ID                            int      `json:"id"`
	WellKnownSupplier             string   `json:"wellKnownSupplier"`
	SupplierName                  string   `json:"supplierName"`
	CurrencyEnum                  string   `json:"currencyEnum"`
	Language                      string   `json:"language"`
	BillingContactName            string   `json:"billingContactName"`
	BillingContactEmail           string   `json:"billingContactEmail"`
	BillingContactPhone           string   `json:"billingContactPhone"`
	Address1                      string   `json:"address1"`
	Postcode                      string   `json:"postcode"`
	Country                       string   `json:"country"`
	City                          string   `json:"city"`
	State                         string   `json:"state"`
	InvoiceTemplate               string   `json:"invoiceTemplate"`
	TaxRate                       float64  `json:"taxRate"`
	EstimateInvoice               float64  `json:"estimateInvoice"`
	FirstPartyID                  int      `json:"firstPartyId"`
	SecondPartyID                 int      `json:"secondPartyId"`
	AttachInvoiceToEmail          bool     `json:"attachInvoiceToEmail"`
	Region                        string   `json:"region"`
	PaymentTermInDays             int      `json:"paymentTermInDays"`
	StripeAccountPublishableKey   string   `json:"stripeAccountPublishableKey"`
	StripeSupportedBankCurrencies []string `json:"stripeSupportedBankCurrencies"`
	VATExempt                     bool     `json:"vatExempt"`
	Active                        bool     `json:"active"`
	SecuredHash                   string   `json:"securedHash"`
}

// SetBillingMarketRequest represents the request body for setting a billing market in the Megaport API.
type SetBillingMarketRequest struct {
	CurrencyEnum        string  `json:"currencyEnum"`           // Billing currency (e.g., USD, AUD, etc.)
	Language            string  `json:"language"`               // Two-letter language code (e.g., "en")
	BillingContactName  string  `json:"billingContactName"`     // Name of the billing contact
	BillingContactPhone string  `json:"billingContactPhone"`    // Phone number of the billing contact
	BillingContactEmail string  `json:"billingContactEmail"`    // Email address of the billing contact
	Address1            string  `json:"address1"`               // Physical address line 1
	Address2            *string `json:"address2"`               // Physical address line 2 (optional)
	City                string  `json:"city"`                   // City for the billing contact
	State               string  `json:"state"`                  // State or region
	Postcode            string  `json:"postcode"`               // Postal code
	Country             string  `json:"country"`                // Country code (e.g., "AU")
	YourPONumber        string  `json:"yourPoNumber,omitempty"` // Optional PO number for tracking
	TaxNumber           string  `json:"taxNumber,omitempty"`    // Optional tax or VAT registration number
	FirstPartyID        int     `json:"firstPartyId"`           // ID for the billing market (see FIRST_PARTY_ID constants)
}

// FIRST_PARTY_ID constants for supported billing markets.
const (
	FIRST_PARTY_ID_US = 1558
	FIRST_PARTY_ID_AU = 808
	FIRST_PARTY_ID_AT = 20442
	FIRST_PARTY_ID_BE = 20449
	FIRST_PARTY_ID_BG = 4640
	FIRST_PARTY_ID_CA = 1652
	FIRST_PARTY_ID_CH = 8299
	FIRST_PARTY_ID_DE = 4515
	FIRST_PARTY_ID_DK = 20447
	FIRST_PARTY_ID_ES = 30369
	FIRST_PARTY_ID_FI = 20440
	FIRST_PARTY_ID_FR = 20451
	FIRST_PARTY_ID_HK = 819
	FIRST_PARTY_ID_IE = 2683
	FIRST_PARTY_ID_IT = 30367
	FIRST_PARTY_ID_JP = 20453
	FIRST_PARTY_ID_LU = 30423
	FIRST_PARTY_ID_NL = 2685
	FIRST_PARTY_ID_NO = 20438
	FIRST_PARTY_ID_NZ = 855
	FIRST_PARTY_ID_PL = 20444
	FIRST_PARTY_ID_SE = 2681
	FIRST_PARTY_ID_SG = 817
	FIRST_PARTY_ID_UK = 2675
)

type SetBillingMarketResponse struct {
	SupplyID int `json:"supplyId"`
}

// SetBillingMarket creates or updates a billing market using the Megaport API. It sends a POST request to /v2/market with the provided SetBillingMarketRequest.
func (svc *BillingMarketServiceOp) SetBillingMarket(ctx context.Context, req *SetBillingMarketRequest) (*SetBillingMarketResponse, error) {
	url := svc.client.BaseURL.JoinPath("/v2/market").String()

	// Create the HTTP request
	httpReq, err := svc.client.NewRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Perform the request
	resp, err := svc.client.Do(ctx, httpReq, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var apiResp struct {
		Message string `json:"message"`
		Terms   string `json:"terms"`
		Data    struct {
			SupplyID int `json:"supplyId"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode SetBillingMarket response: %w", err)
	}

	return &SetBillingMarketResponse{
		SupplyID: apiResp.Data.SupplyID,
	}, nil
}

// GetBillingMarkets retrieves the billing markets and contact details for the account.
func (svc *BillingMarketServiceOp) GetBillingMarkets(ctx context.Context) ([]*BillingMarket, error) {
	url := svc.client.BaseURL.JoinPath("/v2/market").String()

	req, err := svc.client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Message string           `json:"message"`
		Terms   string           `json:"terms"`
		Data    []*BillingMarket `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode GetBillingMarkets response: %w", err)
	}

	return apiResp.Data, nil
}
