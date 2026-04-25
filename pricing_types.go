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

// PriceBookDiscountElement is a discount applied to the product.
type PriceBookDiscountElement struct {
	DiscountReason   DiscountReason `json:"discountReason"`
	Amount           float64        `json:"amount"`
	Code             string         `json:"code,omitempty"`
	Description      string         `json:"description,omitempty"`
	PercentageAmount float64        `json:"percentageAmount,omitempty"`
}

// PriceBookDto is the pricing response for a single product.
type PriceBookDto struct {
	ProductType     string                      `json:"productType"`
	Currency        string                      `json:"currency"`
	MonthlyRate     float64                     `json:"monthlyRate"`
	MonthlyRackRate float64                     `json:"monthlyRackRate"`
	Prices          []*PriceBookPriceElement    `json:"prices"`
	Discounts       []*PriceBookDiscountElement `json:"discounts"`
}

// PriceBookRequest is implemented by all product-specific pricing request types.
// Pass the concrete type directly to GetProductPricing.
type PriceBookRequest interface {
	pricingProductType() string
}

// VXCPriceBookRequest is a pricing request for a VXC.
type VXCPriceBookRequest struct {
	ProductType     string `json:"productType"`  // always "VXC"
	Currency        string `json:"currency,omitempty"`
	ALocationID     int    `json:"aLocationId"`
	BLocationID     int    `json:"bLocationId"`
	Speed           int    `json:"speed"`
	AEndProductType string `json:"aEndProductType,omitempty"`
	ConnectType     string `json:"connectType,omitempty"`
	Term            int    `json:"term,omitempty"`
	ProductUID      string `json:"productUid,omitempty"`
	BuyoutPort      bool   `json:"buyoutPort,omitempty"`
}

func (r *VXCPriceBookRequest) pricingProductType() string { return "VXC" }

// MCRPriceBookRequest is a pricing request for an MCR.
type MCRPriceBookRequest struct {
	ProductType string `json:"productType"` // always "MCR2"
	Currency    string `json:"currency,omitempty"`
	LocationID  int    `json:"locationId"`
	Speed       int    `json:"speed"`
	Term        int    `json:"term,omitempty"`
	ProductUID  string `json:"productUid,omitempty"`
}

func (r *MCRPriceBookRequest) pricingProductType() string { return "MCR2" }

// MegaportPriceBookRequest is a pricing request for a Port.
type MegaportPriceBookRequest struct {
	ProductType string `json:"productType"` // always "MEGAPORT"
	Currency    string `json:"currency,omitempty"`
	LocationID  int    `json:"locationId"`
	Speed       int    `json:"speed"`
	Term        int    `json:"term,omitempty"`
	ProductUID  string `json:"productUid,omitempty"`
}

func (r *MegaportPriceBookRequest) pricingProductType() string { return "MEGAPORT" }

// MVEPriceBookRequest is a pricing request for an MVE.
type MVEPriceBookRequest struct {
	ProductType string `json:"productType"` // always "MVE"
	Currency    string `json:"currency,omitempty"`
	LocationID  int    `json:"locationId,omitempty"`
	Size        string `json:"size,omitempty"`
	MVELabel    string `json:"mveLabel,omitempty"`
	Term        int    `json:"term,omitempty"`
	ProductUID  string `json:"productUid,omitempty"`
}

func (r *MVEPriceBookRequest) pricingProductType() string { return "MVE" }

// IXPriceBookRequest is a pricing request for an IX.
type IXPriceBookRequest struct {
	ProductType    string `json:"productType"` // always "IX"
	Currency       string `json:"currency,omitempty"`
	PortLocationID int    `json:"portLocationId"`
	IXType         string `json:"ixType"`
	Speed          int    `json:"speed"`
	Term           int    `json:"term,omitempty"`
	ProductUID     string `json:"productUid,omitempty"`
}

func (r *IXPriceBookRequest) pricingProductType() string { return "IX" }

// NATGatewayPriceBookRequest is a pricing request for a NAT Gateway.
type NATGatewayPriceBookRequest struct {
	ProductType  string `json:"productType"` // always "NAT_GATEWAY"
	Currency     string `json:"currency,omitempty"`
	LocationID   int    `json:"locationId"`
	Speed        int    `json:"speed"`
	SessionCount int    `json:"sessionCount"`
	Term         int    `json:"term,omitempty"`
	ProductUID   string `json:"productUid,omitempty"`
}

func (r *NATGatewayPriceBookRequest) pricingProductType() string { return "NAT_GATEWAY" }

// productPricingResponse is the envelope returned by POST /v4/pricebook/product.
type productPricingResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    *PriceBookDto `json:"data"`
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
)
