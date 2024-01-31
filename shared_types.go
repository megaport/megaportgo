package megaport

const (
	// The CONFIGURED service state.
	SERVICE_CONFIGURED = "CONFIGURED"

	// The LIVE service state.
	SERVICE_LIVE = "LIVE"
)

var (
	SERVICE_STATE_READY = []string{SERVICE_CONFIGURED, SERVICE_LIVE}
)

type GenericResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    map[string]interface{} `json:"data"`
}

const APPLICATION_SHORT_NAME = "mpt"

const PRODUCT_MEGAPORT = "megaport"
const PRODUCT_VXC = "vxc"
const PRODUCT_MCR = "mcr2"
const PRODUCT_MVE = "mve"
const PRODUCT_IX = "ix"

const STATUS_DECOMMISSIONED string = "DECOMMISSIONED"
const STATUS_CANCELLED string = "CANCELLED"
const SINGLE_PORT string = "Single"
const LAG_PORT string = "LAG"
const CONNECT_TYPE_AWS_VIF string = "AWS"
const CONNECT_TYPE_AWS_HOSTED_CONNECTION string = "AWSHC"

const MODIFY_NAME string = "NAME"
const MODIFY_COST_CENTRE = "COST_CENTRE"
const MODIFY_MARKETPLACE_VISIBILITY string = "MARKETPLACE_VISIBILITY"
const MODIFY_RATE_LIMIT = "RATE_LIMIT"
const MODIFY_A_END_VLAN = "A_VLAN"
const MODIFY_B_END_VLAN = "B_VLAN"
