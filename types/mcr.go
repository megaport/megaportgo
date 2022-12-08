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

type MCROrder struct {
	LocationID int            `json:"locationId"`
	Name       string         `json:"productName"`
	Type       string         `json:"productType"`
	PortSpeed  int            `json:"portSpeed"`
	Config     MCROrderConfig `json:"config"`
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
	Description   string               `json:"description"`
	AddressFamily string               `json:"addressFamily"`
	Entries       []MCRPrefixListEntry `json:"entries"`
}

type MCRPrefixListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"`
	Le     int    `json:"le,omitempty"`
}
