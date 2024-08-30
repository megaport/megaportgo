package megaport

// MVEOrderConfig represents a request to buy an MVE from the Megaport Products API.
type MVEOrderConfig struct {
	LocationID        int                   `json:"locationId"`
	Name              string                `json:"productName"`
	Term              int                   `json:"term"`
	ProductType       string                `json:"productType"`
	PromoCode         string                `json:"promoCode,omitempty"`
	CostCentre        string                `json:"costCentre,omitempty"`
	NetworkInterfaces []MVENetworkInterface `json:"vnics"`
	VendorConfig      VendorConfig          `json:"vendorConfig"`
	Config            MVEConfig             `json:"config"`
}

// Nested configuration Fields for the MVE Order
type MVEConfig struct {
	DiversityZone string `json:"diversityZone,omitempty"`
}

// VendorConfig is an interface for MVE vendor configuration.
type VendorConfig interface {
	IsVendorConfig()
}

// ArubaConfig represents the configuration for an Aruba MVE.
type ArubaConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	AccountName string `json:"accountName"`
	AccountKey  string `json:"accountKey"`
	SystemTag   string `json:"systemTag"`
}

// CiscoConfig represents the configuration for a Cisco MVE.
type CiscoConfig struct {
	VendorConfig
	Vendor             string `json:"vendor"`
	ImageID            int    `json:"imageId"`
	ProductSize        string `json:"productSize"`
	MVELabel           string `json:"mveLabel,omitempty"`
	ManageLocally      bool   `json:"manageLocally"`
	AdminSSHPublicKey  string `json:"adminSshPublicKey"`
	SSHPublicKey       string `json:"sshPublicKey"`
	CloudInit          string `json:"cloudInit"`
	FMCIPAddress       string `json:"fmcIpAddress"`
	FMCRegistrationKey string `json:"fmcRegistrationKey"`
	FMCNatID           string `json:"fmcNatId"`
}

// FortinetConfig represents the configuration for a Fortinet MVE.
type FortinetConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	AdminSSHPublicKey string `json:"adminSshPublicKey"`
	SSHPublicKey      string `json:"sshPublicKey"`
	LicenseData       string `json:"licenseData"`
}

// PaloAltoConfig represents the configuration for a Palo Alto MVE.
type PaloAltoConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize,omitempty"`
	MVELabel          string `json:"mveLabel,omitempty"`
	AdminSSHPublicKey string `json:"adminSshPublicKey,omitempty"`
	SSHPublicKey      string `json:"sshPublicKey,omitempty"`
	AdminPasswordHash string `json:"adminPasswordHash,omitempty"`
	LicenseData       string `json:"licenseData,omitempty"`
}

// VersaConfig represents the configuration for a Versa MVE.
type VersaConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	DirectorAddress   string `json:"directorAddress"`
	ControllerAddress string `json:"controllerAddress"`
	LocalAuth         string `json:"localAuth"`
	RemoteAuth        string `json:"remoteAuth"`
	SerialNumber      string `json:"serialNumber"`
}

// VmwareConfig represents the configuration for a VMware MVE.
type VmwareConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	AdminSSHPublicKey string `json:"adminSshPublicKey"`
	SSHPublicKey      string `json:"sshPublicKey"`
	VcoAddress        string `json:"vcoAddress"`
	VcoActivationCode string `json:"vcoActivationCode"`
}

// MerakiConfig represents the configuration for a Meraki MVE.
type MerakiConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	Token       string `json:"token"`
}

// MVENetworkInterface represents a vNIC.
type MVENetworkInterface struct {
	Description string `json:"description"`
	VLAN        int    `json:"vlan"`
}

// InstanceSize encodes the available MVE instance sizes.
type MVEInstanceSize string

// MVE instance sizes.
const (
	MVE_SMALL  MVEInstanceSize = "SMALL"
	MVE_MEDIUM MVEInstanceSize = "MEDIUM"
	MVE_LARGE  MVEInstanceSize = "LARGE"
	MVE_XLARGE MVEInstanceSize = "X_LARGE_12"
)

// MVEOrderConfirmation represents the response to an MVE order request.
type MVEOrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

