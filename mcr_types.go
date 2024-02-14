package megaport

type MCROrder struct {
	LocationID    int            `json:"locationId"`
	Name          string         `json:"productName"`
	DiversityZone string         `json:"diversityZone"`
	Term          int            `json:"term"`
	Type          string         `json:"productType"`
	PortSpeed     int            `json:"portSpeed"`
	Config        MCROrderConfig `json:"config"`
}

type MCROrderConfig struct {
	ASN int `json:"mcrAsn,omitempty"`
}

type MCROrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

type MCR struct {
	ID                    int               `json:"productId"`
	UID                   string            `json:"productUid"`
	Name                  string            `json:"productName"`
	Type                  string            `json:"productType"`
	ProvisioningStatus    string            `json:"provisioningStatus"`
	CreateDate            int               `json:"createDate"`
	CreatedBy             string            `json:"createdBy"`
	CostCentre 	          string            `json:"costCentre"`
	PortSpeed             int               `json:"portSpeed"`
	TerminateDate         int               `json:"terminateDate"`
	LiveDate              int               `json:"liveDate"`
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
	ContractStartDate     int               `json:"contractStartDate"`
	ContractEndDate       int               `json:"contractEndDate"`
	ContractTermMonths    int               `json:"contractTermMonths"`
	AttributeTags         map[string]string `json:"attributeTags"`
	Virtual               bool              `json:"virtual"`
	BuyoutPort            bool              `json:"buyoutPort"`
	Locked                bool              `json:"locked"`
	AdminLocked           bool              `json:"adminLocked"`
	Cancelable            bool              `json:"cancelable"`
	Resources             MCRResources      `json:"resources"`
}

type MCRResources struct {
	Interface     PortInterface    `json:"interface"`
	VirtualRouter MCRVirtualRouter `json:"virtual_router"`
}

type MCRVirtualRouter struct {
	ID           int    `json:"id"`
	ASN          int    `json:"mcrAsn"`
	Name         string `json:"name"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Speed        int    `json:"speed"`
}

type MCRPrefixFilterList struct {
	Description   string                `json:"description"`
	AddressFamily string                `json:"addressFamily"`
	Entries       []*MCRPrefixListEntry `json:"entries"`
}

type MCRPrefixListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"`
	Le     int    `json:"le,omitempty"`
}

type MCROrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*MCROrderConfirmation `json:"data"`
}

type MCRResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    *MCR   `json:"data"`
}

type PrefixFilterList struct {
	Id            int    `json:"id"`
	Description   string `json:"description"`
	AddressFamily string `json:"addressFamily"`
}

type MCRPrefixFilterListResponse struct {
	Message string              `json:"message"`
	Terms   string              `json:"terms"`
	Data    []*PrefixFilterList `json:"data"`
}
