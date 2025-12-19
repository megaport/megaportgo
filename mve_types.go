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

	ResourceTags []ResourceTag `json:"resourceTags,omitempty"`
}

// Nested configuration Fields for the MVE Order
type MVEConfig struct {
	DiversityZone string `json:"diversityZone,omitempty"`
}

// VendorConfig is an interface for MVE vendor configuration.
type VendorConfig interface {
	IsVendorConfig()
}

// VSRConfig represents the configuration for a 6WIND VSR MVE.
type SixwindVSRConfig struct {
	VendorConfig
	Vendor       string `json:"vendor"`
	ImageID      int    `json:"imageId"`
	ProductSize  string `json:"productSize"`
	MVELabel     string `json:"mveLabel,omitempty"`
	SSHPublicKey string `json:"sshPublicKey,omitempty"`
}

// ArubaConfig represents the configuration for an Aruba MVE.
type ArubaConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	AccountKey  string `json:"accountKey,omitempty"`
	SystemTag   string `json:"systemTag,omitempty"`
}

// AviatrixConfig represents the configuration for an Aviatrix Secure Edge MVE.
type AviatrixConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	CloudInit   string `json:"cloudInit,omitempty"`
}

// CiscoConfig represents the configuration for a Cisco MVE.
type CiscoConfig struct {
	VendorConfig
	Vendor             string `json:"vendor"`
	ImageID            int    `json:"imageId"`
	ProductSize        string `json:"productSize"`
	MVELabel           string `json:"mveLabel,omitempty"`
	ManageLocally      bool   `json:"manageLocally,omitempty"`
	AdminSSHPublicKey  string `json:"adminSshPublicKey,omitempty"`
	SSHPublicKey       string `json:"sshPublicKey,omitempty"`
	CloudInit          string `json:"cloudInit,omitempty"`
	FMCIPAddress       string `json:"fmcIpAddress,omitempty"`
	FMCRegistrationKey string `json:"fmcRegistrationKey,omitempty"`
	FMCNatID           string `json:"fmcNatId,omitempty"`
}

// FortinetConfig represents the configuration for a Fortinet MVE.
type FortinetConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	AdminSSHPublicKey string `json:"adminSshPublicKey,omitempty"`
	SSHPublicKey      string `json:"sshPublicKey,omitempty"`
	LicenseData       string `json:"licenseData,omitempty"`
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

// PrismaConfig represents the configuration for a Palo Alto Prisma MVE.
type PrismaConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	IONKey      string `json:"ionKey,omitempty"`
	SecretKey   string `json:"secretKey,omitempty"`
}

// VersaConfig represents the configuration for a Versa MVE.
type VersaConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	DirectorAddress   string `json:"directorAddress,omitempty"`
	ControllerAddress string `json:"controllerAddress,omitempty"`
	LocalAuth         string `json:"localAuth,omitempty"`
	RemoteAuth        string `json:"remoteAuth,omitempty"`
	SerialNumber      string `json:"serialNumber,omitempty"`
}

// VmwareConfig represents the configuration for a VMware MVE.
type VmwareConfig struct {
	VendorConfig
	Vendor            string `json:"vendor"`
	ImageID           int    `json:"imageId"`
	ProductSize       string `json:"productSize"`
	MVELabel          string `json:"mveLabel,omitempty"`
	AdminSSHPublicKey string `json:"adminSshPublicKey,omitempty"`
	SSHPublicKey      string `json:"sshPublicKey,omitempty"`
	VcoAddress        string `json:"vcoAddress,omitempty"`
	VcoActivationCode string `json:"vcoActivationCode,omitempty"`
}

// MerakiConfig represents the configuration for a Meraki MVE.
type MerakiConfig struct {
	VendorConfig
	Vendor      string `json:"vendor"`
	ImageID     int    `json:"imageId"`
	ProductSize string `json:"productSize"`
	MVELabel    string `json:"mveLabel,omitempty"`
	Token       string `json:"token,omitempty"`
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
	AssociatedVXCs        []*VXC                  `json:"associatedVxcs"`
	AssociatedIXs         []*IX                   `json:"associatedIxs"`
}

func (m *MVE) GetType() string {
	return m.Type
}

func (m *MVE) GetUID() string {
	return m.UID
}

func (m *MVE) GetProvisioningStatus() string {
	return m.ProvisioningStatus
}

func (m *MVE) GetAssociatedVXCs() []*VXC {
	return m.AssociatedVXCs
}

func (m *MVE) GetAssociatedIXs() []*IX {
	return m.AssociatedIXs
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
// In the v4 API, Product and Vendor are at the parent level (MVEImageProduct), but we denormalize
// them here for backward compatibility with existing code.
type MVEImage struct {
	ID                int      `json:"id"`
	Version           string   `json:"version"`
	Product           string   `json:"product"`           // Denormalized from parent in v4 API
	Vendor            string   `json:"vendor"`            // Denormalized from parent in v4 API
	VendorDescription string   `json:"vendorDescription"`
	ReleaseImage      bool     `json:"releaseImage"`
	ProductCode       string   `json:"productCode"`
	AvailableSizes    []string `json:"availableSizes"` // New in v4 API - list of compatible MVE sizes
}

// MVEImageProduct represents a vendor/product grouping of MVE images from the v4 API.
// Each product contains multiple image versions.
type MVEImageProduct struct {
	Product         string              `json:"product"`
	Vendor          string              `json:"vendor"`
	VendorProductID string              `json:"vendorProductId"`
	Images          []*MVEImageVersion  `json:"images"`
}

// MVEImageVersion represents an individual MVE image version within a product group (v4 API structure).
type MVEImageVersion struct {
	ID                int      `json:"id"`
	Version           string   `json:"version"`
	ProductCode       string   `json:"productCode"`
	VendorDescription string   `json:"vendorDescription"`
	ReleaseImage      bool     `json:"releaseImage"`
	AvailableSizes    []string `json:"availableSizes"`
}

// MVEImageAPIResponse represents the response to an MVE image request (v3 API - deprecated).
type MVEImageAPIResponse struct {
	Message string                   `json:"message"`
	Terms   string                   `json:"terms"`
	Data    *MVEImageAPIResponseData `json:"data"`
}

// MVEImageAPIResponseData represents the data in an MVE image response (v3 API - deprecated).
type MVEImageAPIResponseData struct {
	Images []*MVEImage `json:"mveImages"`
}

// MVEImageAPIResponseV4 represents the response to an MVE image request from the v4 API.
type MVEImageAPIResponseV4 struct {
	Message string                     `json:"message"`
	Terms   string                     `json:"terms"`
	Data    *MVEImageAPIResponseDataV4 `json:"data"`
}

// MVEImageAPIResponseDataV4 represents the data in an MVE image response from the v4 API.
type MVEImageAPIResponseDataV4 struct {
	Images []*MVEImageProduct `json:"mveImages"`
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
