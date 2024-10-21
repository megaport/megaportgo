package megaport

import "strconv"

// MCROrder represents a request to buy an MCR from the Megaport Products API.
type MCROrder struct {
	LocationID int            `json:"locationId"`
	Name       string         `json:"productName"`
	Term       int            `json:"term"`
	Type       string         `json:"productType"`
	PortSpeed  int            `json:"portSpeed"`
	CostCentre string         `json:"costCentre"`
	PromoCode  string         `json:"promoCode,omitempty"`
	Config     MCROrderConfig `json:"config"`

	ResourceTags []ResourceTag `json:"resourceTags,omitempty"`
}

// MCROrderConfig represents the configuration for an MCR order.
type MCROrderConfig struct {
	ASN           int    `json:"mcrAsn,omitempty"`
	DiversityZone string `json:"diversityZone,omitempty"`
}

// MCROrderConfirmation represents a response from the Megaport Products API after ordering an MCR.
type MCROrderConfirmation struct {
	TechnicalServiceUID string `json:"technicalServiceUid"`
}

// MCR represents a Megaport Cloud Router in the Megaport MCR API.
type MCR struct {
	ID                    int                     `json:"productId"`
	UID                   string                  `json:"productUid"`
	Name                  string                  `json:"productName"`
	Type                  string                  `json:"productType"`
	ProvisioningStatus    string                  `json:"provisioningStatus"`
	CreateDate            *Time                   `json:"createDate"`
	CreatedBy             string                  `json:"createdBy"`
	CostCentre            string                  `json:"costCentre"`
	PortSpeed             int                     `json:"portSpeed"`
	TerminateDate         *Time                   `json:"terminateDate"`
	LiveDate              *Time                   `json:"liveDate"`
	Market                string                  `json:"market"`
	LocationID            int                     `json:"locationId"`
	UsageAlgorithm        string                  `json:"usageAlgorithm"`
	MarketplaceVisibility bool                    `json:"marketplaceVisibility"`
	VXCPermitted          bool                    `json:"vxcpermitted"`
	VXCAutoApproval       bool                    `json:"vxcAutoApproval"`
	MaxVXCSpeed           int                     `json:"maxVxcSpeed"`
	SecondaryName         string                  `json:"secondaryName"`
	LAGPrimary            bool                    `json:"lagPrimary"`
	LAGID                 int                     `json:"lagId"`
	AggregationID         int                     `json:"aggregationId"`
	CompanyUID            string                  `json:"companyUid"`
	CompanyName           string                  `json:"companyName"`
	ContractStartDate     *Time                   `json:"contractStartDate"`
	ContractEndDate       *Time                   `json:"contractEndDate"`
	ContractTermMonths    int                     `json:"contractTermMonths"`
	AttributeTags         map[string]string       `json:"attributeTags"`
	Virtual               bool                    `json:"virtual"`
	BuyoutPort            bool                    `json:"buyoutPort"`
	Locked                bool                    `json:"locked"`
	AdminLocked           bool                    `json:"adminLocked"`
	Cancelable            bool                    `json:"cancelable"`
	Resources             MCRResources            `json:"resources"`
	LocationDetails       *ProductLocationDetails `json:"locationDetail"`
}

// MCRResources represents the resources associated with an MCR.
type MCRResources struct {
	Interface     PortInterface    `json:"interface"`
	VirtualRouter MCRVirtualRouter `json:"virtual_router"`
}

