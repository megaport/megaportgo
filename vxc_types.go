package megaport

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Partner Providers
const PARTNER_AZURE string = "AZURE"
const PARTNER_GOOGLE string = "GOOGLE"
const PARTNER_AWS string = "AWS"
const PARTNER_OCI string = "ORACLE"

// VXC represents a Virtual Cross Connect in the Megaport VXC API.
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
	Resources          *VXCResources       `json:"resources"`
	VXCApproval        *VXCApproval        `json:"vxcApproval"`
	Shutdown           bool                `json:"shutdown"`
	ContractStartDate  *Time               `json:"contractStartDate"`
	ContractEndDate    *Time               `json:"contractEndDate"`
	ContractTermMonths int                 `json:"contractTermMonths"`
	CompanyUID         string              `json:"companyUid"`
	CompanyName        string              `json:"companyName"`
	CostCentre         string              `json:"costCentre"`
	Locked             bool                `json:"locked"`
	AdminLocked        bool                `json:"adminLocked"`
	AttributeTags      map[string]string   `json:"attributeTags"`
	Cancelable         bool                `json:"cancelable"`
}

var emptyVLL = []byte(`[]`)
var jsonNull = []byte(`null`)

// we need to have a custom unmarshal function because the API returns an "[]" (empty array)
// for vll when the VXC is decommissioned, which will break the regular unmarhsal.
func (w *VXCResources) UnmarshalJSON(data []byte) error {
	// these are all the empty cases for this type
	if len(data) == 0 || bytes.Equal(data, jsonNull) {
		return nil
	}

	type TempVXC VXCResources // A new type that doesn't have UnmarshalJSON method
	if err := json.Unmarshal(data, (*TempVXC)(w)); err != nil {
		return err
	}

	// set this to nil if the json was null. this is a workaround to handle the api response coming back as `[]` for the empty case sometimes
	if w.VLL != nil && w.VLL.isEmpty {
		w.VLL = nil
	}

	return nil
}

// VXCEndConfiguration represents the configuration of an endpoint of a VXC.
type VXCEndConfiguration struct {
	OwnerUID              string                  `json:"ownerUid"`
	UID                   string                  `json:"productUid"`
	Name                  string                  `json:"productName"`
	LocationID            int                     `json:"locationId"`
	Location              string                  `json:"location"`
	VLAN                  int                     `json:"vlan"`
	InnerVLAN             int                     `json:"innerVlan"`
	NetworkInterfaceIndex int                     `json:"vNicIndex"`
	SecondaryName         string                  `json:"secondaryName"`
	LocationDetails       *ProductLocationDetails `json:"locationDetail"`
}

// VXCResources represents the resources associated with a VXC.
type VXCResources struct {
	Interface     []*PortInterface `json:"interface"`
	VirtualRouter *VirtualRouter   `json:"virtual_router"`
	CSPConnection *CSPConnection   `json:"csp_connection"`
	VLL           *VLLConfig       `json:"vll"`
}

// we need to have a custom unmarshal function because the API returns an "[]" (empty array)
// when the VXC is decommissioned, which will break the regular unmarhsal.
func (w *VLLConfig) UnmarshalJSON(data []byte) error {
	// these are all the empty cases for this type
	if len(data) == 0 || bytes.Equal(data, emptyVLL) || bytes.Equal(data, jsonNull) {
		w.isEmpty = true
		return nil
	}

	type TempVLLConfig VLLConfig // A new type that doesn't have UnmarshalJSON method
	return json.Unmarshal(data, (*TempVLLConfig)(w))
}

// VirtualRouter represents the configuration of a virtual router.
type VirtualRouter struct {
	MCRAsn             int    `json:"mcrAsn"`
	ResourceName       string `json:"resource_name"`
	ResourceType       string `json:"resource_type"`
	Speed              int    `json:"speed"`
	BGPShutdownDefault bool   `json:"bgpShutdownDefault"`
}

