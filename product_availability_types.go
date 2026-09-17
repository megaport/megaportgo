package megaport

// AvailableProductCategory is the category of an available product.
type AvailableProductCategory string

const (
	AvailableProductCategoryProduct AvailableProductCategory = "PRODUCT"
	AvailableProductCategoryAddon   AvailableProductCategory = "ADDON"
	AvailableProductCategoryBundle  AvailableProductCategory = "BUNDLE"
)

// AvailabilitySource is where an effective availability comes from: the market
// default, a per-company override, or no rule configured for the market.
type AvailabilitySource string

const (
	AvailabilitySourceMarket             AvailabilitySource = "MARKET"
	AvailabilitySourceOverride           AvailabilitySource = "OVERRIDE"
	AvailabilitySourceUnconfiguredMarket AvailabilitySource = "UNCONFIGURED_MARKET"
)

// CompanyProductAvailability is a product and the markets a company can order it in.
type CompanyProductAvailability struct {
	Category    AvailableProductCategory `json:"category"`
	DisplayName string                   `json:"displayName"`
	Variant     string                   `json:"variant"`
	Markets     []CompanyProductMarket   `json:"markets"`
}

// CompanyProductMarket is one market a product is available in for a company.
type CompanyProductMarket struct {
	Market      string             `json:"market"`
	MarketName  string             `json:"marketName"`
	MarketOwner string             `json:"marketOwner"`
	CountryCode string             `json:"countryCode"`
	Source      AvailabilitySource `json:"source"`
}

// CompanyProductAvailabilityAPIResponse is the envelope returned by the company
// product availability endpoint.
type CompanyProductAvailabilityAPIResponse struct {
	Message string                        `json:"message"`
	Terms   string                        `json:"terms"`
	Data    []*CompanyProductAvailability `json:"data"`
}
