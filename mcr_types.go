package megaport

// MCROrder represents a request to buy an MCR from the Megaport Products API.
type MCROrder struct {
	LocationID    int            `json:"locationId"`
	Name          string         `json:"productName"`
	DiversityZone string         `json:"diversityZone"`
	Term          int            `json:"term"`
	Type          string         `json:"productType"`
	PortSpeed     int            `json:"portSpeed"`
	CostCentre    string         `json:"costCentre"`
	Config        MCROrderConfig `json:"config"`
}

// MCROrderConfig represents the configuration for an MCR order.
type MCROrderConfig struct {
	ASN int `json:"mcrAsn,omitempty"`
}

// MCROrderConfirmation represents a response from the Megaport Products API after ordering an MCR.
type MCROrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

// MCR represents a Megaport Cloud Router in the Megaport MCR API.
type MCR struct {
	ID                    int               `json:"productId"`
	UID                   string            `json:"productUid"`
	Name                  string            `json:"productName"`
	Type                  string            `json:"productType"`
	ProvisioningStatus    string            `json:"provisioningStatus"`
	CreateDate            *Time             `json:"createDate"`
	CreatedBy             string            `json:"createdBy"`
	CostCentre            string            `json:"costCentre"`
	PortSpeed             int               `json:"portSpeed"`
	TerminateDate         *Time             `json:"terminateDate"`
	LiveDate              *Time             `json:"liveDate"`
	Market                string            `json:"market"`
	LocationID            int               `json:"locationId"`
	UsageAlgorithm        string            `json:"usageAlgorithm"`
	MarketplaceVisibility bool              `json:"marketplaceVisibility"`
	VXCPermitted          bool              `json:"vxcpermitted"`
	VXCAutoApproval       bool              `json:"vxcAutoApproval"`
	SecondaryName         string            `json:"secondaryName"`
	LAGPrimary            bool              `json:"lagPrimary"`
	LAGID                 int               `json:"lagId"`
	AggregationID         int               `json:"aggregationId"`
	CompanyUID            string            `json:"companyUid"`
	CompanyName           string            `json:"companyName"`
	ContractStartDate     *Time             `json:"contractStartDate"`
	ContractEndDate       *Time             `json:"contractEndDate"`
	ContractTermMonths    int               `json:"contractTermMonths"`
	AttributeTags         map[string]string `json:"attributeTags"`
	Virtual               bool              `json:"virtual"`
	BuyoutPort            bool              `json:"buyoutPort"`
	Locked                bool              `json:"locked"`
	AdminLocked           bool              `json:"adminLocked"`
	Cancelable            bool              `json:"cancelable"`
	Resources             MCRResources      `json:"resources"`
}

// MCRResources represents the resources associated with an MCR.
type MCRResources struct {
	Interface     PortInterface    `json:"interface"`
	VirtualRouter MCRVirtualRouter `json:"virtual_router"`
}

// MCRVirtualRouter represents the virtual router associated with an MCR.
type MCRVirtualRouter struct {
	ID           int    `json:"id"`
	ASN          int    `json:"mcrAsn"`
	Name         string `json:"name"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Speed        int    `json:"speed"`
}

// MCRPrefixFilterList represents a prefix filter list associated with an MCR.
type MCRPrefixFilterList struct {
	Description   string                `json:"description"`
	AddressFamily string                `json:"addressFamily"`
	Entries       []*MCRPrefixListEntry `json:"entries"`
}

// MCRPrefixListEntry represents an entry in a prefix filter list.
type MCRPrefixListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"`
	Le     int    `json:"le,omitempty"`
}

// MCROrdersResponse represents a response from the Megaport Products API after ordering an MCR.
type MCROrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*MCROrderConfirmation `json:"data"`
}

// MCRResponse represents a response from the Megaport MCR API after querying an MCR.
type MCRResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    *MCR   `json:"data"`
}

// PrefixFilterList represents a prefix filter list associated with an MCR.
type PrefixFilterList struct {
	Id            int    `json:"id"`
	Description   string `json:"description"`
	AddressFamily string `json:"addressFamily"`
}

// MCRPrefixFilterListResponse represents a response from the Megaport MCR API after querying an MCR's prefix filter list.
type MCRPrefixFilterListResponse struct {
	Message string              `json:"message"`
	Terms   string              `json:"terms"`
	Data    []*PrefixFilterList `json:"data"`
}
