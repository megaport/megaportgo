package megaport

// PartnerLookupResponse represents a response from the Megaport API after looking up a Partner Megaport.
type PartnerLookupResponse struct {
	Message string        `json:"message"`
	Data    PartnerLookup `json:"data"`
	Terms   string        `json:"terms"`
}

// PartnerMegaportResponse represents a response from the Megaport API after querying a Partner Megaport.
type PartnerMegaportResponse struct {
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    []*PartnerMegaport `json:"data"`
}

// PartnerMegaport represents a Partner Megaport in the Megaport API.
type PartnerMegaport struct {
	ConnectType   string `json:"connectType"`
	ProductUID    string `json:"productUid"`
	ProductName   string `json:"title"`
	CompanyUID    string `json:"companyUid"`
	CompanyName   string `json:"companyName"`
	DiversityZone string `json:"diversityZone"`
	LocationId    int    `json:"locationId"`
	Speed         int    `json:"speed"`
	Rank          int    `json:"rank"`
	VXCPermitted  bool   `json:"vxcPermitted"`
}