// VLLConfig represents the configuration of a VLL.
type VLLConfig struct {
	AEndVLAN      int    `json:"a_vlan"`
	BEndVLAN      int    `json:"b_vlan"`
	Description   string `json:"description"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	RateLimitMBPS int    `json:"rate_limit_mbps"`
	ResourceName  string `json:"resource_name"`
	ResourceType  string `json:"resource_type"`
	Shutdown      bool   `json:"shutdown"`

	isEmpty bool // to signal that this field was empty with the API response `[]` we need to do this so we can correctly assign this field as nil
}

// VXCApproval represents the approval status of a VXC.
type VXCApproval struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	UID      string `json:"uid"`
	Type     string `json:"type"`
	NewSpeed int    `json:"newSpeed"`
}

// PartnerLookup represents the response from the Partner Lookup API.
type PartnerLookup struct {
	Bandwidth    int                 `json:"bandwidth"`
	Bandwidths   []int               `json:"bandwidths"`
	Megaports    []PartnerLookupItem `json:"megaports"`
	Peers        []Peer              `json:"peers"`
	ResourceType string              `json:"resource_type"`
	ServiceKey   string              `json:"service_key"`
	VLAN         int                 `json:"vlan"`
}

// PartnerLookupItem represents an item in the Partner Lookup response.
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

// Peer represents a VXC Peer.
type Peer struct {
	PeerASN         int    `json:"peer_asn"`
	Prefixes        string `json:"prefixes"`
	PrimarySubnet   string `json:"primary_subnet"`
	SecondarySubnet string `json:"secondary_subnet"`
	Type            string `json:"type"`
	VLAN            int    `json:"vlan"`
	SharedKey       string `json:"shared_key"`
}

// VXCUpdate represents the fields that can be updated on a VXC.
type VXCUpdate struct {
	Name           string `json:"name,omitempty"`
	RateLimit      *int   `json:"rateLimit,omitempty"`
	CostCentre     string `json:"costCentre,omitempty"`
	Shutdown       *bool  `json:"shutdown,omitempty"`
	AEndVLAN       *int   `json:"aEndVlan,omitempty"`
	BEndVLAN       *int   `json:"bEndVlan,omitempty"`
	AEndInnerVLAN  *int   `json:"aEndInnerVlan,omitempty"`
	BEndInnerVLAN  *int   `json:"bEndInnerVlan,omitempty"`
	AEndProductUID string `json:"aEndProductUid,omitempty"`
	BEndProductUID string `json:"bEndProductUid,omitempty"`
	Term           *int   `json:"term,omitempty"`

	AEndPartnerConfig VXCPartnerConfiguration `json:"aEndConfig,omitempty"`
	BEndPartnerConfig VXCPartnerConfiguration `json:"bEndConfig,omitempty"`
}

// VXCOrderResponse represents the response from the VXC Order API.
type VXCOrderResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []VXCOrderConfirmation `json:"data"`
}

// VXCResponse represents the response from the VXC API.
type VXCResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    VXC    `json:"data"`
}

// VXCOrder represents the request to order a VXC from the Megaport Products API.
type VXCOrder struct {
	AssociatedVXCs []VXCOrderConfiguration `json:"associatedVxcs"`
	PortID         string                  `json:"productUid"`
}

// VXCOrderConfiguration represents the configuration of a VXC to be ordered from the Megaport Products API.
type VXCOrderConfiguration struct {
	Name       string                        `json:"productName"`
	ServiceKey string                        `json:"serviceKey,omitempty"`
	PromoCode  string                        `json:"promoCode,omitempty"`
	RateLimit  int                           `json:"rateLimit"`
	Term       int                           `json:"term"`
	Shutdown   bool                          `json:"shutdown"`
	CostCentre string                        `json:"costCentre,omitempty"`
	AEnd       VXCOrderEndpointConfiguration `json:"aEnd"`
	BEnd       VXCOrderEndpointConfiguration `json:"bEnd"`

	ResourceTags []ResourceTag `json:"resourceTags,omitempty"`
}

// VXCOrderEndpointConfiguration represents the configuration of an endpoint of a VXC to be ordered from the Megaport Products API.
type VXCOrderEndpointConfiguration struct {
	ProductUID    string                  `json:"productUid,omitempty"`
	VLAN          int                     `json:"vlan,omitempty"`
	DiversityZone string                  `json:"diversityZone,omitempty"`
	PartnerConfig VXCPartnerConfiguration `json:"partnerConfig,omitempty"`
	*VXCOrderMVEConfig
}

// VXCPartnerConfiguration represents the configuration of a VXC partner.
type VXCPartnerConfiguration interface {
	IsParnerConfiguration()
}

// VXCPartnerConfigAWS represents the configuration of a VXC partner for AWS Virtual Interface.
type VXCPartnerConfigAWS struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string `json:"connectType"`
	Type                    string `json:"type"`
	OwnerAccount            string `json:"ownerAccount"`
	ASN                     int    `json:"asn,omitempty"`
	AmazonASN               int    `json:"amazonAsn,omitempty"`
	AuthKey                 string `json:"authKey,omitempty"`
	Prefixes                string `json:"prefixes,omitempty"`
	CustomerIPAddress       string `json:"customerIpAddress,omitempty"`
	AmazonIPAddress         string `json:"amazonIpAddress,omitempty"`
	ConnectionName          string `json:"name,omitempty"`
}

// VXCPartnerConfigAzure represents the configuration of a VXC partner for Azure ExpressRoute.
type VXCPartnerConfigAzure struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string                           `json:"connectType"`
	ServiceKey              string                           `json:"serviceKey"`
	Peers                   []PartnerOrderAzurePeeringConfig `json:"peers"`
}

// VXCPartnerConfigGoogle represents the configuration of a VXC partner for Google Cloud Interconnect.
type VXCPartnerConfigGoogle struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string `json:"connectType"`
	PairingKey              string `json:"pairingKey"`
}

// VXCPartnerConfigOracle represents the configuration of a VXC partner for Oracle Cloud Infrastructure FastConnect.
type VXCPartnerConfigOracle struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string `json:"connectType"`
	VirtualCircuitId        string `json:"virtualCircuitId"`
}

type VXCPartnerConfigTransit struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string `json:"connectType"`
}

// VXCPartnerConfigIBM represents the configuration of a VXC partner for IBM Cloud Direct Link.
type VXCPartnerConfigIBM struct {
	VXCPartnerConfiguration `json:"-"`
	ConnectType             string `json:"connectType"`
	AccountID               string `json:"account_id"`          // Customer's IBM Acount ID.  32 Hexadecimal Characters. REQUIRED
	CustomerASN             int    `json:"customer_asn"`        // Customer's ASN. Valid ranges: 1-64495, 64999, 131072-4199999999, 4201000000-4201064511. Required unless the connection at the other end of the VXC is an MCR.
	Name                    string `json:"name"`                // Description of this connection for identification purposes. Max 100 characters from 0-9 a-z A-Z / - _ , Defaults to "MEGAPORT".
	CustomerIPAddress       string `json:"customer_ip_address"` // IPv4 network address including subnet mask. Default is /30 assigned from 169.254.0.0/16.
	ProviderIPAddress       string `json:"provider_ip_address"` // IPv4 network address including subnet mask. Default is /30 assigned from 169.254.0.0/16. Must be in the same subnet as customer_ip_address.
}

// VXCOrderMVEConfig represents the configuration of a VXC endpoint for MVE.
type VXCOrderMVEConfig struct {
	InnerVLAN             int `json:"innerVlan,omitempty"`
	NetworkInterfaceIndex int `json:"vNicIndex"`
}

// VXCOrderVrouterPartnerConfig represents the configuration of a VXC Vrouter Configuration partner.
type VXCOrderVrouterPartnerConfig struct {
	VXCPartnerConfiguration `json:"-"`
	Interfaces              []PartnerConfigInterface `json:"interfaces,omitempty"`
}

// DEPRECATED - Use VXCOrderVrouterPartnerConfig instead
// VXCOrderAEndPartnerConfig represents the configuration of a VXC A-End partner.
type VXCOrderAEndPartnerConfig struct {
	VXCPartnerConfiguration `json:"-"`
	Interfaces              []PartnerConfigInterface `json:"interfaces,omitempty"`
}

// VXCOrderConfirmation represents the confirmation of a VXC order from the Megaport Products API.
type VXCOrderConfirmation struct {
	TechnicalServiceUID string `json:"vxcJTechnicalServiceUid"`
}

// BGP CONFIG STUFF

// PartnerConfigInterface represents the configuration of a partner interface.
type PartnerConfigInterface struct {
	IpAddresses    []string              `json:"ipAddresses,omitempty"`
	IpRoutes       []IpRoute             `json:"ipRoutes,omitempty"`
	NatIpAddresses []string              `json:"natIpAddresses,omitempty"`
	Bfd            BfdConfig             `json:"bfd,omitempty"`
	BgpConnections []BgpConnectionConfig `json:"bgpConnections,omitempty"`
	VLAN           int                   `json:"vlan,omitempty"`
}

// IpRoute represents an IP route.
type IpRoute struct {
	Prefix      string `json:"prefix"`
	Description string `json:"description,omitempty"`
	NextHop     string `json:"nextHop"`
}

// BfdConfig represents the configuration of BFD.
type BfdConfig struct {
	TxInterval int `json:"txInterval,omitempty"`
	RxInterval int `json:"rxInterval,omitempty"`
	Multiplier int `json:"multiplier,omitempty"`
}

// BgpConnectionConfig represents the configuration of a BGP connection.
type BgpConnectionConfig struct {
	PeerAsn            int      `json:"peerAsn"`
	LocalAsn           *int     `json:"localAsn,omitempty"`
	LocalIpAddress     string   `json:"localIpAddress"`
	PeerIpAddress      string   `json:"peerIpAddress"`
	Password           string   `json:"password,omitempty"`
	Shutdown           bool     `json:"shutdown"`
	Description        string   `json:"description,omitempty"`
	MedIn              int      `json:"medIn,omitempty"`
	MedOut             int      `json:"medOut,omitempty"`
	BfdEnabled         bool     `json:"bfdEnabled"`
	ExportPolicy       string   `json:"exportPolicy,omitempty"`
	PermitExportTo     []string `json:"permitExportTo,omitempty"`
	DenyExportTo       []string `json:"denyExportTo,omitempty"`
	ImportWhitelist    int      `json:"importWhitelist,omitempty"`
	ImportBlacklist    int      `json:"importBlacklist,omitempty"`
	ExportWhitelist    int      `json:"exportWhitelist,omitempty"`
	ExportBlacklist    int      `json:"exportBlacklist,omitempty"`
	AsPathPrependCount int      `json:"asPathPrependCount,omitempty"`
	PeerType           string   `json:"peerType,omitempty"` // can be NON_CLOUD, PRIV_CLOUD, or PUB_CLOUD
}

// AWS STUFF

// AWSVXCOrder represents the request to order an AWS VXC from the Megaport Products API.
type AWSVXCOrder struct {
	AssociatedVXCs []AWSVXCOrderConfiguration `json:"associatedVxcs"`
	PortID         string                     `json:"productUid"`
}

// AWSVXCOrderConfiguration represents the configuration of an AWS VXC to be ordered from the Megaport Products API.
type AWSVXCOrderConfiguration struct {
	Name      string                        `json:"productName"`
	RateLimit int                           `json:"rateLimit"`
	AEnd      VXCOrderEndpointConfiguration `json:"aEnd"`
	BEnd      VXCOrderEndpointConfiguration `json:"bEnd"`
}

// Partner

// PartnerOrder represents the request to order a partner VXC from the Megaport Products API.
type PartnerOrder struct {
	PortID         string                 `json:"productUid"`
	AssociatedVXCs []PartnerOrderContents `json:"associatedVxcs"`
}

// PartnerOrderContents represents the configuration of a partner VXC to be ordered from the Megaport Products API.
type PartnerOrderContents struct {
	Name      string                        `json:"productName"`
	RateLimit int                           `json:"rateLimit"`
	AEnd      VXCOrderEndpointConfiguration `json:"aEnd"`
	BEnd      VXCOrderEndpointConfiguration `json:"bEnd"`
}

// PartnerOrderAzurePeeringConfig represents the configuration of an Azure peering partner.
type PartnerOrderAzurePeeringConfig struct {
	Type            string `json:"type"`
	PeerASN         string `json:"peer_asn,omitempty"`
	PrimarySubnet   string `json:"primary_subnet,omitempty"`
	SecondarySubnet string `json:"secondary_subnet,omitempty"`
	Prefixes        string `json:"prefixes,omitempty"`
	SharedKey       string `json:"shared_key,omitempty"`
	VLAN            int    `json:"vlan,omitempty"`
}

// CSPConnection represents the configuration of a CSP connection.
type CSPConnection struct {
	CSPConnection []CSPConnectionConfig
}

// CSPConnectionConfig represents the configuration of a CSP connection.
type CSPConnectionConfig interface {
	IsCSPConnectionConfig()
}

// CSPConnectionAWS represents the configuration of a CSP connection for AWS Virtual Interface.
type CSPConnectionAWS struct {
	CSPConnectionConfig
	ConnectType       string `json:"connectType"`
	ResourceName      string `json:"resource_name"`
	ResourceType      string `json:"resource_type"`
	VLAN              int    `json:"vlan"`
	Account           string `json:"account"`
	AmazonAddress     string `json:"amazon_address"`
	ASN               int    `json:"asn"`
	AmazonASN         int    `json:"amazonAsn"`
	AuthKey           string `json:"authKey"`
	CustomerAddress   string `json:"customer_address"`
	CustomerIPAddress string `json:"customerIpAddress"`
	ID                int    `json:"id"`
	Name              string `json:"name"`
	OwnerAccount      string `json:"ownerAccount"`
	PeerASN           int    `json:"peerAsn"`
	Type              string `json:"type"`
	VIFID             string `json:"vif_id"`
}

// CSPConnectionAWSHC represents the configuration of a CSP connection for AWS Hosted Connection.
type CSPConnectionAWSHC struct {
	CSPConnectionConfig
	ConnectType  string `json:"connectType"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Bandwidth    int    `json:"bandwidth"`
	Name         string `json:"name"`
	OwnerAccount string `json:"ownerAccount"`
	Bandwidths   []int  `json:"bandwidths"`
	ConnectionID string `json:"connectionId"`
}

