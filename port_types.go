package megaport

// PortOrder represents a Megaport Port Order from the Megaport Products API.
type PortOrder struct {
	Name                  string              `json:"productName"`
	Term                  int                 `json:"term"`
	ProductType           string              `json:"productType"`
	PortSpeed             int                 `json:"portSpeed"`
	LocationID            int                 `json:"locationId"`
	CreateDate            int64               `json:"createDate"`
	Virtual               bool                `json:"virtual"`
	Market                string              `json:"market"`
	CostCentre            string              `json:"costCentre,omitempty"`
	LagPortCount          int                 `json:"lagPortCount,omitempty"`
	MarketplaceVisibility bool                `json:"marketplaceVisibility"`
	Config                PortOrderConfig     `json:"config"`
	PromoCode             string              `json:"promoCode,omitempty"`
	ResourceTags          []map[string]string `json:"resourceTags,omitempty"`
}

type PortOrderConfig struct {
	DiversityZone string `json:"diversityZone,omitempty"`
}

// PortOrderConfirmation represents a response from the Megaport Products API after ordering a port.
type PortOrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

// Port represents a Megaport Port in the Megaport Port API.
type Port struct {
	ID                    int                     `json:"productId"`
	UID                   string                  `json:"productUid"`
	Name                  string                  `json:"productName"`
	Type                  string                  `json:"productType"`
	ProvisioningStatus    string                  `json:"provisioningStatus"`
	CreateDate            *Time                   `json:"createDate"`
	CreatedBy             string                  `json:"createdBy"`
	PortSpeed             int                     `json:"portSpeed"`
	TerminateDate         *Time                   `json:"terminateDate"`
	LiveDate              *Time                   `json:"liveDate"`
	Market                string                  `json:"market"`
	LocationID            int                     `json:"locationId"`
	UsageAlgorithm        string                  `json:"usageAlgorithm"`
	MarketplaceVisibility bool                    `json:"marketplaceVisibility"`
	VXCPermitted          bool                    `json:"vxcpermitted"`
	VXCAutoApproval       bool                    `json:"vxcAutoApproval"`
	SecondaryName         string                  `json:"secondaryName"`
	LAGPrimary            bool                    `json:"lagPrimary"`
	LAGID                 int                     `json:"lagId"`
	AggregationID         int                     `json:"aggregationId"`
	CompanyUID            string                  `json:"companyUid"`
	CompanyName           string                  `json:"companyName"`
	CostCentre            string                  `json:"costCentre"`
	ContractStartDate     *Time                   `json:"contractStartDate"`
	ContractEndDate       *Time                   `json:"contractEndDate"`
	ContractTermMonths    int                     `json:"contractTermMonths"`
	AttributeTags         PortAttributeTags       `json:"attributeTags"`
	Virtual               bool                    `json:"virtual"`
	BuyoutPort            bool                    `json:"buyoutPort"`
	Locked                bool                    `json:"locked"`
	AdminLocked           bool                    `json:"adminLocked"`
	Cancelable            bool                    `json:"cancelable"`
	DiversityZone         string                  `json:"diversityZone"`
	VXCResources          PortResources           `json:"resources"`
	LocationDetails       *ProductLocationDetails `json:"locationDetail"`
	ResourceTags          map[string]string       `json:"resourceTags"`
}

// PortResources represents the resources associated with a Megaport Port.
type PortResources struct {
	Interface PortInterface `json:"interface"`
}

// PortResourcesInterface represents the resources interface associated with a Megaport Port.
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

// PortInterface represents the interface associated with a Megaport Port.
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

// PortOrderResponse represents a response from the Megaport Products API after ordering a port.
type PortOrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []PortOrderConfirmation `json:"data"`
}

// PortResponse represents a response from the Megaport Port API after querying a port.
type PortResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    Port   `json:"data"`
}

// PortAttributes represents attributes associated with a Megaport Port.
type PortAttributeTags struct {
	TerminatedServiceDetails PortTerminatedServiceDetails `json:"terminatedServiceDetails"`
}

// PortTerminatedServiceDetails represents terminated service details associated with a Megaport Port.
type PortTerminatedServiceDetails struct {
	Location  PortTerminatedServiceDetailsLocation  `json:"location"`
	Interface PortTerminatedServiceDetailsInterface `json:"interface"`
	Device    string                                `json:"device"`
}

// PortTerminatedServiceDetailsLocation represents the location of a terminated service associated with a Megaport Port.
type PortTerminatedServiceDetailsLocation struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	SiteCode string `json:"site_code"`
}

// PortTerminatedServiceDetailsInterface represents the interface of a terminated service associated with a Megaport Port.
type PortTerminatedServiceDetailsInterface struct {
	ResourceType string `json:"resource_type"`
	Demarcation  string `json:"demarcation"`
	LOATemplate  string `json:"loa_template"`
	Media        string `json:"media"`
	PortSpeed    int    `json:"port_speed"`
	ResourceName string `json:"resource_name"`
	Up           int    `json:"up"`
	Shutdown     bool   `json:"shutdown"`
}

type PortVLANAvailabilityAPIResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    []int  `json:"data"`
}
