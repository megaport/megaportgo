package megaport

const (
	SERVICE_CONFIGURED = "CONFIGURED" // The CONFIGURED service state.
	SERVICE_LIVE       = "LIVE"       // The LIVE service state.

	PRODUCT_MEGAPORT = "megaport"
	PRODUCT_VXC      = "vxc"
	PRODUCT_MCR      = "mcr2"
	PRODUCT_MVE      = "mve"
	PRODUCT_IX       = "ix"

	STATUS_DECOMMISSIONED = "DECOMMISSIONED"
	STATUS_CANCELLED      = "CANCELLED"

	SINGLE_PORT                        = "Single"
	LAG_PORT                           = "LAG"
	CONNECT_TYPE_AWS_VIF               = "AWS"
	CONNECT_TYPE_AWS_HOSTED_CONNECTION = "AWSHC"
)

var (
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)