// CSPConnectionAzure represents the configuration of a CSP connection for Azure ExpressRoute.
type CSPConnectionAzure struct {
	CSPConnectionConfig
	ConnectType  string                            `json:"connectType"`
	ResourceName string                            `json:"resource_name"`
	ResourceType string                            `json:"resource_type"`
	Bandwidth    int                               `json:"bandwidth"`
	Managed      bool                              `json:"managed"`
	Megaports    []CSPConnectionAzureMegaport      `json:"megaports"`
	Ports        []CSPConnectionAzurePort          `json:"ports"`
	ServiceKey   string                            `json:"service_key"`
	VLAN         int                               `json:"vlan"`
	Peers        []CSPConnectionAzurePeeringConfig `json:"peers"`
}

// CSPConnectionAzurePeeringConfig represents the configuration of an Azure peering partner.
type CSPConnectionAzurePeeringConfig struct {
	Type            string `json:"type"`
	PeerASN         int    `json:"peer_asn"`
	PrimarySubnet   string `json:"primary_subnet"`
	SecondarySubnet string `json:"secondary_subnet"`
	Prefixes        string `json:"prefixes,omitempty"`
	SharedKey       string `json:"shared_key,omitempty"`
	VLAN            int    `json:"vlan"`
}

// CSPConnectionAzureMegaport represents the configuration of a CSP connection for Azure ExpressRoute megaport.
type CSPConnectionAzureMegaport struct {
	Port int    `json:"port"`
	Type string `json:"type"`
	VXC  int    `json:"vxc,omitempty"`
}

