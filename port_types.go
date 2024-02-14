package megaport

type PortOrder struct {
	Name                  string `json:"productName"`
	Term                  int    `json:"term"`
	ProductType           string `json:"productType"`
	PortSpeed             int    `json:"portSpeed"`
	LocationID            int    `json:"locationId"`
	CreateDate            int64  `json:"createDate"`
	Virtual               bool   `json:"virtual"`
	Market                string `json:"market"`
	LagPortCount          int    `json:"lagPortCount,omitempty"`
	MarketplaceVisibility bool   `json:"marketplaceVisibility"`
	DiversityZone         string `json:"diversityZone"`
}

type PortOrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

type Port struct {
	ID                    int                    `json:"productId"`
	UID                   string                 `json:"productUid"`
	Name                  string                 `json:"productName"`
	Type                  string                 `json:"productType"`
	ProvisioningStatus    string                 `json:"provisioningStatus"`
	CreateDate            int                    `json:"createDate"`
	CreatedBy             string                 `json:"createdBy"`
	PortSpeed             int                    `json:"portSpeed"`
	TerminateDate         int                    `json:"terminateDate"`
	LiveDate              int                    `json:"liveDate"`
	Market                string                 `json:"market"`
	LocationID            int                    `json:"locationId"`
	UsageAlgorithm        string                 `json:"usageAlgorithm"`
	MarketplaceVisibility bool                   `json:"marketplaceVisibility"`
	VXCPermitted          bool                   `json:"vxcpermitted"`
	VXCAutoApproval       bool                   `json:"vxcAutoApproval"`
	SecondaryName         string                 `json:"secondaryName"`
	LAGPrimary            bool                   `json:"lagPrimary"`
	LAGID                 int                    `json:"lagId"`
	AggregationID         int                    `json:"aggregationId"`
	CompanyUID            string                 `json:"companyUid"`
	CompanyName           string                 `json:"companyName"`
	CostCentre 			  string	             `json:"costCentre"`
	ContractStartDate     int                    `json:"contractStartDate"`
	ContractEndDate       int                    `json:"contractEndDate"`
	ContractTermMonths    int                    `json:"contractTermMonths"`
	AttributeTags         map[string]interface{} `json:"attributeTags"`
	Virtual               bool                   `json:"virtual"`
	BuyoutPort            bool                   `json:"buyoutPort"`
	Locked                bool                   `json:"locked"`
	AdminLocked           bool                   `json:"adminLocked"`
	Cancelable            bool                   `json:"cancelable"`
	VXCResources          PortResources          `json:"resources"`
}

type PortResources struct {
	Interface PortInterface `json:"interface"`
}

type PortResourcesInterface struct {
	Demarcation  string `json:"demarcation"`
	Description  string `json:"description"`
	ID           int    `json:"id"`
	LOATemplate  string `json:"loa_template"`
	Media        string `json:"media"`
	Name         string `json:"name"`
	PortSpeed    int    `json:"port_speed"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Up           int    `json:"up"`
}

type PortInterface struct {
	Demarcation  string `json:"demarcation"`
	Description  string `json:"description"`
	ID           int    `json:"id"`
	LOATemplate  string `json:"loa_template"`
	Media        string `json:"media"`
	Name         string `json:"name"`
	PortSpeed    int    `json:"port_speed"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Up           int    `json:"up"`
}

type PortOrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []PortOrderConfirmation `json:"data"`
}

type PortResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    Port   `json:"data"`
}