// MVE represents a Megaport Virtual Edge from the Megaport MVE API.
type MVE struct {
	ID                    int                     `json:"productId"`
	UID                   string                  `json:"productUid"`
	Name                  string                  `json:"productName"`
	Type                  string                  `json:"productType"`
	ProvisioningStatus    string                  `json:"provisioningStatus"`
	CreateDate            *Time                   `json:"createDate"`
	CreatedBy             string                  `json:"createdBy"`
	TerminateDate         *Time                   `json:"terminateDate"`
	LiveDate              *Time                   `json:"liveDate"`
	Market                string                  `json:"market"`
	LocationID            int                     `json:"locationId"`
	UsageAlgorithm        string                  `json:"usageAlgorithm"`
	MarketplaceVisibility bool                    `json:"marketplaceVisibility"`
	VXCPermitted          bool                    `json:"vxcpermitted"`
	VXCAutoApproval       bool                    `json:"vxcAutoApproval"`
	SecondaryName         string                  `json:"secondaryName"`
	CompanyUID            string                  `json:"companyUid"`
	CompanyName           string                  `json:"companyName"`
	ContractStartDate     *Time                   `json:"contractStartDate"`
	ContractEndDate       *Time                   `json:"contractEndDate"`
	ContractTermMonths    int                     `json:"contractTermMonths"`
	AttributeTags         map[string]string       `json:"attributeTags"`
	CostCentre            string                  `json:"costCentre"`
	Virtual               bool                    `json:"virtual"`
	BuyoutPort            bool                    `json:"buyoutPort"`
	Locked                bool                    `json:"locked"`
	AdminLocked           bool                    `json:"adminLocked"`
	Cancelable            bool                    `json:"cancelable"`
	Resources             *MVEResources           `json:"resources"`
	Vendor                string                  `json:"vendor"`
	Size                  string                  `json:"mveSize"`
	DiversityZone         string                  `json:"diversityZone"`
	NetworkInterfaces     []*MVENetworkInterface  `json:"vnics"`
	LocationDetails       *ProductLocationDetails `json:"locationDetail"`
}

// MVEResources represents the resources associated with an MVE.
type MVEResources struct {
	Interface       *PortInterface       `json:"interface"`
	VirtualMachines []*MVEVirtualMachine `json:"virtual_machine"`
}

// MVEVirtualMachine represents a virtual machine associated with an MVE.
type MVEVirtualMachine struct {
	ID           int                     `json:"id"`
	CpuCount     int                     `json:"cpu_count"`
	Image        *MVEVirtualMachineImage `json:"image"`
	ResourceType string                  `json:"resource_type"`
	Up           bool                    `json:"up"`
	Vnics        []*MVENetworkInterface  `json:"vnics"`
}

// MVVEVirtualMachineImage represents the image associated with an MVE virtual machine.
type MVEVirtualMachineImage struct {
	ID      int    `json:"id"`
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Version string `json:"version"`
}

// MVEOrderResponse represents the response to an MVE order request.
type MVEOrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*MVEOrderConfirmation `json:"data"`
}

// MVEResponse represents the response to an MVE request.
type MVEResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    *MVE   `json:"data"`
}

// MVEImage represents details for an MVE image, including image ID, version, product, and vendor.
type MVEImage struct {
	ID                int    `json:"id"`
	Version           string `json:"version"`
	Product           string `json:"product"`
	Vendor            string `json:"vendor"`
	VendorDescription string `json:"vendorDescription"`
	ReleaseImage      bool   `json:"releaseImage"`
	ProductCode       string `json:"productCode"`
}

// MVEImageAPIResponse represents the response to an MVE image request.
type MVEImageAPIResponse struct {
	Message string                   `json:"message"`
	Terms   string                   `json:"terms"`
	Data    *MVEImageAPIResponseData `json:"data"`
}

// MVEImageAPIResponseData represents the data in an MVE image response.
type MVEImageAPIResponseData struct {
	Images []*MVEImage `json:"mveImages"`
}

// MVESize represents the details on the MVE size. The instance size determines the MVE capabilities, such as how many concurrent connections it can support. The compute sizes are 2/8, 4/16, 8/32, and 12/48, where the first number is the CPU and the second number is the GB of available RAM. Each size has 4 GB of RAM for every vCPU allocated.
type MVESize struct {
	Size         string `json:"size"`
	Label        string `json:"label"`
	CPUCoreCount int    `json:"cpuCoreCount"`
	RamGB        int    `json:"ramGB"`
}

// MVESizeAPIResponse represents the response to an MVE size request, returning a list of currently available MVE sizes and details for each size.
type MVESizeAPIResponse struct {
	Message string     `json:"message"`
	Terms   string     `json:"terms"`
	Data    []*MVESize `json:"data"`
}