// CSPConnectionAzurePort represents the configuration of a CSP connection for Azure ExpressRoute port.
type CSPConnectionAzurePort struct {
	ServiceID     int    `json:"service_id"`
	Type          string `json:"type"`
	VXCServiceIDs []int  `json:"vxc_service_ids"`
}

// CSPConnectionGoogle represents the configuration of a CSP connection for Google Cloud Interconnect.
type CSPConnectionGoogle struct {
	CSPConnectionConfig
	Bandwidth    int                           `json:"bandwidth"`
	ConnectType  string                        `json:"connectType"`
	ResourceName string                        `json:"resource_name"`
	ResourceType string                        `json:"resource_type"`
	Bandwidths   []int                         `json:"bandwidths"`
	Megaports    []CSPConnectionGoogleMegaport `json:"megaports"`
	Ports        []CSPConnectionGooglePort     `json:"ports"`
	CSPName      string                        `json:"csp_name"`
	PairingKey   string                        `json:"pairingKey"`
}

// CSPConnectionGoogleMegaport represents the configuration of a CSP connection for Google Cloud Interconnect megaport.
type CSPConnectionGoogleMegaport struct {
	Port int `json:"port"`
	VXC  int `json:"vxc"`
}

// CSPConnectionGooglePort represents the configuration of a CSP connection for Google Cloud Interconnect port.
type CSPConnectionGooglePort struct {
	ServiceID     int   `json:"service_id"`
	VXCServiceIDs []int `json:"vxc_service_ids"`
}

