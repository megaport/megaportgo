package megaport

import (
	"encoding/json"
)

const PARTNER_AZURE string = "AZURE"
const PARTNER_GOOGLE string = "GOOGLE"
const PARTNER_AWS string = "AWS"
const PARTNER_OCI string = "ORACLE"

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
	LiveDate           *Time               `json:"liveDate"`
	CreateDate         *Time               `json:"createDate"`
	Resources          VXCResources        `json:"resources"`
	VXCApproval        VXCApproval         `json:"vxcApproval"`
	ContractStartDate  *Time               `json:"contractStartDate"`
	ContractEndDate    *Time               `json:"contractEndDate"`
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
	Interface     []PortInterface 		  `json:"interface"`
	VirtualRouter VirtualRouter   		  `json:"virtual_router"`
	CSPConnection CSPConnection	 		  `json:"csp_connection"`
	VLL           VLLConfig       		  `json:"vll"`
}

type VirtualRouter struct {
	MCRAsn				int		`json:"mcrAsn"`
	ResourceName 		string	`json:"resource_name"`
	ResourceType 		string	`json:"resource_type"`
	Speed				int		`json:"speed"`
	BGPShutdownDefault	bool 	`json:"bgpShutdownDefault"`
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
	Peers        []Peer       		 `json:"peers"`
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

type Peer struct {
	PeerASN string `json:"peer_asn"`
	Prefixes string `json:"prefixes"`
	PrimarySubnet string `json:"primary_subnet"`
	SecondarySubnet string `json:"secondary_subnet"`
	Type string `json:"type"`
	VLAN string `json:"vlan"`
	SharedKey string `json:"shared_key"`
}

type VXCUpdate struct {
	Name      *string `json:"name,omitempty"`
	RateLimit *int    `json:"rateLimit,omitempty"`
	AEndVLAN  *int    `json:"aEndVlan,omitempty"`
	BEndVLAN  *int    `json:"bEndVlan,omitempty"`
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
	Name		string 							`json:"productName"`
	RateLimit 	int 							`json:"rateLimit"`
	AEnd 		VXCOrderEndpointConfiguration 	`json:"aEnd"`
	BEnd 		VXCOrderEndpointConfiguration 	`json:"bEnd"`
}

type VXCOrderEndpointConfiguration struct {
	ProductUID 		string 					`json:"productUid,omitempty"`
	VLAN 			int 					`json:"vlan,omitempty"`
	PartnerConfig 	VXCPartnerConfiguration `json:"partnerConfig,omitempty"`
	*VXCOrderMVEConfig
}

type VXCPartnerConfiguration interface {
	IsParnerConfiguration()
}

type VXCPartnerConfigAWS struct {
	VXCPartnerConfiguration
	ConnectType 	  string `json:"connectType"`
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

type VXCPartnerConfigAzure struct {
	VXCPartnerConfiguration
	ConnectType 	  	string 							 `json:"connectType"`
	ServiceKey  		string                           `json:"serviceKey"`
	Peers       		[]PartnerOrderAzurePeeringConfig `json:"peers"`
}

type VXCPartnerConfigGoogle struct {
	VXCPartnerConfiguration
	ConnectType 	string `json:"connectType"`
	PairingKey  	string `json:"pairingKey"`
}

type VXCPartnerConfigOracle struct {
	VXCPartnerConfiguration
	ConnectType 	  string `json:"connectType"`
	VirtualCircutId   string `json:"virtualCircuitId"`
}

type VXCOrderMVEConfig struct {
	InnerVLAN             int `json:"innerVlan,omitempty"`
	NetworkInterfaceIndex int `json:"vNicIndex"`
}

type VXCOrderAEndPartnerConfig struct {
	VXCPartnerConfiguration
	Interfaces []PartnerConfigInterface `json:"interfaces,omitempty"`
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
	Name      string                       		`json:"productName"`
	RateLimit int                          		`json:"rateLimit"`
	AEnd      VXCOrderEndpointConfiguration    	`json:"aEnd"`
	BEnd      VXCOrderEndpointConfiguration 	`json:"bEnd"`
}

// Partner

type PartnerOrder struct {
	PortID         string                 `json:"productUid"`
	AssociatedVXCs []PartnerOrderContents `json:"associatedVxcs"`
}

type PartnerOrderContents struct {
	Name      string                        	`json:"productName"`
	RateLimit int                           	`json:"rateLimit"`
	AEnd      VXCOrderEndpointConfiguration     `json:"aEnd"`
	BEnd      VXCOrderEndpointConfiguration 	`json:"bEnd"`
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

type CSPConnection struct {
	CSPConnection []CSPConnectionConfig
}

type CSPConnectionConfig interface {
	IsCSPConnectionConfig()
}

type CSPConnectionAWS struct {
	CSPConnectionConfig
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	VLAN int `json:"vlan"`
	Account string `json:"account"`
	AmazonAddress string `json:"amazon_address"`
	ASN int `json:"asn"`
	AuthKey string `json:"authKey"`
	CustomerAddress string `json:"customer_address"`
	CustomerIPAddress string `json:"customerIpAddress"`
	ID int `json:"id"`
	Name string `json:"name"`
	OwnerAccount string `json:"ownerAccount"`
	PeerASN int `json:"peerAsn"`
	Type string `json:"type"`
	VIFID string `json:"vif_id"`
}

type CSPConnectionAWSHC struct {
	CSPConnectionConfig
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Bandwidth int `json:"bandwidth"`
	Name string `json:"name"`
	OwnerAccount string `json:"ownerAccount"`
	Bandwidths []int `json:"bandwidths"`
	ConnectionID string `json:"connectionId"`
}

type CSPConnectionAzure struct {
	CSPConnectionConfig
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Bandwidth int `json:"bandwidth"`
	Managed bool `json:"managed"`
	Megaports []CSPConnectionAzureMegaport `json:"megaports"`
	Ports []CSPConnectionAzurePort `json:"ports"`
	ServiceKey string `json:"service_key"`
	VLAN int `json:"vlan"`
}
type CSPConnectionAzureMegaport struct {
	Port int `json:"port"`
	Type string `json:"type"`
	VXC int `json:"vxc,omitempty"`
}

type CSPConnectionAzurePort struct {
	ServiceID int `json:"service_id"`
	Type string `json:"type"`
	VXCServiceIDs []int `json:"vxc_service_ids"`
}

type CSPConnectionGoogle struct {
	CSPConnectionConfig
	Bandwidth int `json:"bandwidth"`
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Bandwidths []int `json:"bandwidths"`
	Megaports []CSPConnectionGoogleMegaport `json:"megaports"`
	Ports []CSPConnectionGooglePort `json:"ports"`
	CSPName string `json:"csp_name"`
	PairingKey string `json:"pairingKey"`
}

type CSPConnectionGoogleMegaport struct {
	Port int `json:"port"`
	VXC int `json:"vxc"`
}

type CSPConnectionGooglePort struct {
	ServiceID int `json:"service_id"`
	VXCServiceIDs []int `json:"vxc_service_ids"`
}

type CSPConnectionVirtualRouter struct {
	CSPConnectionConfig
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	VLAN int `json:"vlan"`
	Interfaces []CSPConnectionVirtualRouterInterface `json:"interfaces"`
	IPAddresses []string `json:"ip_addresses"`
	VirtualRouterName string `json:"virtualRouterName"`
}

type CSPConnectionVirtualRouterInterface struct {
	IPAddresses []string `json:"ipAddresses"`
}

type CSPConnectionTransit struct {
	CSPConnectionConfig
	ConnectType string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	CustomerIP4Address string `json:"customer_ip4_address"`
	CustomerIP6Network string `json:"customer_ip6_network"`
	IPv4GatewayAddress string `json:"ipv4_gateway_address"`
	IPv6GatewayAddress string `json:"ipv6_gateway_address"`
}

type CSPConnectionOther struct {
	CSPConnectionConfig
	CSPConnection map[string]interface{}
}


func (c *CSPConnection) UnmarshalJSON(data []byte) error {
	c.CSPConnection = []CSPConnectionConfig{}
	var i interface{}
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	switch v := i.(type) {
		case map[string]interface{}:
			cn := v
			switch v["connectType"] {
				case "AWSHC":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					awsHC := CSPConnectionAWSHC{}
					if err := json.Unmarshal(marshaled, &awsHC); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, awsHC)
				case "AWS":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					aws := CSPConnectionAWS{}
					if err := json.Unmarshal(marshaled, &aws); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, aws)
				case "GOOGLE":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					google := CSPConnectionGoogle{}
					if err := json.Unmarshal(marshaled, &google); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, google)
				case "AZURE":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					azure := CSPConnectionAzure{}
					if err := json.Unmarshal(marshaled, &azure); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, azure)
				case "VIRTUAL_ROUTER":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					vr := CSPConnectionVirtualRouter{}
					if err := json.Unmarshal(marshaled, &vr); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, vr)
				case "TRANSIT":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					transit := CSPConnectionTransit{}
					if err := json.Unmarshal(marshaled, &transit); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, transit)
				default: // Any other cases will be marshaled into a map[string]interface{}
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					other := CSPConnectionOther{}
					cspMap := map[string]interface{}{}
					if err := json.Unmarshal(marshaled, &cspMap); err != nil {
						return err
					}
					other.CSPConnection = cspMap
					c.CSPConnection = append(c.CSPConnection, other)
		}
	 	case[]interface{}:
			for _, m := range v {
				cn := m.(map[string]interface{})
				switch cn["connectType"] {
				case "AWSHC":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					awsHC := CSPConnectionAWSHC{}
					if err := json.Unmarshal(marshaled, &awsHC); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, awsHC)
				case "AWS":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					aws := CSPConnectionAWS{}
					if err := json.Unmarshal(marshaled, &aws); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, aws)
				case "GOOGLE":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					google := CSPConnectionGoogle{}
					if err := json.Unmarshal(marshaled, &google); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, google)
				case "AZURE":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					azure := CSPConnectionAzure{}
					if err := json.Unmarshal(marshaled, &azure); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, azure)
				case "VIRTUAL_ROUTER":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					vr := CSPConnectionVirtualRouter{}
					if err := json.Unmarshal(marshaled, &vr); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, vr)
				case "TRANSIT":
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					transit := CSPConnectionTransit{}
					if err := json.Unmarshal(marshaled, &transit); err != nil {
						return err
					}
					c.CSPConnection = append(c.CSPConnection, transit)
				default: // Any other cases will be marshaled into a map[string]interface{}
					marshaled, err := json.Marshal(cn)
					if err != nil {
						return err
					}
					other := CSPConnectionOther{}
					cspMap := map[string]interface{}{}
					if err := json.Unmarshal(marshaled, &cspMap); err != nil {
						return err
					}
					other.CSPConnection = cspMap
					c.CSPConnection = append(c.CSPConnection, other)
			}
		}
	}
	return nil
}