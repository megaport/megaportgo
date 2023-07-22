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

type ProductUpdate struct {
	Name                 string `json:"name"`
	CostCentre           string `json:"costCentre"`
	MarketplaceVisbility bool   `json:"marketplaceVisibility"`
}

type ProductTelemetry struct {
	ServiceUid string `json:"serviceUid"`
	Type       string `json:"type"`
	TimeFrame  struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	} `json:"timeFrame"`
	Data []struct {
		Type    string      `json:"type"`
		Subtype string      `json:"subtype"`
		Samples [][]float64 `json:"samples"`
		Unit    struct {
			Name     string `json:"name"`
			FullName string `json:"fullName"`
		} `json:"unit"`
	} `json:"data"`
}

type Product struct {
	ProductUid            string `json:"productUid"`
	ProductName           string `json:"productName"`
	PortSpeed             int    `json:"portSpeed"`
	LocationId            int    `json:"locationId"`
	VxcPermitted          bool   `json:"vxcPermitted"`
	CompanyUid            string `json:"companyUid"`
	CompanyName           string `json:"companyName"`
	ProductType           string `json:"productType"`
	ProvisioningStatus    string `json:"provisioningStatus"`
	CreateDate            int64  `json:"createDate"`
	CreatedBy             string `json:"createdBy"`
	TerminateDate         int64  `json:"terminateDate"`
	LiveDate              int64  `json:"liveDate"`
	Market                string `json:"market"`
	UsageAlgorithm        string `json:"usageAlgorithm"`
	MarketplaceVisibility bool   `json:"marketplaceVisibility"`
	VxcAutoApproval       bool   `json:"vxcAutoApproval"`
	SecondaryName         string `json:"secondaryName"`
	LagPrimary            bool   `json:"lagPrimary"`
	LagId                 int    `json:"lagId"`
	AggregationId         int    `json:"aggregationId"`
	ContractStartDate     int64  `json:"contractStartDate"`
	ContractEndDate       int64  `json:"contractEndDate"`
	ContractTermMonths    int    `json:"contractTermMonths"`
	AssociatedVxcs        []struct {
		ProductId          int    `json:"productId"`
		ProductUid         string `json:"productUid"`
		ProductName        string `json:"productName"`
		ProductType        string `json:"productType"`
		RateLimit          int    `json:"rateLimit"`
		DistanceBand       string `json:"distanceBand"`
		ProvisioningStatus string `json:"provisioningStatus"`
		AEnd               struct {
			OwnerUid      string `json:"ownerUid"`
			ProductUid    string `json:"productUid"`
			ProductName   string `json:"productName"`
			LocationId    int    `json:"locationId"`
			Location      string `json:"location"`
			Vlan          int    `json:"vlan"`
			InnerVlan     int    `json:"innerVlan"`
			SecondaryName int    `json:"secondaryName"`
			VNicIndex     int    `json:"vNicIndex,omitempty"`
		} `json:"aEnd"`
		BEnd struct {
			OwnerUid      string `json:"ownerUid"`
			ProductUid    string `json:"productUid"`
			ProductName   string `json:"productName"`
			LocationId    int    `json:"locationId"`
			Location      string `json:"location"`
			Vlan          int    `json:"vlan"`
			InnerVlan     int    `json:"innerVlan"`
			SecondaryName string `json:"secondaryName"`
		} `json:"bEnd"`
		SecondaryName  string `json:"secondaryName"`
		UsageAlgorithm string `json:"usageAlgorithm"`
		CreatedBy      string `json:"createdBy"`
		CreateDate     int64  `json:"createDate"`
		Resources      struct {
			CspConnection interface{} `json:"csp_connection,omitempty"`
			Vll           struct {
				AVlan         int    `json:"a_vlan,omitempty"`
				BVlan         int    `json:"b_vlan"`
				RateLimitMbps int    `json:"rate_limit_mbps"`
				ResourceName  string `json:"resource_name"`
				ResourceType  string `json:"resource_type"`
				Up            int    `json:"up"`
				Shutdown      bool   `json:"shutdown"`
				AInnerVlan    int    `json:"a_inner_vlan,omitempty"`
				BInnerVlan    int    `json:"b_inner_vlan,omitempty"`
				AinnerVlan    int    `json:"ainnerVlan,omitempty"`
				BinnerVlan    int    `json:"binnerVlan,omitempty"`
			} `json:"vll"`
		} `json:"resources"`
		VxcApproval struct {
			Status   string `json:"status"`
			Message  string `json:"message"`
			Uid      string `json:"uid"`
			Type     string `json:"type"`
			NewSpeed int    `json:"newSpeed"`
		} `json:"vxcApproval"`
		ContractStartDate  int  `json:"contractStartDate"`
		ContractEndDate    int  `json:"contractEndDate"`
		ContractTermMonths int  `json:"contractTermMonths"`
		Locked             bool `json:"locked"`
		AdminLocked        bool `json:"adminLocked"`
		AttributeTags      struct {
		} `json:"attributeTags"`
		NserviceId string `json:"nserviceId"`
		Cancelable bool   `json:"cancelable"`
		CostCentre string `json:"costCentre,omitempty"`
	} `json:"associatedVxcs"`
	AssociatedIxs []interface{} `json:"associatedIxs"`
	Resources     struct {
		Interface struct {
			Demarcation  string `json:"demarcation"`
			LoaTemplate  string `json:"loa_template"`
			Media        string `json:"media"`
			PortSpeed    int    `json:"port_speed"`
			ResourceName string `json:"resource_name"`
			ResourceType string `json:"resource_type"`
			Up           int    `json:"up"`
			Shutdown     bool   `json:"shutdown,omitempty"`
		} `json:"interface"`
		VirtualRouter struct {
			McrAsn             int    `json:"mcrAsn"`
			ResourceName       string `json:"resource_name"`
			ResourceType       string `json:"resource_type"`
			Speed              int    `json:"speed"`
			BgpShutdownDefault bool   `json:"bgpShutdownDefault"`
		} `json:"virtual_router,omitempty"`
		VirtualMachine []struct {
			CpuCount int `json:"cpu_count"`
			Id       int `json:"id"`
			Image    struct {
				Id      int    `json:"id"`
				Vendor  string `json:"vendor"`
				Product string `json:"product"`
				Version string `json:"version"`
			} `json:"image"`
			ResourceType string `json:"resource_type"`
			Up           bool   `json:"up"`
			Vnics        []struct {
				Vlan        int    `json:"vlan"`
				Description string `json:"description,omitempty"`
			} `json:"vnics"`
		} `json:"virtual_machine,omitempty"`
	} `json:"resources"`
	AttributeTags struct {
	} `json:"attributeTags"`
	Virtual       bool   `json:"virtual"`
	BuyoutPort    bool   `json:"buyoutPort"`
	Locked        bool   `json:"locked"`
	AdminLocked   bool   `json:"adminLocked"`
	DiversityZone string `json:"diversityZone"`
	NserviceId    string `json:"nserviceId"`
	Cancelable    bool   `json:"cancelable"`
	CostCentre    string `json:"costCentre,omitempty"`
	Vendor        string `json:"vendor,omitempty"`
	MveSize       string `json:"mveSize,omitempty"`
	MveLabel      string `json:"mveLabel,omitempty"`
}

type ProductList struct {
	Message string    `json:"message"`
	Terms   string    `json:"terms"`
	Data    []Product `json:"data"`
}