// CSPConnectionVirtualRouter represents the configuration of a CSP connection for Virtual Router.
type CSPConnectionVirtualRouter struct {
	CSPConnectionConfig
	ConnectType       string                                `json:"connectType"`
	ResourceName      string                                `json:"resource_name"`
	ResourceType      string                                `json:"resource_type"`
	VLAN              int                                   `json:"vlan"`
	Interfaces        []CSPConnectionVirtualRouterInterface `json:"interfaces"`
	IPAddresses       []string                              `json:"ip_addresses"`
	VirtualRouterName string                                `json:"virtualRouterName"`
}

// CSPConnectionVirtualRouterInterface represents the configuration of a CSP connection for Virtual Router interface.
type CSPConnectionVirtualRouterInterface struct {
	IPAddresses    []string              `json:"ipAddresses"`
	IPRoutes       []IpRoute             `json:"ipRoutes"`
	BGPConnections []BgpConnectionConfig `json:"bgpConnections"`
	NatIPAddresses []string              `json:"natIpAddresses"`
	BFD            BfdConfig             `json:"bfd"`
}

type CSPConnectionOracle struct {
	CSPConnectionConfig
	ConnectType      string                        `json:"connectType"`
	ResourceName     string                        `json:"resource_name"`
	ResourceType     string                        `json:"resource_type"`
	CSPName          string                        `json:"csp_name"`
	Bandwidth        int                           `json:"bandwidth"`
	Megaports        []CSPConnectionOracleMegaport `json:"megaports"`
	Ports            []CSPConnectionOraclePort     `json:"ports"`
	VirtualCircuitId string                        `json:"virtualCircuitId"`
}

