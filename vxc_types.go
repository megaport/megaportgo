package megaport

// ---- VXC Detail Types //
type VXC struct {
	ID                 int                 `json:"productId"`
	UID                string              `json:"productUid"`
	ServiceID          int                 `json:"nServiceId"`
	Name               string              `json:"productName"`
	Type               string              `json:"productType"`
	RateLimit          int                 `json:"rateLimit"`
	DistanceBand       string              `json:"distanceBand"`
	ProvisioningStatus string              `json:"provisioningStatus"`
	AEndConfiguration  VXCEndConfiguration `json:"aEnd"`
	BEndConfiguration  VXCEndConfiguration `json:"bEnd"`
	SecondaryName      string              `json:"secondaryName"`
	UsageAlgorithm     string              `json:"usageAlgorithm"`
	CreatedBy          string              `json:"createdBy"`
	LiveDate           int                 `json:"liveDate"`
	CreateDate         int                 `json:"createDate"`
	Resources          VXCResources        `json:"resources"`
	VXCApproval        VXCApproval         `json:"vxcApproval"`
	ContractStartDate  int                 `json:"contractStartDate"`
	ContractEndDate    int                 `json:"contractEndDate"`
	ContractTermMonths int                 `json:"contractTermMonths"`
	CompanyUID         string              `json:"companyUid"`
	CompanyName        string              `json:"companyName"`
	Locked             bool                `json:"locked"`
	AdminLocked        bool                `json:"adminLocked"`
	AttributeTags      map[string]string   `json:"attributeTags"`
	Cancelable         bool                `json:"cancelable"`
}

type VXCEndConfiguration struct {
	OwnerUID              string `json:"ownerUid"`
	UID                   string `json:"productUid"`
	Name                  string `json:"productName"`
	LocationID            int    `json:"locationId"`
	Location              string `json:"location"`
	VLAN                  int    `json:"vlan"`
	InnerVLAN             int    `json:"innerVlan"`
	NetworkInterfaceIndex int    `json:"vNicIndex"`
	SecondaryName         string `json:"secondaryName"`
}

type VXCResources struct {
	Interface     []PortInterface `json:"interface"`
	VirtualRouter interface{}     `json:"virtual_router"`
	CspConnection interface{}     `json:"csp_connection"`
	VLL           VLLConfig       `json:"vll"`
}

