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

type MVEOrderConfig struct {
	LocationID        int                    `json:"locationId"`
	Name              string                 `json:"productName"`
	Term              int                    `json:"term"`
	ProductType       string                 `json:"productType"`
	NetworkInterfaces []*MVENetworkInterface `json:"vnics"`
	VendorConfig      map[string]interface{} `json:"vendorConfig"`
}

// NetworkInterface represents a vNIC.
type MVENetworkInterface struct {
	Description string `json:"description"`
	VLAN        int    `json:"vlan"`
}

// InstanceSize encodes the available MVE instance sizes.
type MVEInstanceSize string

const (
	SMALL  MVEInstanceSize = "SMALL"
	MEDIUM MVEInstanceSize = "MEDIUM"
	LARGE  MVEInstanceSize = "LARGE"
	XLARGE MVEInstanceSize = "X_LARGE_12"
)

type MVEOrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

type MVE struct {
	ID                    int                    `json:"productId"`
	UID                   string                 `json:"productUid"`
	Name                  string                 `json:"productName"`
	Type                  string                 `json:"productType"`
	ProvisioningStatus    string                 `json:"provisioningStatus"`
	CreateDate            int                    `json:"createDate"`
	CreatedBy             string                 `json:"createdBy"`
	TerminateDate         int                    `json:"terminateDate"`
	LiveDate              int                    `json:"liveDate"`
	Market                string                 `json:"market"`
	LocationID            int                    `json:"locationId"`
	UsageAlgorithm        string                 `json:"usageAlgorithm"`
	MarketplaceVisibility bool                   `json:"marketplaceVisibility"`
	VXCPermitted          bool                   `json:"vxcpermitted"`
	VXCAutoApproval       bool                   `json:"vxcAutoApproval"`
	SecondaryName         string                 `json:"secondaryName"`
	CompanyUID            string                 `json:"companyUid"`
	CompanyName           string                 `json:"companyName"`
	ContractStartDate     int                    `json:"contractStartDate"`
	ContractEndDate       int                    `json:"contractEndDate"`
	ContractTermMonths    int                    `json:"contractTermMonths"`
	AttributeTags         map[string]string      `json:"attributeTags"`
	Virtual               bool                   `json:"virtual"`
	BuyoutPort            bool                   `json:"buyoutPort"`
	Locked                bool                   `json:"locked"`
	AdminLocked           bool                   `json:"adminLocked"`
	Cancelable            bool                   `json:"cancelable"`
	Resources             map[string]interface{} `json:"resources"`
	Vendor                string                 `json:"vendor"`
	Size                  string                 `json:"mveSize"`
	NetworkInterfaces     []*MVENetworkInterface `json:"vnics"`
}
