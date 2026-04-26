package megaport

import "errors"

// PricingFrequency is the billing frequency for a price element.
type PricingFrequency string

const (
	PricingFrequencyOnce    PricingFrequency = "ONCE"
	PricingFrequencyMonthly PricingFrequency = "MONTHLY"
	PricingFrequencyYearly  PricingFrequency = "YEARLY"
)

// PriceBookChargeReason is the reason for a charge element.
type PriceBookChargeReason string

const (
	PriceBookChargeReasonCore          PriceBookChargeReason = "CORE"
	PriceBookChargeReasonCoreSurcharge PriceBookChargeReason = "CORE_SURCHARGE"
	PriceBookChargeReasonAddOn         PriceBookChargeReason = "ADD_ON_CHARGE"
)

// DiscountReason is the reason for a discount element.
type DiscountReason string

const (
	DiscountReasonTerm           DiscountReason = "TERM"
	DiscountReasonWholesale      DiscountReason = "WHOLESALE"
	DiscountReasonPartner        DiscountReason = "PARTNER"
	DiscountReasonPartnerManaged DiscountReason = "PARTNER_MANAGED"
	DiscountReasonReseller       DiscountReason = "RESELLER"
)

// PriceBookPriceElement is a single price component in a pricing response.
type PriceBookPriceElement struct {
	ChargeReason PriceBookChargeReason `json:"chargeReason"`
	Frequency    PricingFrequency      `json:"frequency"`
	Amount       float64               `json:"amount"`
	AddOnType    string                `json:"addOnType,omitempty"` // set when ChargeReason == ADD_ON_CHARGE
}

// DiscountDetails contains additional information about an applied discount.
type DiscountDetails struct {
	UID              string  `json:"uid"`
	Code             string  `json:"code"`
	Description      string  `json:"description"`
	PercentageAmount float64 `json:"percentageAmount,omitempty"`
	Shared           bool    `json:"shared,omitempty"`
	AllocationMethod string  `json:"allocationMethod,omitempty"`
	Mechanism        string  `json:"mechanism,omitempty"`
}

// PriceBookDiscountElement is a discount applied to the product.
type PriceBookDiscountElement struct {
	DiscountReason  DiscountReason   `json:"discountReason"`
	Amount          float64          `json:"amount"`
	DiscountDetails *DiscountDetails `json:"discountDetails,omitempty"`
}

// PriceBookDTO is the pricing response for a single product.
type PriceBookDTO struct {
	ProductType     string                      `json:"productType"`
	Currency        string                      `json:"currency"`
	MonthlyRate     float64                     `json:"monthlyRate"`
	MonthlyRackRate float64                     `json:"monthlyRackRate"`
	Prices          []*PriceBookPriceElement    `json:"prices"`
	Discounts       []*PriceBookDiscountElement `json:"discounts"`
}

// ProductAddOnPriceBookRequest represents a product add-on for pricing requests.
// Set AddOnType to "CROSS_CONNECT" or "IPSEC" and populate the relevant fields.
type ProductAddOnPriceBookRequest struct {
	AddOnType             string `json:"addOnType"`
	TunnelCount           int    `json:"tunnelCount,omitempty"`
	CrossConnectRequested *bool  `json:"crossConnectRequested,omitempty"`
}

// PriceBookRequest is implemented by all product-specific pricing request types.
// Pass the concrete type directly to GetProductPricing.
type PriceBookRequest interface {
	pricingProductType() string
}

// GetProductPricingRequest wraps a PriceBookRequest with optional company context query params.
type GetProductPricingRequest struct {
	// Req is the product-specific pricing request.
	Req PriceBookRequest
	// CompanyID optionally scopes the pricing to a specific company (by numeric ID).
	// Mutually exclusive with CompanyUID.
	CompanyID int
	// CompanyUID optionally scopes the pricing to a specific company (by UUID).
	// Mutually exclusive with CompanyID.
	CompanyUID string
}

// VXCPriceBookRequest is a pricing request for a VXC.
type VXCPriceBookRequest struct {
	Currency        string                          `json:"currency,omitempty"`
	ALocationID     int                             `json:"aLocationId"`
	BLocationID     int                             `json:"bLocationId"`
	Speed           int                             `json:"speed"`
	AEndProductType string                          `json:"aEndProductType,omitempty"`
	ConnectType     string                          `json:"connectType,omitempty"`
	Term            int                             `json:"term,omitempty"`
	ProductUID      string                          `json:"productUid,omitempty"`
	BuyoutPort      bool                            `json:"buyoutPort,omitempty"`
	AddOns          []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *VXCPriceBookRequest) pricingProductType() string { return "VXC" }

