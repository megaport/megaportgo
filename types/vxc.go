// Copyright 2020 Megaport Pty Ltd
//
// Licensed under the Mozilla Public License, Version 2.0 (the
// "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//       https://mozilla.org/MPL/2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

// ---- VXC Order Types //
type VXCOrder struct {
	AssociatedVXCs []VXCConfiguration `json:"associatedVxcs"`
	PortID         string             `json:"productUid"`
}

type AWSHostedVIFOrder struct {
	AssociatedVXCs []AWSHostedVIFOrderConfiguration `json:"associatedVxcs"`
	PortID         string                           `json:"productUid"`
}

type VXCConfiguration struct {
	Name      string                    `json:"productName"`
	RateLimit int                       `json:"rateLimit"`
	AEnd      VXCOrderAEndConfiguration `json:"aEnd"`
	BEnd      VXCOrderBEndConfiguration `json:"bEnd"`
}

type AWSHostedVIFOrderConfiguration struct {
	Name      string                             `json:"productName"`
	RateLimit int                                `json:"rateLimit"`
	AEnd      VXCOrderAEndConfiguration          `json:"aEnd"`
	BEnd      AWSHostedVIFOrderBEndConfiguration `json:"bEnd"`
}

type VXCOrderAEndConfiguration struct {
	VLAN int `json:"vlan"`
}

type VXCOrderBEndConfiguration struct {
	VLAN       int    `json:"vlan,omitempty"`
	ProductUID string `json:"productUid"`
}

type AWSHostedVIFOrderBEndConfiguration struct {
	ProductUID    string                         `json:"productUid"`
	PartnerConfig AWSHostedVIFOrderPartnerConfig `json:"partnerConfig"`
}

type AWSHostedVIFOrderPartnerConfig struct {
	ConnectType       string `json:"connectType"`
	Type              string `json:"type"`
	ASN               int    `json:"asn"`
	AmazonASN         int    `json:"amazonAsn"`
	OwnerAccount      string `json:"ownerAccount"`
	AuthKey           string `json:"authKey,omitempty"`
	Prefixes          string `json:"prefixes,omitempty"`
	CustomerIPAddress string `json:"customerIpAddress,omitempty"`
	AmazonIPAddress   string `json:"amazonIpAddress,omitempty"`
}

type VXCOrderConfirmation struct {
	TechnicalServiceUID string `json:"vxcJTechnicalServiceUid"`
}

// ---- VXC Detail Types //
type VXC struct {
	ID                 int                    `json:"productId"`
	UID                string                 `json:"productUid"`
	ServiceID          int                    `json:"nServiceId"`
	Name               string                 `json:"productName"`
	Type               string                 `json:"productType"`
	RateLimit          int                    `json:"rateLimit"`
	DistanceBand       string                 `json:"distanceBand"`
	ProvisioningStatus string                 `json:"provisioningStatus"`
	AEndConfiguration  VXCEndConfiguration    `json:"aEnd"`
	BEndConfiguration  VXCEndConfiguration    `json:"bEnd"`
	SecondaryName      string                 `json:"secondaryName"`
	UsageAlgorithm     string                 `json:"usageAlgorithm"`
	CreatedBy          string                 `json:"createdBy"`
	LiveDate           int                    `json:"liveDate"`
	CreateDate         int                    `json:"createDate"`
	Resources          VXCResources           `json:"resources"`
	VXCApproval        VXCApproval            `json:"vxcApproval"`
	ContractStartDate  int                    `json:"contractStartDate"`
	ContractEndDate    int                    `json:"contractEndDate"`
	ContractTermMonths int                    `json:"contractTermMonths"`
	CompanyUID         string                 `json:"companyUid"`
	CompanyName        string                 `json:"companyName"`
	Locked             bool                   `json:"locked"`
	AdminLocked        bool                   `json:"adminLocked"`
	AttributeTags      map[string]string      `json:"attributeTags"`
	Cancelable         bool                   `json:"cancelable"`
}

type VXCEndConfiguration struct {
	OwnerUID      string `json:"ownerUid"`
	UID           string `json:"productUid"`
	Name          string `json:"productName"`
	LocationID    int    `json:"locationId"`
	Location      string `json:"location"`
	VLAN          int    `json:"vlan"`
	SecondaryName string `json:"secondaryName"`
}

type VXCResources struct {
	Interface     []PortInterface          `json:"interface"`
	VirtualRouter map[string]interface{}   `json:"virtual_router"`
	CspConnection interface{}  `json:"csp_connection"`
	VLL           VLLConfig                `json:"vll"`
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

type PartnerOrder struct {
	PortID         string                 `json:"productUid"`
	AssociatedVXCs []PartnerOrderContents `json:"associatedVxcs"`
}

type PartnerOrderContents struct {
	Name      string                        `json:"productName"`
	RateLimit int                           `json:"rateLimit"`
	AEnd      PartnerOrderAEndConfiguration `json:"aEnd"`
	BEnd      PartnerOrderBEndConfiguration `json:"bEnd"`
}

type PartnerOrderAEndConfiguration struct {
	VLAN int `json:"vlan,omitempty"`
}

type PartnerOrderBEndConfiguration struct {
	PartnerPortID string      `json:"productUid"`
	PartnerConfig interface{} `json:"partnerConfig"`
}

type PartnerOrderAzurePartnerConfig struct {
	ConnectType string `json:"connectType"`
	ServiceKey  string `json:"serviceKey"`
	Peers       []map[string]string `json:"peers"`
}

type PartnerOrderGooglePartnerConfig struct {
	ConnectType string `json:"connectType"`
	PairingKey  string `json:"pairingKey"`
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
