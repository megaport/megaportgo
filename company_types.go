package megaport

// Market represents a market in the Megaport API.
type Market struct {
	Currency               string `json:"currencyEnum"`
	Language               string `json:"language"`
	CompanyLegalIdentifier string `json:"companyLegalIdentifier"`
	CompanyLegalName       string `json:"companyLegalName"`
	BillingContactName     string `json:"billingContactName"`
	BillingContactPhone    string `json:"billingContactPhone"`
	BillingContactEmail    string `json:"billingContactEmail"`
	AddressLine1           string `json:"address1"`
	AddressLine2           string `json:"address2"`
	City                   string `json:"city"`
	State                  string `json:"state"`
	Postcode               string `json:"postcode"`
	Country                string `json:"country"`
	PONumber               string `json:"yourPoNumber"`
	TaxNumber              string `json:"taxNumber"`
	FirstPartyID           int    `json:"firstPartyId"`
}

// CompanyEnablement represents a company enablement in the Megaport API.
type CompanyEnablement struct {
	TradingName string `json:"tradingName"`
}