// MCRPriceBookRequest is a pricing request for an MCR.
type MCRPriceBookRequest struct {
	Currency   string                          `json:"currency,omitempty"`
	LocationID int                             `json:"locationId"`
	Speed      int                             `json:"speed"`
	Term       int                             `json:"term,omitempty"`
	ProductUID string                          `json:"productUid,omitempty"`
	AddOns     []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *MCRPriceBookRequest) pricingProductType() string { return "MCR2" }

// MegaportPriceBookRequest is a pricing request for a Port.
type MegaportPriceBookRequest struct {
	Currency   string                          `json:"currency,omitempty"`
	LocationID int                             `json:"locationId"`
	Speed      int                             `json:"speed"`
	Term       int                             `json:"term,omitempty"`
	ProductUID string                          `json:"productUid,omitempty"`
	AddOns     []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *MegaportPriceBookRequest) pricingProductType() string { return "MEGAPORT" }

// MVEPriceBookRequest is a pricing request for an MVE.
type MVEPriceBookRequest struct {
	Currency   string                          `json:"currency,omitempty"`
	LocationID int                             `json:"locationId,omitempty"`
	Size       string                          `json:"size,omitempty"`
	MVELabel   string                          `json:"mveLabel,omitempty"`
	Term       int                             `json:"term,omitempty"`
	ProductUID string                          `json:"productUid,omitempty"`
	AddOns     []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *MVEPriceBookRequest) pricingProductType() string { return "MVE" }

// IXPriceBookRequest is a pricing request for an IX.
type IXPriceBookRequest struct {
	Currency       string                          `json:"currency,omitempty"`
	PortLocationID int                             `json:"portLocationId"`
	IXType         string                          `json:"ixType"`
	Speed          int                             `json:"speed"`
	Term           int                             `json:"term,omitempty"`
	ProductUID     string                          `json:"productUid,omitempty"`
	AddOns         []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *IXPriceBookRequest) pricingProductType() string { return "IX" }

// NATGatewayPriceBookRequest is a pricing request for a NAT Gateway.
type NATGatewayPriceBookRequest struct {
	Currency     string                          `json:"currency,omitempty"`
	LocationID   int                             `json:"locationId"`
	Speed        int                             `json:"speed"`
	SessionCount int                             `json:"sessionCount"`
	Term         int                             `json:"term,omitempty"`
	ProductUID   string                          `json:"productUid,omitempty"`
	AddOns       []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *NATGatewayPriceBookRequest) pricingProductType() string { return "NAT_GATEWAY" }

// IPAddressPriceBookRequest is a pricing request for an IP Address block.
type IPAddressPriceBookRequest struct {
	Currency   string                          `json:"currency,omitempty"`
	LocationID int                             `json:"locationId"`
	IPBlock    string                          `json:"ipBlock"` // e.g. "/24"
	ProductUID string                          `json:"productUid,omitempty"`
	AddOns     []*ProductAddOnPriceBookRequest `json:"addOns,omitempty"`
}

func (r *IPAddressPriceBookRequest) pricingProductType() string { return "IP_ADDRESS" }

// productPricingResponse is the envelope returned by POST /v4/pricebook/product.
type productPricingResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    *PriceBookDTO `json:"data"`
}

var (
	ErrPricingRequestNil          = errors.New("pricing request is required")
	ErrPricingVXCLocationRequired = errors.New("VXC pricing requires aLocationId and bLocationId")
	ErrPricingVXCSpeedRequired    = errors.New("VXC pricing requires speed")
	ErrPricingLocationRequired    = errors.New("pricing requires locationId")
	ErrPricingSpeedRequired       = errors.New("pricing requires speed")
	ErrPricingIXTypeRequired      = errors.New("IX pricing requires ixType")
	ErrPricingIXLocationRequired  = errors.New("IX pricing requires portLocationId")
	ErrPricingNATSessionRequired  = errors.New("NAT Gateway pricing requires sessionCount")
	ErrPricingIPBlockRequired     = errors.New("IP Address pricing requires ipBlock")
	ErrPricingIPLocationRequired  = errors.New("IP Address pricing requires locationId")
	ErrPricingCompanyIDAndUIDSet  = errors.New("companyId and companyUid are mutually exclusive; set only one")
	ErrPricingMVELocationRequired = errors.New("MVE pricing requires locationId")
)