type CSPConnectionOracleMegaport struct {
	Port int `json:"port"`
	VXC  int `json:"vxc"`
}

type CSPConnectionOraclePort struct {
	ServiceID         int   `json:"service_id"`
	VXCServiceIDs     []int `json:"vxc_service_ids"`
	FirstVXCServiceID int   `json:"firstVxcServiceId,omitempty"`
}

// CSPConnectionTransit represents the configuration of a CSP connection for a Transit VXC.
type CSPConnectionTransit struct {
	CSPConnectionConfig
	ConnectType        string `json:"connectType"`
	ResourceName       string `json:"resource_name"`
	ResourceType       string `json:"resource_type"`
	CustomerIP4Address string `json:"customer_ip4_address"`
	CustomerIP6Network string `json:"customer_ip6_network"`
	IPv4GatewayAddress string `json:"ipv4_gateway_address"`
	IPv6GatewayAddress string `json:"ipv6_gateway_address"`
}

// CSPConnectionIBM represents the configuration of a CSP connection for IBM Cloud Direct Link.
type CSPConnectionIBM struct {
	CSPConnectionConfig
	ConnectType       string `json:"connectType"`
	ResourceName      string `json:"resource_name"`
	ResourceType      string `json:"resource_type"`
	CSPName           string `json:"csp_name"`
	Bandwidth         int    `json:"bandwidth"`
	AccountID         string `json:"account_id"`
	CustomerASN       int    `json:"customer_asn"`
	ProviderIPAddress string `json:"provider_ip_address"`
	CustomerIPAddress string `json:"customer_ip_address"`
	Bandwidths        []int  `json:"bandwidths"`
}

// CSPConnectionOther represents the configuration of a CSP connection for any other CSP that is not presently defined.
type CSPConnectionOther struct {
	CSPConnectionConfig
	CSPConnection map[string]interface{}
}

// UnmarshalJSON is a custom unmarshaler for the CSPConnection type.
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
		case "VROUTER":
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
		case "ORACLE":
			marshaled, err := json.Marshal(cn)
			if err != nil {
				return err
			}
			oracle := CSPConnectionOracle{}
			if err := json.Unmarshal(marshaled, &oracle); err != nil {
				return err
			}
			c.CSPConnection = append(c.CSPConnection, oracle)
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
	case []interface{}:
		for _, m := range v {
			cn, ok := m.(map[string]interface{})
			if !ok {
				return fmt.Errorf("can't process CSP connections, expected map[string]interface{} but got %T", m)
			}

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
			case "VROUTER":
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
			case "ORACLE":
				marshaled, err := json.Marshal(cn)
				if err != nil {
					return err
				}
				oracle := CSPConnectionOracle{}
				if err := json.Unmarshal(marshaled, &oracle); err != nil {
					return err
				}
				c.CSPConnection = append(c.CSPConnection, oracle)
			case "IBM":
				marshaled, err := json.Marshal(cn)
				if err != nil {
					return err
				}
				ibm := CSPConnectionIBM{}
				if err := json.Unmarshal(marshaled, &ibm); err != nil {
					return err
				}
				c.CSPConnection = append(c.CSPConnection, ibm)
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