// MCRVirtualRouter represents the virtual router associated with an MCR.
type MCRVirtualRouter struct {
	ID           int    `json:"id"`
	ASN          int    `json:"mcrAsn"`
	Name         string `json:"name"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Speed        int    `json:"speed"`
}

// MCRPrefixFilterList represents a prefix filter list associated with an MCR.
type MCRPrefixFilterList struct {
	ID            int                   `json:"id"` // ID of the prefix filter list.
	Description   string                `json:"description"`
	AddressFamily string                `json:"addressFamily"`
	Entries       []*MCRPrefixListEntry `json:"entries"`
}

// APIMCRPrefixFilterListEntry represents an entry in a prefix filter list.
type APIMCRPrefixFilterListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     string `json:"ge,omitempty"` // Greater than or equal to - (Optional) The minimum starting prefix length to be matched. Valid values are from 0 to 32 (IPv4), or 0 to 128 (IPv6). The minimum (ge) must be no greater than or equal to the maximum value (le).
	Le     string `json:"le,omitempty"` // Less than or equal to - (Optional) The maximum ending prefix length to be matched. The prefix length is greater than or equal to the minimum value (ge). Valid values are from 0 to 32 (IPv4), or 0 to 128 (IPv6), but the maximum must be no less than the minimum value (ge).
}

// MCRPrefixListEntry represents an entry in a prefix filter list.
type MCRPrefixListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"` // Great than or equal to - (Optional) The minimum starting prefix length to be matched. Valid values are from 0 to 32 (IPv4), or 0 to 128 (IPv6). The minimum (ge) must be no greater than or equal to the maximum value (le).
	Le     int    `json:"le,omitempty"` // Less than or equal to - (Optional) The maximum ending prefix length to be matched. The prefix length is greater than or equal to the minimum value (ge). Valid values are from 0 to 32 (IPv4), or 0 to 128 (IPv6), but the maximum must be no less than the minimum value (ge).
}

// MCROrdersResponse represents a response from the Megaport Products API after ordering an MCR.
type MCROrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*MCROrderConfirmation `json:"data"`
}

// MCRResponse represents a response from the Megaport MCR API after querying an MCR.
type MCRResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    *MCR   `json:"data"`
}

// PrefixFilterList represents a prefix filter list associated with an MCR.
type PrefixFilterList struct {
	Id            int    `json:"id"`
	Description   string `json:"description"`
	AddressFamily string `json:"addressFamily"`
}

// CreateMCRPrefixFilterListResponse represents a response from the Megaport MCR API after creating a prefix filter list.
type CreateMCRPrefixFilterListAPIResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    *APIMCRPrefixFilterList `json:"data"`
}

// ListMCRPrefixFilterListResponse represents a response from the Megaport MCR API after querying an MCR's prefix filter list.
type ListMCRPrefixFilterListResponse struct {
	Message string              `json:"message"`
	Terms   string              `json:"terms"`
	Data    []*PrefixFilterList `json:"data"`
}

type APIMCRPrefixFilterListResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    *APIMCRPrefixFilterList `json:"data"`
}

type APIMCRPrefixFilterList struct {
	ID            int                            `json:"id"`
	Description   string                         `json:"description"`
	AddressFamily string                         `json:"addressFamily"`
	Entries       []*APIMCRPrefixFilterListEntry `json:"entries"`
}

func (e *APIMCRPrefixFilterList) ToMCRPrefixFilterList() (*MCRPrefixFilterList, error) {
	entries := make([]*MCRPrefixListEntry, len(e.Entries))
	for i, entry := range e.Entries {
		mcrEntry, err := entry.ToMCRPrefixFilterListEntry()
		if err != nil {
			return nil, err
		}
		entries[i] = mcrEntry
	}
	return &MCRPrefixFilterList{
		ID:            e.ID,
		Description:   e.Description,
		AddressFamily: e.AddressFamily,
		Entries:       entries,
	}, nil
}

func (e *APIMCRPrefixFilterListEntry) ToMCRPrefixFilterListEntry() (*MCRPrefixListEntry, error) {
	var ge, le int
	if e.Ge != "" {
		geVal, err := strconv.Atoi(e.Ge)
		if err != nil {
			return nil, err
		}
		ge = geVal
	}
	if e.Le != "" {
		leVal, err := strconv.Atoi(e.Le)
		if err != nil {
			return nil, err
		}
		le = leVal
	}
	return &MCRPrefixListEntry{
		Action: e.Action,
		Prefix: e.Prefix,
		Ge:     ge,
		Le:     le,
	}, nil
}
