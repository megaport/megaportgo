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

// AccountType is the type of a Megaport company account.
type AccountType string

const (
	AccountTypeDirectCustomer AccountType = "DIRECT_CUSTOMER"
	AccountTypePartner        AccountType = "PARTNER"
	AccountTypeManagedAccount AccountType = "MANAGED_ACCOUNT"
	AccountTypeAdmin          AccountType = "ADMIN"
	AccountTypeRoot           AccountType = "ROOT"
	AccountTypeDistributor    AccountType = "DISTRIBUTOR"
)

// CurrentCompany is the company the authenticated caller acts in, as returned by GET /v2/me.
// An API-key (M2M) caller receives CompanyUID, CompanyName, AccountType and the parent fields.
// A session user additionally receives CompanyID, RootAccountType, IsManagedContext and Companies.
type CurrentCompany struct {
	CompanyUID        string          `json:"companyUid"`
	CompanyName       string          `json:"companyName"`
	AccountType       AccountType     `json:"accountType,omitempty"`
	ParentCompanyUID  string          `json:"parentCompanyUid,omitempty"`
	ParentCompanyName string          `json:"parentCompanyName,omitempty"`
	CompanyID         int             `json:"companyId,omitempty"`
	RootAccountType   AccountType     `json:"rootAccountType,omitempty"`
	IsManagedContext  bool            `json:"isManagedContext,omitempty"`
	Companies         []MemberCompany `json:"companies,omitempty"`
}

// MemberCompany is a company the authenticated person is a current member of.
type MemberCompany struct {
	CompanyUID  string      `json:"companyUid"`
	Name        string      `json:"name"`
	AccountType AccountType `json:"accountType,omitempty"`
}

// currentCompanyAPIResponse is the envelope returned by GET /v2/me.
type currentCompanyAPIResponse struct {
	Message string          `json:"message"`
	Terms   string          `json:"terms"`
	Data    *CurrentCompany `json:"data"`
}
