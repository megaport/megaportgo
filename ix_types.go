package megaport

// IX represents an Internet Exchange in the Megaport API
type IX struct {
	ProductID          int               `json:"productId"`
	ProductUID         string            `json:"productUid"`
	LocationID         int               `json:"locationId"`
	LocationDetail     IXLocationDetail  `json:"locationDetail"`
	Term               int               `json:"term"`
	LocationUID        string            `json:"locationUid"`
	ProductName        string            `json:"productName"`
	ProvisioningStatus string            `json:"provisioningStatus"`
	RateLimit          int               `json:"rateLimit"`
	PromoCode          string            `json:"promoCode"`
	CreateDate         *Time             `json:"createDate"`
	DeployDate         *Time             `json:"deployDate"`
	SecondaryName      string            `json:"secondaryName"`
	AttributeTags      map[string]string `json:"attributeTags"`
	VLAN               int               `json:"vlan"`
	MACAddress         string            `json:"macAddress"`
	IXPeerMacro        string            `json:"ixPeerMacro"`
	ASN                int               `json:"asn"`
	NetworkServiceType string            `json:"networkServiceType"`
	PublicGraph        bool              `json:"publicGraph"`
	UsageAlgorithm     string            `json:"usageAlgorithm"`
	Resources          IXResources       `json:"resources"`
}

// IXLocationDetail represents the location information for an IX
type IXLocationDetail struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Metro   string `json:"metro"`
	Country string `json:"country"`
}

// IXResources represents the resources associated with an IX
type IXResources struct {
	Interface      IXInterface       `json:"interface"`
	BGPConnections []IXBGPConnection `json:"bgp_connection"`
	IPAddresses    []IXIPAddress     `json:"ip_address"`
	VPLSInterface  IXVPLSInterface   `json:"vpls_interface"`
}

// IXInterface represents the physical interface details for an IX
type IXInterface struct {
	Demarcation  string `json:"demarcation"`
	LOATemplate  string `json:"loa_template"`
	Media        string `json:"media"`
	PortSpeed    int    `json:"port_speed"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Up           int    `json:"up"`
	Shutdown     bool   `json:"shutdown"`
}

// IXBGPConnection represents a BGP connection configuration for an IX
type IXBGPConnection struct {
	ASN               int    `json:"asn"`
	CustomerASN       int    `json:"customer_asn"`
	CustomerIPAddress string `json:"customer_ip_address"`
	ISPASN            int    `json:"isp_asn"`
	ISPIPAddress      string `json:"isp_ip_address"`
	IXPeerPolicy      string `json:"ix_peer_policy"`
	MaxPrefixes       int    `json:"max_prefixes"`
	ResourceName      string `json:"resource_name"`
	ResourceType      string `json:"resource_type"`
}

// IXIPAddress represents an IP address allocated for an IX
type IXIPAddress struct {
	Address      string `json:"address"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Version      int    `json:"version"`
	ReverseDNS   string `json:"reverse_dns"`
}

// IXVPLSInterface represents the VPLS interface configuration for an IX
type IXVPLSInterface struct {
	MACAddress    string `json:"mac_address"`
	RateLimitMbps int    `json:"rate_limit_mbps"`
	ResourceName  string `json:"resource_name"`
	ResourceType  string `json:"resource_type"`
	VLAN          int    `json:"vlan"`
	Shutdown      bool   `json:"shutdown"`
}

// IXOrder represents the full structure of an IX order for the API
type IXOrder struct {
	ProductUID    string              `json:"productUid"`    // The productUid of the port to attach the IX to
	AssociatedIXs []AssociatedIXOrder `json:"associatedIxs"` // List of IX configurations to associate with the port
}

// AssociatedIX represents an IX configuration in an IX order
type AssociatedIXOrder struct {
	ProductName        string `json:"productName"`         // Name of the IX
	NetworkServiceType string `json:"networkServiceType"`  // The IX type/network service to connect to (e.g. "Los Angeles IX")
	ASN                int    `json:"asn"`                 // ASN (Autonomous System Number) for BGP peering
	MACAddress         string `json:"macAddress"`          // MAC address for the IX interface
	RateLimit          int    `json:"rateLimit"`           // Rate limit in Mbps
	VLAN               int    `json:"vlan"`                // VLAN ID for the IX connection
	Shutdown           bool   `json:"shutdown"`            // Whether the IX is initially shut down (true) or enabled (false)
	PromoCode          string `json:"promoCode,omitempty"` // Optional promotion code for discounts
}

// IXResponse represents a response from the Megaport IX API after querying an IX.
type IXResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    IX     `json:"data"`
}

// ConvertBuyIXRequestToIXOrder converts a BuyIXRequest to an IX order
func ConvertBuyIXRequestToIXOrder(req BuyIXRequest) []IXOrder {
	return []IXOrder{{
		ProductUID: req.ProductUID,
		AssociatedIXs: []AssociatedIXOrder{
			{
				ProductName:        req.Name,
				NetworkServiceType: req.NetworkServiceType,
				ASN:                req.ASN,
				MACAddress:         req.MACAddress,
				RateLimit:          req.RateLimit,
				VLAN:               req.VLAN,
				Shutdown:           req.Shutdown,
				PromoCode:          req.PromoCode,
			},
		},
	}}
}

// IXUpdate represents the structure for updating an IX
type IXUpdate struct {
	Name           string `json:"name,omitempty"`
	RateLimit      *int   `json:"rateLimit,omitempty"`
	CostCentre     string `json:"costCentre,omitempty"`
	VLAN           *int   `json:"vlan,omitempty"`
	MACAddress     string `json:"macAddress,omitempty"`
	ASN            *int   `json:"asn,omitempty"`
	Password       string `json:"password,omitempty"`
	PublicGraph    *bool  `json:"publicGraph,omitempty"`
	ReverseDns     string `json:"reverseDns,omitempty"`
	AEndProductUid string `json:"aEndProductUid,omitempty"`
	Shutdown       *bool  `json:"shutdown,omitempty"`
}