type VLLConfig struct {
	AEndVLAN      int    `json:"a_vlan"`
	BEndVLAN      int    `json:"b_vlan"`
	Description   string `json:"description"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	RateLimitMBPS int    `json:"rate_limit_mbps"`
	ResourceName  string `json:"resource_name"`
	ResourceType  string `json:"resource_type"`
}

type VXCApproval struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	UID      string `json:"uid"`
	Type     string `json:"type"`
	NewSpeed int    `json:"newSpeed"`
}

type PartnerLookup struct {
	Bandwidth    int                 `json:"bandwidth"`
	Bandwidths   []int               `json:"bandwidths"`
	Megaports    []PartnerLookupItem `json:"megaports"`
	Peers        []interface{}       `json:"peers"`
	ResourceType string              `json:"resource_type"`
	ServiceKey   string              `json:"service_key"`
	VLAN         int                 `json:"vlan"`
}

type PartnerLookupItem struct {
	ID          int    `json:"port"`
	Type        string `json:"type"`
	VXC         int    `json:"vxc"`
	ProductID   int    `json:"productId"`
	ProductUID  string `json:"productUid"`
	Name        string `json:"name"`
	ServiceID   int    `json:"nServiceId"`
	Description string `json:"description"`
	CompanyID   int    `json:"companyId"`
	CompanyName string `json:"companyName"`
	PortSpeed   int    `json:"portSpeed"`
	LocationID  int    `json:"locationId"`
	State       string `json:"state"`
	Country     string `json:"country"`
}

type VXCUpdate struct {
	Name      string `json:"name"`
	RateLimit int    `json:"rateLimit"`
	AEndVLAN  int    `json:"aEndVlan"`
	BEndVLAN  *int   `json:"bEndVlan,omitempty"`
}

type PartnerVXCUpdate struct {
	Name      string `json:"name"`
	RateLimit int    `json:"rateLimit"`
	AEndVLAN  int    `json:"aEndVlan"`
}

type VXCOrderResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []VXCOrderConfirmation `json:"data"`
}

type VXCResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    VXC    `json:"data"`
}

type VXCOrder struct {
	AssociatedVXCs []VXCOrderConfiguration `json:"associatedVxcs"`
	PortID         string                  `json:"productUid"`
}

type VXCOrderConfiguration struct {
	Name      string                    `json:"productName"`
	RateLimit int                       `json:"rateLimit"`
	AEnd      VXCOrderAEndConfiguration `json:"aEnd"`
	BEnd      VXCOrderBEndConfiguration `json:"bEnd"`
}

type VXCOrderAEndConfiguration struct {
	VLAN          int                       `json:"vlan,omitempty"`
	PartnerConfig VXCOrderAEndPartnerConfig `json:"partnerConfig,omitempty"`
	// Embed this instead of having it top-level so that it's omitted if nil
	// and the marshalled JSON looks right.
	*VXCOrderMVEConfig
}

type VXCOrderMVEConfig struct {
	InnerVLAN             int `json:"innerVlan,omitempty"`
	NetworkInterfaceIndex int `json:"vNicIndex"`
}

type VXCOrderAEndPartnerConfig struct {
	Interfaces []PartnerConfigInterface `json:"interfaces,omitempty"`
}

type VXCOrderBEndConfiguration struct {
	ProductUID string `json:"productUid"`
	VLAN       int    `json:"vlan,omitempty"`
	// Embed this instead of having it top-level so that it's omitted if nil
	// and the marshalled JSON looks right.
	*VXCOrderMVEConfig
}

type VXCOrderConfirmation struct {
	TechnicalServiceUID string `json:"vxcJTechnicalServiceUid"`
}

// BGP CONFIG STUFF

type PartnerConfigInterface struct {
	IpAddresses    []string              `json:"ipAddresses,omitempty"`
	IpRoutes       []IpRoute             `json:"ipRoutes,omitempty"`
	NatIpAddresses []string              `json:"natIpAddresses,omitempty"`
	Bfd            BfdConfig             `json:"bfd,omitempty"`
	BgpConnections []BgpConnectionConfig `json:"bgpConnections,omitempty"`
}

type IpRoute struct {
	Prefix      string `json:"prefix"`
	Description string `json:"description,omitempty"`
	NextHop     string `json:"nextHop"`
}

type BfdConfig struct {
	TxInterval int `json:"txInterval,omitempty"`
	RxInterval int `json:"rxInterval,omitempty"`
	Multiplier int `json:"multiplier,omitempty"`
}

type BgpConnectionConfig struct {
	PeerAsn         int      `json:"peerAsn"`
	LocalIpAddress  string   `json:"localIpAddress"`
	PeerIpAddress   string   `json:"peerIpAddress"`
	Password        string   `json:"password,omitempty"`
	Shutdown        bool     `json:"shutdown"`
	Description     string   `json:"description,omitempty"`
	MedIn           int      `json:"medIn,omitempty"`
	MedOut          int      `json:"medOut,omitempty"`
	BfdEnabled      bool     `json:"bfdEnabled"`
	ExportPolicy    string   `json:"exportPolicy,omitempty"`
	PermitExportTo  []string `json:"permitExportTo,omitempty"`
	DenyExportTo    []string `json:"denyExportTo,omitempty"`
	ImportWhitelist int      `json:"importWhitelist,omitempty"`
	ImportBlacklist int      `json:"importBlacklist,omitempty"`
	ExportWhitelist int      `json:"exportWhitelist,omitempty"`
	ExportBlacklist int      `json:"exportBlacklist,omitempty"`
}

// AWS STUFF

type AWSVXCOrder struct {
	AssociatedVXCs []AWSVXCOrderConfiguration `json:"associatedVxcs"`
	PortID         string                     `json:"productUid"`
}

type AWSVXCOrderConfiguration struct {
	Name      string                       `json:"productName"`
	RateLimit int                          `json:"rateLimit"`
	AEnd      VXCOrderAEndConfiguration    `json:"aEnd"`
	BEnd      AWSVXCOrderBEndConfiguration `json:"bEnd"`
}

type AWSVXCOrderBEndConfiguration struct {
	ProductUID    string                       `json:"productUid"`
	PartnerConfig AWSVXCOrderBEndPartnerConfig `json:"partnerConfig"`
}

type AWSVXCOrderBEndPartnerConfig struct {
	ConnectType       string `json:"connectType"`
	Type              string `json:"type"`
	OwnerAccount      string `json:"ownerAccount"`
	ASN               int    `json:"asn,omitempty"`
	AmazonASN         int    `json:"amazonAsn,omitempty"`
	AuthKey           string `json:"authKey,omitempty"`
	Prefixes          string `json:"prefixes,omitempty"`
	CustomerIPAddress string `json:"customerIpAddress,omitempty"`
	AmazonIPAddress   string `json:"amazonIpAddress,omitempty"`
	ConnectionName    string `json:"name,omitempty"`
}

// Partner

type PartnerOrder struct {
	PortID         string                 `json:"productUid"`
	AssociatedVXCs []PartnerOrderContents `json:"associatedVxcs"`
}

type PartnerOrderContents struct {
	Name      string                        `json:"productName"`
	RateLimit int                           `json:"rateLimit"`
	AEnd      VXCOrderAEndConfiguration     `json:"aEnd"`
	BEnd      PartnerOrderBEndConfiguration `json:"bEnd"`
}

type PartnerOrderBEndConfiguration struct {
	PartnerPortID string      `json:"productUid"`
	PartnerConfig interface{} `json:"partnerConfig"`
}

type PartnerOrderAzurePartnerConfig struct {
	ConnectType string                           `json:"connectType"`
	ServiceKey  string                           `json:"serviceKey"`
	Peers       []PartnerOrderAzurePeeringConfig `json:"peers"`
}

type PartnerOrderAzurePeeringConfig struct {
	Type            string `json:"type"`
	PeerASN         string `json:"peer_asn"`
	PrimarySubnet   string `json:"primary_subnet"`
	SecondarySubnet string `json:"secondary_subnet"`
	Prefixes        string `json:"prefixes,omitempty"`
	SharedKey       string `json:"shared_key,omitempty"`
	VLAN            int    `json:"vlan"`
}

type PartnerOrderGooglePartnerConfig struct {
	ConnectType string `json:"connectType"`
	PairingKey  string `json:"pairingKey"`
}

type PartnerOrderOciPartnerConfig struct {
	ConnectType     string `json:"connectType"`
	VirtualCircutId string `json:"virtualCircuitId"`
}

type PartnerLookupResponse struct {
	Message string        `json:"message"`
	Data    PartnerLookup `json:"data"`
	Terms   string        `json:"terms"`
}

type PartnerMegaportResponse struct {
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    []PartnerMegaport `json:"data"`
}

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
